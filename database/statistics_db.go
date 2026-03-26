package database

import (
	"fmt"
	"log"

	// "log"
	"time"
)

// HourlyProduction 小时产量统计
type HourlyProduction struct {
	Hour     int `json:"hour"`      // 小时 (0-23)
	DeviceID int `json:"device_id"` // 设备ID
	OkQty    int `json:"ok_qty"`    // 良品数量
	NgQty    int `json:"ng_qty"`    // 不良品数量
	TotalQty int `json:"total_qty"` // 总数量
}

// GetHourlyProduction 获取今日按小时统计的产量（按设备）
func GetHourlyProduction(deviceID *int) ([]HourlyProduction, error) {
	// 获取今日起始时间
	todayStart := time.Now().Truncate(24 * time.Hour)

	query := `
		SELECT 
			HOUR(start_time) as hour,
			device_id,
			SUM(run_ok_qty) as ok_qty,
			SUM(run_ng_qty) as ng_qty,
			SUM(run_ok_qty + run_ng_qty) as total_qty
		FROM pro_production_runs
		WHERE start_time >= ?
		`

	args := []interface{}{todayStart}

	if deviceID != nil {
		query += " AND device_id = ?"
		args = append(args, *deviceID)
	}

	query += `
		GROUP BY HOUR(start_time), device_id
		ORDER BY device_id, hour
	`

	var results []HourlyProduction
	err := DB.Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询小时产量失败: %w", err)
	}

	return results, nil
}

// StaffEfficiency 员工绩效统计
type StaffEfficiency struct {
	StaffID     int     `json:"staff_id"`     // 员工ID
	StaffName   string  `json:"staff_name"`   // 员工姓名
	TotalOkQty  int     `json:"total_ok_qty"` // 总良品数
	TotalNgQty  int     `json:"total_ng_qty"` // 总不良品数
	TotalQty    int     `json:"total_qty"`    // 总产量
	QualityRate float64 `json:"quality_rate"` // 良品率 (%)
	WorkingMin  int     `json:"working_min"`  // 工作时长(分钟)
	Efficiency  float64 `json:"efficiency"`   // 效率 = 工作时长/总时长 (%)
}

// GetStaffEfficiency 获取员工绩效统计（今日或指定时间范围）
// 绩效 = 员工工作时长 / 从早上7点到现在的总时长 × 100%
func GetStaffEfficiency(startTime, endTime *time.Time) ([]StaffEfficiency, error) {
	// 默认统计今日从早上7点开始
	if startTime == nil {
		today := time.Now().Truncate(24 * time.Hour)
		morning7 := today.Add(7 * time.Hour) // 早上7点
		startTime = &morning7
	}
	if endTime == nil {
		now := time.Now()
		endTime = &now
	}

	// 计算总时长（分钟）
	totalMinutes := int(endTime.Sub(*startTime).Minutes())
	if totalMinutes <= 0 {
		totalMinutes = 1 // 避免除零
	}

	query := `
		SELECT 
			s.id as staff_id,
			s.name as staff_name,
			COALESCE(SUM(r.run_ok_qty), 0) as total_ok_qty,
			COALESCE(SUM(r.run_ng_qty), 0) as total_ng_qty,
			COALESCE(SUM(r.run_ok_qty + r.run_ng_qty), 0) as total_qty,
			CASE 
				WHEN SUM(r.run_ok_qty + r.run_ng_qty) > 0 
				THEN (SUM(r.run_ok_qty) * 100.0 / SUM(r.run_ok_qty + r.run_ng_qty))
				ELSE 100.0
			END as quality_rate,
			COALESCE(SUM(
				TIMESTAMPDIFF(MINUTE, 
					CASE WHEN sess.login_time < ? THEN ? ELSE sess.login_time END,
					CASE 
						WHEN sess.logout_time IS NULL THEN ?
						WHEN sess.logout_time > ? THEN ?
						ELSE sess.logout_time
					END
				)
			), 0) as working_min,
			CASE 
				WHEN ? > 0
				THEN (COALESCE(SUM(
					TIMESTAMPDIFF(MINUTE, 
						CASE WHEN sess.login_time < ? THEN ? ELSE sess.login_time END,
						CASE 
							WHEN sess.logout_time IS NULL THEN ?
							WHEN sess.logout_time > ? THEN ?
							ELSE sess.logout_time
						END
					)
				), 0) * 100.0 / ?)
				ELSE 0
			END as efficiency
		FROM sys_staff s
		LEFT JOIN pro_machine_sessions sess ON JSON_CONTAINS(sess.staff_ids, CAST(s.id AS CHAR))
			AND sess.login_time < ?
			AND (sess.logout_time IS NULL OR sess.logout_time > ?)
		LEFT JOIN pro_production_runs r ON JSON_CONTAINS(r.operator_ids, CAST(s.id AS CHAR))
			AND r.start_time >= ? AND r.start_time <= ?
		WHERE s.is_active = 1
		GROUP BY s.id, s.name
		ORDER BY efficiency DESC
	`

	var results []StaffEfficiency
	// 参数顺序：working_min计算中的6个，efficiency计算中的6个，最后的session过滤2个，run过滤2个
	err := DB.Raw(query,
		// working_min 计算参数
		startTime, startTime, endTime, endTime, endTime,
		// efficiency 计算参数
		totalMinutes, startTime, startTime, endTime, endTime, endTime, totalMinutes,
		// session 过滤参数
		endTime, startTime,
		// run 过滤参数
		startTime, endTime,
	).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询员工绩效失败: %w", err)
	}

	return results, nil
}

// HourlyProductionPulse 每小时产量脉冲统计（基于历史数据表）
type HourlyProductionPulse struct {
	DeviceID   int    `json:"device_id"`   // 设备ID
	DeviceName string `json:"device_name"` // 设备名称
	Hour       int    `json:"hour"`        // 小时 (0-23)
	OkQty      int    `json:"ok_qty"`      // 良品脉冲次数
	NgQty      int    `json:"ng_qty"`      // 不良品脉冲次数
	TotalQty   int    `json:"total_qty"`   // 总脉冲次数
}

// GetHourlyProductionPulse 获取今日每小时产量脉冲统计（从历史数据表）
func GetHourlyProductionPulse(deviceID *int) ([]HourlyProductionPulse, error) {
	query := `
	SELECT 
		v.device_id,
		d.device_name,
		HOUR(h.created_at) as hour,
		SUM(CASE WHEN v.var_name LIKE '%产量加1' THEN 1 ELSE 0 END) as ok_qty,
		SUM(CASE WHEN v.var_name LIKE '%NG加1' THEN 1 ELSE 0 END) as ng_qty,
		COUNT(*) as total_qty
	FROM sys_data_history h
	INNER JOIN sys_variables v ON h.var_id = v.id
	INNER JOIN sys_devices d ON v.device_id = d.id
	WHERE 
		h.created_at >= CURDATE()
		AND h.created_at < CURDATE() + INTERVAL 1 DAY
		AND h.val = 1
		AND h.var_id IN (
			SELECT id FROM sys_variables 
			WHERE var_name LIKE '%产量加1' OR var_name LIKE '%NG加1'
		)
	`

	args := []interface{}{}

	if deviceID != nil {
		query += " AND v.device_id = ?"
		args = append(args, *deviceID)
	}

	query += `
	GROUP BY v.device_id, d.device_name, HOUR(h.created_at)
	ORDER BY v.device_id, hour
	`

	var results []HourlyProductionPulse
	err := DB.Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询每小时产量脉冲失败: %w", err)
	}

	return results, nil
}

// DeviceVarConfig 设备变量配置（用于参数化查询）
type DeviceVarConfig struct {
	DeviceName      string `json:"device_name"`       // 设备名称
	ProductionVarID int    `json:"production_var_id"` // 产量累加计数器变量ID
	NgAddVarID      int    `json:"ng_add_var_id"`     // NG增加按钮变量ID
	NgSubVarID      int    `json:"ng_sub_var_id"`     // NG减少按钮变量ID
}

// HourlyProductionAccurate 每小时精确产量统计（支持累加计数器+NG按钮）
type HourlyProductionAccurate struct {
	TimeSlot    string  `json:"time_slot"`    // 时间段 (YYYY-MM-DD HH:00:00)
	DeviceName  string  `json:"device_name"`  // 设备名称
	TotalQty    int     `json:"total_qty"`    // 总产量
	NgQty       int     `json:"ng_qty"`       // NG数量（净增量）
	OkQty       int     `json:"ok_qty"`       // 良品数量（总产量-NG）
	QualityRate float64 `json:"quality_rate"` // 良品率（%）
}

// GetHourlyProductionAccurate 获取今日每小时精确产量统计（完全参数化，无硬编码）
// configs: 设备变量配置数组，每个设备指定产量ID、NG加减按钮ID和设备名称
// 如果传入nil，使用默认配置（设备1和设备2）
func GetHourlyProductionAccurate(configs []DeviceVarConfig) ([]HourlyProductionAccurate, error) {
	// 默认配置：设备1和设备2
	if len(configs) == 0 {
		configs = []DeviceVarConfig{
			{DeviceName: "一号机", ProductionVarID: 1, NgAddVarID: 72, NgSubVarID: 71},
			{DeviceName: "二号机", ProductionVarID: 95, NgAddVarID: 97, NgSubVarID: 96},
		}
	}

	// 收集所有变量ID并构建动态SQL
	var allVarIDs []int
	deviceNameCases := ""
	productionCases := ""
	ngAddCases := ""
	ngSubCases := ""

	for _, cfg := range configs {
		allVarIDs = append(allVarIDs, cfg.ProductionVarID, cfg.NgAddVarID, cfg.NgSubVarID)

		// 设备名称映射
		deviceNameCases += fmt.Sprintf("WHEN var_id IN (%d, %d, %d) THEN '%s' ",
			cfg.ProductionVarID, cfg.NgAddVarID, cfg.NgSubVarID, cfg.DeviceName)

		// 产量变量（累加计数器）- 每个设备的完整WHEN...THEN语句
		productionCases += fmt.Sprintf(
			"WHEN var_id = %d THEN CASE "+
				"WHEN prev_val = 0 AND last_nonzero_val IS NOT NULL AND val >= last_nonzero_val THEN val - last_nonzero_val "+
				"WHEN val >= prev_val THEN val - prev_val "+
				"ELSE val END ",
			cfg.ProductionVarID)

		// NG增加按钮
		ngAddCases += fmt.Sprintf("WHEN var_id = %d AND val = 1 THEN 1 ", cfg.NgAddVarID)

		// NG减少按钮
		ngSubCases += fmt.Sprintf("WHEN var_id = %d AND val = 1 THEN 1 ", cfg.NgSubVarID)
	}

	// 构建 IN 子句
	varIDsStr := ""
	for i, id := range allVarIDs {
		if i > 0 {
			varIDsStr += ", "
		}
		varIDsStr += fmt.Sprintf("%d", id)
	}

	// 完整的CASE语句
	deviceNameSQL := fmt.Sprintf("CASE %s ELSE 'Unknown' END", deviceNameCases)
	productionSQL := fmt.Sprintf("CASE %s ELSE 0 END", productionCases)
	ngAddSQL := fmt.Sprintf("CASE %s ELSE 0 END", ngAddCases)
	ngSubSQL := fmt.Sprintf("CASE %s ELSE 0 END", ngSubCases)

	// 调试输出
	// log.Printf("[GetHourlyProductionAccurate] 生成的productionSQL: %s", productionSQL)

	query := fmt.Sprintf(`
	WITH RawData AS (
		SELECT 
			var_id,
			val,
			created_at,
			LAG(val, 1, 0) OVER (PARTITION BY var_id ORDER BY created_at) AS prev_val,
			(
				SELECT h2.val
				FROM sys_data_history h2
				WHERE h2.var_id = h1.var_id
					AND h2.created_at < h1.created_at
					AND h2.val > 0
				ORDER BY h2.created_at DESC
				LIMIT 1
			) AS last_nonzero_val
		FROM sys_data_history h1
		WHERE 
			var_id IN (%s)
			AND DATE(created_at) = CURDATE()
	),
	ProcessedData AS (
		SELECT 
			DATE_FORMAT(created_at, '%%Y-%%m-%%d %%H:00:00') AS time_slot,
			%s AS machine_name,
			%s AS production_delta,
			%s AS ng_add,
			%s AS ng_sub
		FROM RawData
	)
	SELECT 
		time_slot,
		machine_name AS device_name,
		SUM(production_delta) AS total_qty,
		GREATEST(0, SUM(ng_add) - SUM(ng_sub)) AS ng_qty,
		SUM(production_delta) - GREATEST(0, SUM(ng_add) - SUM(ng_sub)) AS ok_qty,
		CASE 
			WHEN SUM(production_delta) = 0 THEN 100.0
			ELSE ROUND(
				(SUM(production_delta) - GREATEST(0, SUM(ng_add) - SUM(ng_sub))) * 100.0 / 
				NULLIF(SUM(production_delta), 0), 
			2)
		END AS quality_rate
	FROM ProcessedData
	WHERE machine_name != 'Unknown'
	GROUP BY time_slot, machine_name
	ORDER BY time_slot ASC, machine_name ASC
	`, varIDsStr, deviceNameSQL, productionSQL, ngAddSQL, ngSubSQL)

	var results []HourlyProductionAccurate
	err := DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询每小时精确产量失败: %w", err)
	}

	return results, nil
}

// MonthlyProductionAccurate 月度产量汇总（按设备）
type MonthlyProductionAccurate struct {
	DeviceName  string  `json:"device_name"`  // 设备名称
	TotalQty    int     `json:"total_qty"`    // 总产量
	NgQty       int     `json:"ng_qty"`       // NG数量
	OkQty       int     `json:"ok_qty"`       // 良品数量
	QualityRate float64 `json:"quality_rate"` // 良品率（%）
}

// GetMonthlyProductionAccurate 获取当月产量汇总统计（按设备，不走工单表）
func GetMonthlyProductionAccurate(configs []DeviceVarConfig) ([]MonthlyProductionAccurate, error) {
	if len(configs) == 0 {
		configs = []DeviceVarConfig{
			{DeviceName: "一号机", ProductionVarID: 1, NgAddVarID: 72, NgSubVarID: 71},
			{DeviceName: "二号机", ProductionVarID: 95, NgAddVarID: 97, NgSubVarID: 96},
		}
	}

	var allVarIDs []int
	deviceNameCases := ""
	productionCases := ""
	ngAddCases := ""
	ngSubCases := ""

	for _, cfg := range configs {
		allVarIDs = append(allVarIDs, cfg.ProductionVarID, cfg.NgAddVarID, cfg.NgSubVarID)

		deviceNameCases += fmt.Sprintf("WHEN var_id IN (%d, %d, %d) THEN '%s' ",
			cfg.ProductionVarID, cfg.NgAddVarID, cfg.NgSubVarID, cfg.DeviceName)

		productionCases += fmt.Sprintf(
			"WHEN var_id = %d THEN CASE "+
				"WHEN prev_val = 0 AND last_nonzero_val IS NOT NULL AND val >= last_nonzero_val THEN val - last_nonzero_val "+
				"WHEN val >= prev_val THEN val - prev_val "+
				"ELSE val END ",
			cfg.ProductionVarID)

		ngAddCases += fmt.Sprintf("WHEN var_id = %d AND val = 1 THEN 1 ", cfg.NgAddVarID)
		ngSubCases += fmt.Sprintf("WHEN var_id = %d AND val = 1 THEN 1 ", cfg.NgSubVarID)
	}

	varIDsStr := ""
	for i, id := range allVarIDs {
		if i > 0 {
			varIDsStr += ", "
		}
		varIDsStr += fmt.Sprintf("%d", id)
	}

	deviceNameSQL := fmt.Sprintf("CASE %s ELSE 'Unknown' END", deviceNameCases)
	productionSQL := fmt.Sprintf("CASE %s ELSE 0 END", productionCases)
	ngAddSQL := fmt.Sprintf("CASE %s ELSE 0 END", ngAddCases)
	ngSubSQL := fmt.Sprintf("CASE %s ELSE 0 END", ngSubCases)

	query := fmt.Sprintf(`
	WITH RawData AS (
		SELECT 
			var_id,
			val,
			created_at,
			LAG(val, 1, 0) OVER (PARTITION BY var_id ORDER BY created_at) AS prev_val,
			(
				SELECT h2.val
				FROM sys_data_history h2
				WHERE h2.var_id = h1.var_id
					AND h2.created_at < h1.created_at
					AND h2.val > 0
				ORDER BY h2.created_at DESC
				LIMIT 1
			) AS last_nonzero_val
		FROM sys_data_history h1
		WHERE 
			var_id IN (%s)
			AND DATE(created_at) >= DATE_FORMAT(NOW(), '%%Y-%%m-01')
	),
	ProcessedData AS (
		SELECT 
			%s AS machine_name,
			%s AS production_delta,
			%s AS ng_add,
			%s AS ng_sub
		FROM RawData
	)
	SELECT 
		machine_name AS device_name,
		SUM(production_delta) AS total_qty,
		GREATEST(0, SUM(ng_add) - SUM(ng_sub)) AS ng_qty,
		SUM(production_delta) - GREATEST(0, SUM(ng_add) - SUM(ng_sub)) AS ok_qty,
		CASE 
			WHEN SUM(production_delta) = 0 THEN 100.0
			ELSE ROUND(
				(SUM(production_delta) - GREATEST(0, SUM(ng_add) - SUM(ng_sub))) * 100.0 /
				NULLIF(SUM(production_delta), 0),
			2)
		END AS quality_rate
	FROM ProcessedData
	WHERE machine_name != 'Unknown'
	GROUP BY machine_name
	ORDER BY machine_name ASC
	`, varIDsStr, deviceNameSQL, productionSQL, ngAddSQL, ngSubSQL)

	var results []MonthlyProductionAccurate
	err := DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询月度产量失败: %w", err)
	}

	return results, nil
}

// DeviceQualityStat 设备质量统计（通用）
type DeviceQualityStat struct {
	DeviceID    int     `json:"device_id"`    // 设备ID
	DeviceName  string  `json:"device_name"`  // 设备名称
	TotalQty    int     `json:"total_qty"`    // 总产量
	OkQty       int     `json:"ok_qty"`       // 良品数
	NgQty       int     `json:"ng_qty"`       // 不良品数
	QualityRate float64 `json:"quality_rate"` // 良品率（%）
}

// GetMonthlyQualityByOrder 从工单表获取当月各设备良品率汇总
// 判断依据：pro_production_runs 中本月有班次记录的工单（含跨月工单）
func GetMonthlyQualityByOrder() ([]DeviceQualityStat, error) {
	query := `
	SELECT
		d.id            AS device_id,
		d.device_name   AS device_name,
		SUM(o.ok_qty + o.ng_qty) AS total_qty,
		SUM(o.ok_qty)   AS ok_qty,
		SUM(o.ng_qty)   AS ng_qty,
		CASE
			WHEN SUM(o.ok_qty + o.ng_qty) = 0 THEN 100.0
			ELSE ROUND(SUM(o.ok_qty) * 100.0 / NULLIF(SUM(o.ok_qty + o.ng_qty), 0), 2)
		END AS quality_rate
	FROM pro_orders o
	JOIN sys_devices d ON d.id = o.target_device_id
	WHERE o.id IN (
		SELECT DISTINCT order_id
		FROM pro_production_runs
		WHERE start_time >= DATE_FORMAT(NOW(), '%Y-%m-01')
	)
	GROUP BY d.id, d.device_name
	ORDER BY d.id ASC
	`
	var results []DeviceQualityStat
	err := DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询月度工单良品率失败: %w", err)
	}
	return results, nil
}

// DailyQualityTrend 每日良品率趋势点
type DailyQualityTrend struct {
	Day         string  `json:"day"`          // 日期 YYYY-MM-DD
	OkQty       int     `json:"ok_qty"`       // 当日良品数
	NgQty       int     `json:"ng_qty"`       // 当日不良品数
	QualityRate float64 `json:"quality_rate"` // 当日良品率（%）
}

// GetMonthlyDailyQualityTrend 获取本月每日良品率趋势（走pro_production_runs，含跨月工单）
func GetMonthlyDailyQualityTrend() ([]DailyQualityTrend, error) {
	query := `
	SELECT
		DATE(r.start_time) AS day,
		SUM(r.run_ok_qty)  AS ok_qty,
		SUM(r.run_ng_qty)  AS ng_qty,
		CASE
			WHEN SUM(r.run_ok_qty + r.run_ng_qty) = 0 THEN 100.0
			ELSE ROUND(SUM(r.run_ok_qty) * 100.0 / NULLIF(SUM(r.run_ok_qty + r.run_ng_qty), 0), 2)
		END AS quality_rate
	FROM pro_production_runs r
	WHERE r.start_time >= DATE_FORMAT(NOW(), '%Y-%m-01')
	GROUP BY DATE(r.start_time)
	ORDER BY day ASC
	`
	var results []DailyQualityTrend
	err := DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询每日良品率趋势失败: %w", err)
	}
	return results, nil
}

// GetDailyQualityByRun 从生产运行记录表获取今日各设备良品率
func GetDailyQualityByRun() ([]DeviceQualityStat, error) {
	query := `
	SELECT
		d.id            AS device_id,
		d.device_name   AS device_name,
		SUM(r.run_ok_qty + r.run_ng_qty) AS total_qty,
		SUM(r.run_ok_qty) AS ok_qty,
		SUM(r.run_ng_qty) AS ng_qty,
		CASE
			WHEN SUM(r.run_ok_qty + r.run_ng_qty) = 0 THEN 100.0
			ELSE ROUND(SUM(r.run_ok_qty) * 100.0 / NULLIF(SUM(r.run_ok_qty + r.run_ng_qty), 0), 2)
		END AS quality_rate
	FROM pro_production_runs r
	JOIN sys_devices d ON d.id = r.device_id
	WHERE DATE(r.start_time) = CURDATE()
	GROUP BY d.id, d.device_name
	ORDER BY d.id ASC
	`
	var results []DeviceQualityStat
	err := DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询今日运行良品率失败: %w", err)
	}
	return results, nil
}

// GetActiveOrderQuality 获取当前在产工单（生产中+暂停）各设备良品率
// 不限日期，只要工单状态是 1(生产中) 或 2(暂停) 就纳入统计
func GetActiveOrderQuality() ([]DeviceQualityStat, error) {
	query := `
	SELECT
		d.id            AS device_id,
		d.device_name   AS device_name,
		SUM(o.ok_qty + o.ng_qty) AS total_qty,
		SUM(o.ok_qty)   AS ok_qty,
		SUM(o.ng_qty)   AS ng_qty,
		CASE
			WHEN SUM(o.ok_qty + o.ng_qty) = 0 THEN 100.0
			ELSE ROUND(SUM(o.ok_qty) * 100.0 / NULLIF(SUM(o.ok_qty + o.ng_qty), 0), 2)
		END AS quality_rate
	FROM pro_orders o
	JOIN sys_devices d ON d.id = o.target_device_id
	WHERE o.status IN (1, 2)
	GROUP BY d.id, d.device_name
	ORDER BY d.id ASC
	`
	var results []DeviceQualityStat
	err := DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询在产工单良品率失败: %w", err)
	}
	return results, nil
}

// DeviceUtilizationTrend 设备利用率趋势
type DeviceUtilizationTrend struct {
	Hour        int     `json:"hour"`        // 小时 (0-23)
	DeviceID    int     `json:"device_id"`   // 设备ID
	Utilization float64 `json:"utilization"` // 利用率 (%)
}

// GetDeviceUtilizationTrend 获取设备利用率趋势（今日按小时）
func GetDeviceUtilizationTrend(deviceID *int) ([]DeviceUtilizationTrend, error) {
	// 获取今日起始时间
	todayStart := time.Now().Truncate(24 * time.Hour)

	query := `
		SELECT 
			HOUR(start_time) as hour,
			device_id,
			(SUM(CASE WHEN status = 1 THEN duration_min ELSE 0 END) * 100.0 / 
			 NULLIF(SUM(duration_min), 0)) as utilization
		FROM (
			SELECT 
				device_id,
				status,
				start_time,
				TIMESTAMPDIFF(MINUTE, start_time, COALESCE(end_time, NOW())) as duration_min
			FROM sys_device_status
			WHERE start_time >= ?
		) as hourly_status
		`

	args := []interface{}{todayStart}

	if deviceID != nil {
		query += " WHERE device_id = ?"
		args = append(args, *deviceID)
	}

	query += `
		GROUP BY HOUR(start_time), device_id
		ORDER BY device_id, hour
	`

	var results []DeviceUtilizationTrend
	err := DB.Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询设备利用率趋势失败: %w", err)
	}

	return results, nil
}

// ========================================================
// OEE 计算
// ========================================================

// DeviceOEEConfig 设备OEE配置
type DeviceOEEConfig struct {
	DeviceID   int     `json:"device_id"`   // 设备ID
	DeviceName string  `json:"device_name"` // 设备名称
	VarOK      int     `json:"var_ok"`      // 良品计数器变量ID
	VarNGAdd   int     `json:"var_ng_add"`  // NG加1按钮变量ID
	VarNGSub   int     `json:"var_ng_sub"`  // NG减1按钮变量ID
	CycleTime  float64 `json:"cycle_time"`  // 理论节拍（秒/件）
}

// BreakTimeConfig 休息时间配置（用于OEE计算）
type BreakTimeConfig struct {
	Name      string `json:"name"`       // 名称（如"上午茶歇"）
	StartHour int    `json:"start_hour"` // 开始小时
	StartMin  int    `json:"start_min"`  // 开始分钟
	EndHour   int    `json:"end_hour"`   // 结束小时
	EndMin    int    `json:"end_min"`    // 结束分钟
}

// HourlyOEE 每小时OEE统计
type HourlyOEE struct {
	TimePeriod    string  `json:"time_period"`      // 时间段（如 "7:00 - 8:00" 或 "=== 全天合计 ==="）
	DeviceName    string  `json:"device_name"`      // 设备名称
	TotalRunSec   int     `json:"total_run_sec"`    // 总运行时间（秒）
	TotalPlanSec  int     `json:"total_plan_sec"`   // 总计划时间（秒）
	TotalProducts int     `json:"total_products"`   // 总产量
	Availability  float64 `json:"availability_pct"` // 时间稼动率 (%)
	Performance   float64 `json:"performance_pct"`  // 性能稼动率 (%)
	Quality       float64 `json:"quality_pct"`      // 良品率 (%)
	OEE           float64 `json:"oee_pct"`          // OEE = A × P × Q (%)

	// 兼容旧版前端字段（用于解析hour）
	Hour int `json:"hour"` // 小时 (7-19)，从time_period解析
}


// GetHourlyOEE 获取今日每小时OEE统计
// configs: 设备配置数组，如果为nil则使用默认配置
// breakTimes: 休息时间配置数组，如果为nil则使用默认配置
func GetHourlyOEE(configs []DeviceOEEConfig, breakTimes []BreakTimeConfig) ([]HourlyOEE, error) {
	// 默认配置：设备1和设备2
	if len(configs) == 0 {
		configs = []DeviceOEEConfig{
			{DeviceID: 1, DeviceName: "设备#1", VarOK: 1, VarNGAdd: 72, VarNGSub: 71, CycleTime: 100},
			{DeviceID: 2, DeviceName: "设备#2", VarOK: 95, VarNGAdd: 97, VarNGSub: 96, CycleTime: 100},
		}
	}

	// 默认休息时间配置
	if len(breakTimes) == 0 {
		breakTimes = []BreakTimeConfig{
			{Name: "上午休息", StartHour: 9, StartMin: 40, EndHour: 9, EndMin: 50},
			{Name: "午餐休息", StartHour: 11, StartMin: 40, EndHour: 12, EndMin: 20},
			{Name: "下午休息", StartHour: 14, StartMin: 20, EndHour: 14, EndMin: 30},
		}
	}

	// 构建设备配置SQL
	deviceConfigSQL := ""
	varOKList := ""
	for i, cfg := range configs {
		if i > 0 {
			deviceConfigSQL += " UNION ALL "
			varOKList += ","
		}
		deviceConfigSQL += fmt.Sprintf(
			"SELECT %d as device_id, '%s' as device_name, %d as var_ok, %d as var_ng_add, %d as var_ng_sub, %.2f as cycle_time",
			cfg.DeviceID, cfg.DeviceName, cfg.VarOK, cfg.VarNGAdd, cfg.VarNGSub, cfg.CycleTime,
		)
		varOKList += fmt.Sprintf("%d", cfg.VarOK)
	}

	// ✅ 构建休息时间配置SQL（动态生成）
	breaksSQL := ""
	for i, bt := range breakTimes {
		if i > 0 {
			breaksSQL += " UNION ALL\n    "
		}
		breaksSQL += fmt.Sprintf(
			"SELECT ADDTIME(target_date, '%02d:%02d:00') as b_start, ADDTIME(target_date, '%02d:%02d:00') as b_end FROM Config",
			bt.StartHour, bt.StartMin, bt.EndHour, bt.EndMin,
		)
	}

	// 如果没有休息时间，创建一个空的休息时间段（避免SQL错误）
	if breaksSQL == "" {
		breaksSQL = "SELECT ADDTIME(target_date, '00:00:00') as b_start, ADDTIME(target_date, '00:00:00') as b_end FROM Config WHERE 1=0"
	}

	// log.Printf("[GetHourlyOEE] 使用 %d 个休息时间段配置", len(breakTimes))
	// for _, bt := range breakTimes {
	// 	log.Printf("  - %s: %02d:%02d - %02d:%02d", bt.Name, bt.StartHour, bt.StartMin, bt.EndHour, bt.EndMin)
	// }

	query := fmt.Sprintf(`
WITH RECURSIVE 
-- 1. 配置（工作时间 7:40-16:20）
Config AS ( 
    SELECT CURDATE() as target_date,
           ADDTIME(CURDATE(), '07:40:00') as work_start,
           ADDTIME(CURDATE(), '16:20:00') as work_end
),
Hours AS (
    SELECT 7 as hour_idx, ADDTIME(target_date, '07:00:00') as hour_start, ADDTIME(target_date, '08:00:00') as hour_end FROM Config
    UNION ALL
    SELECT hour_idx + 1, ADDTIME(target_date, SEC_TO_TIME((hour_idx + 1) * 3600)), ADDTIME(target_date, SEC_TO_TIME((hour_idx + 2) * 3600))
    FROM Hours, Config WHERE hour_idx < 16
),
DeviceConfig AS (
    %s
),

-- 2. 定义休息时间段（✅ 动态生成）
Breaks AS (
    %s
),

-- 3. 状态流（统计原始运行秒数，休息时间扣除在 CombinedMetrics 层处理）
StatusStats AS (
    SELECT 
        dh.hour_idx, dh.device_id,
        SUM(CASE WHEN s.status = 1 THEN 
            TIMESTAMPDIFF(SECOND, GREATEST(s.start_time, dh.hour_start), LEAST(COALESCE(s.end_time, NOW()), dh.hour_end))
        ELSE 0 END) as running_sec
    FROM (SELECT h.hour_idx, h.hour_start, h.hour_end, d.device_id FROM Hours h CROSS JOIN DeviceConfig d) dh
    CROSS JOIN Config c
    LEFT JOIN sys_device_status s ON dh.device_id = s.device_id AND s.start_time < dh.hour_end AND COALESCE(s.end_time, NOW()) > dh.hour_start
    AND COALESCE(s.end_time, NOW()) >= (SELECT work_start FROM Config)
    GROUP BY dh.hour_idx, dh.device_id
),

-- 4. 生产流 (查当天数据，额外拉一条昨天最后的记录用于LAG基准)
ProductionRaw AS (
    SELECT 
        d.created_at, d.val, d.var_id, dc.device_id, dc.var_ok, dc.var_ng_add, dc.var_ng_sub,
        CASE WHEN d.var_id IN (dc.var_ok) THEN 
            LAG(d.val) OVER (PARTITION BY d.var_id ORDER BY d.created_at) 
        END as prev_val
    FROM sys_data_history d
    JOIN DeviceConfig dc ON d.var_id IN (dc.var_ok, dc.var_ng_add, dc.var_ng_sub)
    CROSS JOIN Config c
    WHERE (
        -- 今天的数据
        d.created_at >= ADDTIME(c.target_date, '07:00:00')
        AND d.created_at <= ADDTIME(c.target_date, '17:00:00')
    ) OR (
        -- 每个var_id在今天07:00之前的最后一条，用于LAG基准
        d.id IN (
            SELECT MAX(id) FROM sys_data_history
            WHERE var_id IN (%s)
              AND created_at < ADDTIME(CURDATE(), '07:00:00')
            GROUP BY var_id
        )
    )
),

-- var_ok 记录的是总产出累计值（含NG），ng_add/ng_sub 均为脉冲信号（val=1为上升沿）
-- ng_add val=1 → +1NG；ng_sub val=1 → -1NG（撤销）；total_qty = var_ok 差值
-- 注意：ng_qty 不在此处截断，保留负值，由上层 GREATEST 保护，确保合计行撤销能正确抵消
ProductionStats AS (
    SELECT 
        HOUR(created_at) as hour_idx,
        device_id,
        SUM(CASE 
            WHEN var_id = var_ok THEN 
                CASE 
                    WHEN prev_val IS NULL THEN val
                    WHEN val >= prev_val THEN val - prev_val 
                    ELSE val
                END
            ELSE 0 
         END) as total_qty,
         SUM(CASE 
            WHEN var_id = var_ng_add AND val = 1 THEN 1
            WHEN var_id = var_ng_sub AND val = 1 THEN -1
            ELSE 0 
         END) as ng_qty
    FROM ProductionRaw
    CROSS JOIN Config c
    WHERE created_at >= ADDTIME(c.target_date, '07:40:00')
      AND created_at <= ADDTIME(c.target_date, '16:20:00')
    GROUP BY HOUR(created_at), device_id
),

-- 5. 汇总（t_run 和 t_plan 均扣除休息时间，保证口径一致，时间稼动率不超100%）
HourDevice AS (
    SELECT h.hour_idx, h.hour_start, h.hour_end, dc.device_id, dc.device_name, dc.cycle_time
    FROM Hours h CROSS JOIN DeviceConfig dc
),
BreakOverlap AS (
    SELECT hd.hour_idx, hd.device_id,
        COALESCE(SUM(GREATEST(TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, b.b_start), LEAST(hd.hour_end, b.b_end)), 0)), 0) as break_sec
    FROM HourDevice hd
    LEFT JOIN Breaks b ON b.b_start < hd.hour_end AND b.b_end > hd.hour_start
    GROUP BY hd.hour_idx, hd.device_id
),
CombinedMetrics AS (
    SELECT 
        hd.hour_idx, hd.device_id, hd.device_name, hd.cycle_time,
        hd.hour_start, hd.hour_end,
        GREATEST(COALESCE(s.running_sec, 0) - bo.break_sec, 0) as t_run,
        COALESCE(p.total_qty, 0) - COALESCE(p.ng_qty, 0) as q_ok,
        COALESCE(p.ng_qty, 0) as q_ng,
        COALESCE(p.total_qty, 0) as q_total,
        GREATEST((CASE 
            WHEN hd.hour_start > NOW() THEN 0
            WHEN hd.hour_end > NOW() THEN 
                TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, (SELECT work_start FROM Config)), LEAST(NOW(), (SELECT work_end FROM Config)))
            ELSE 
                TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, (SELECT work_start FROM Config)), LEAST(hd.hour_end, (SELECT work_end FROM Config)))
         END) - bo.break_sec, 0) as t_plan
    FROM HourDevice hd
    LEFT JOIN BreakOverlap bo ON hd.hour_idx = bo.hour_idx AND hd.device_id = bo.device_id
    LEFT JOIN StatusStats s ON hd.hour_idx = s.hour_idx AND hd.device_id = s.device_id
    LEFT JOIN ProductionStats p ON hd.hour_idx = p.hour_idx AND hd.device_id = p.device_id
)

-- 6. 输出（q_ok/q_ng 在此处用 GREATEST 保护，保证显示不为负；合计行 SUM 能正确累加含负值的小时）
SELECT 
    CASE WHEN hour_idx IS NULL THEN '=== 全天合计 ===' ELSE CONCAT(hour_idx, ':00 - ', hour_idx+1, ':00') END as time_period,
    MAX(device_name) as device_name,
    SUM(t_run) as total_run_sec, 
    SUM(t_plan) as total_plan_sec, 
    SUM(q_total) as total_products,
    ROUND(CASE WHEN SUM(t_plan) = 0 THEN 0 ELSE SUM(t_run) * 100.0 / SUM(t_plan) END, 2) as availability_pct,
    ROUND(CASE WHEN SUM(t_run) = 0 THEN 0 ELSE (SUM(q_total) * MAX(cycle_time) * 100.0) / SUM(t_run) END, 2) as performance_pct,
    ROUND(CASE WHEN SUM(q_total) = 0 THEN 100 ELSE GREATEST(SUM(q_ok), 0) * 100.0 / SUM(q_total) END, 2) as quality_pct,
    ROUND((CASE WHEN SUM(t_plan) = 0 THEN 0 ELSE SUM(t_run) * 1.0 / SUM(t_plan) END) * (CASE WHEN SUM(t_run) = 0 THEN 0 ELSE (SUM(q_total) * MAX(cycle_time) * 1.0) / SUM(t_run) END) * (CASE WHEN SUM(q_total) = 0 THEN 1 ELSE GREATEST(SUM(q_ok), 0) * 1.0 / SUM(q_total) END) * 100, 2) as oee_pct
FROM CombinedMetrics
GROUP BY device_id, hour_idx WITH ROLLUP
HAVING device_id IS NOT NULL
ORDER BY device_id, hour_idx
	`, deviceConfigSQL, breaksSQL, varOKList)

	// log.Printf("[GetHourlyOEE] 执行OEE查询（包含汇总），设备数: %d, 休息时间段数: %d", len(configs), len(breakTimes))
	log.Printf("[GetHourlyOEE] SQL: %s", query)
	type RawResult struct {
		TimePeriod    string  `gorm:"column:time_period"`
		DeviceName    string  `gorm:"column:device_name"`
		TotalRunSec   int     `gorm:"column:total_run_sec"`
		TotalPlanSec  int     `gorm:"column:total_plan_sec"`
		TotalProducts int     `gorm:"column:total_products"`
		Availability  float64 `gorm:"column:availability_pct"`
		Performance   float64 `gorm:"column:performance_pct"`
		Quality       float64 `gorm:"column:quality_pct"`
		OEE           float64 `gorm:"column:oee_pct"`
	}

	var rawResults []RawResult
	err := DB.Raw(query).Scan(&rawResults).Error
	if err != nil {
		return nil, fmt.Errorf("查询每小时OEE失败: %w", err)
	}

	// 转换为HourlyOEE结构，并解析hour字段
	results := make([]HourlyOEE, 0, len(rawResults))
	for _, raw := range rawResults {
		result := HourlyOEE{
			TimePeriod:    raw.TimePeriod,
			DeviceName:    raw.DeviceName,
			TotalRunSec:   raw.TotalRunSec,
			TotalPlanSec:  raw.TotalPlanSec,
			TotalProducts: raw.TotalProducts,
			Availability:  raw.Availability,
			Performance:   raw.Performance,
			Quality:       raw.Quality,
			OEE:           raw.OEE,
		}

		// 从time_period解析hour（如果不是汇总行）
		if raw.TimePeriod != "=== 全天合计 ===" {
			var hour int
			fmt.Sscanf(raw.TimePeriod, "%d:", &hour)
			result.Hour = hour
		}

		results = append(results, result)
	}

	// log.Printf("[GetHourlyOEE] 查询完成，返回 %d 条记录（包含汇总）", len(results))
	return results, nil
}

// HourlyOEEDebug 带ok_qty/ng_qty的调试结构
type HourlyOEEDebug struct {
	TimePeriod    string  `json:"time_period" gorm:"column:time_period"`
	DeviceName    string  `json:"device_name" gorm:"column:device_name"`
	TotalRunSec   int     `json:"total_run_sec" gorm:"column:total_run_sec"`
	TotalPlanSec  int     `json:"total_plan_sec" gorm:"column:total_plan_sec"`
	TotalProducts int     `json:"total_products" gorm:"column:total_products"`
	OKQty         int     `json:"ok_qty" gorm:"column:ok_qty"`
	NGQty         int     `json:"ng_qty" gorm:"column:ng_qty"`
	Availability  float64 `json:"availability_pct" gorm:"column:availability_pct"`
	Performance   float64 `json:"performance_pct" gorm:"column:performance_pct"`
	Quality       float64 `json:"quality_pct" gorm:"column:quality_pct"`
	OEE           float64 `json:"oee_pct" gorm:"column:oee_pct"`
}

// GetHourlyOEEWithSQL 与GetHourlyOEE使用相同配置，额外返回SQL字符串和ok_qty/ng_qty
func GetHourlyOEEWithSQL(configs []DeviceOEEConfig, breakTimes []BreakTimeConfig) ([]HourlyOEEDebug, string, error) {
	if len(configs) == 0 {
		configs = []DeviceOEEConfig{
			{DeviceID: 1, DeviceName: "设备#1", VarOK: 1, VarNGAdd: 72, VarNGSub: 71, CycleTime: 100},
			{DeviceID: 2, DeviceName: "设备#2", VarOK: 95, VarNGAdd: 97, VarNGSub: 96, CycleTime: 100},
		}
	}
	if len(breakTimes) == 0 {
		breakTimes = []BreakTimeConfig{
			{Name: "上午休息", StartHour: 9, StartMin: 40, EndHour: 9, EndMin: 50},
			{Name: "午餐休息", StartHour: 11, StartMin: 40, EndHour: 12, EndMin: 20},
			{Name: "下午休息", StartHour: 14, StartMin: 20, EndHour: 14, EndMin: 30},
		}
	}

	deviceConfigSQL := ""
	varOKList := ""
	for i, cfg := range configs {
		if i > 0 {
			deviceConfigSQL += " UNION ALL "
			varOKList += ","
		}
		deviceConfigSQL += fmt.Sprintf(
			"SELECT %d as device_id, '%s' as device_name, %d as var_ok, %d as var_ng_add, %d as var_ng_sub, %.2f as cycle_time",
			cfg.DeviceID, cfg.DeviceName, cfg.VarOK, cfg.VarNGAdd, cfg.VarNGSub, cfg.CycleTime,
		)
		varOKList += fmt.Sprintf("%d", cfg.VarOK)
	}

	breaksSQL := ""
	for i, bt := range breakTimes {
		if i > 0 {
			breaksSQL += " UNION ALL\n    "
		}
		breaksSQL += fmt.Sprintf(
			"SELECT ADDTIME(target_date, '%02d:%02d:00') as b_start, ADDTIME(target_date, '%02d:%02d:00') as b_end FROM Config",
			bt.StartHour, bt.StartMin, bt.EndHour, bt.EndMin,
		)
	}
	if breaksSQL == "" {
		breaksSQL = "SELECT ADDTIME(target_date, '00:00:00') as b_start, ADDTIME(target_date, '00:00:00') as b_end FROM Config WHERE 1=0"
	}

	query := fmt.Sprintf(`WITH RECURSIVE 
Config AS (
    SELECT CURDATE() as target_date,
           ADDTIME(CURDATE(), '07:40:00') as work_start,
           ADDTIME(CURDATE(), '16:20:00') as work_end
),
Hours AS (
    SELECT 7 as hour_idx, ADDTIME(target_date, '07:00:00') as hour_start, ADDTIME(target_date, '08:00:00') as hour_end FROM Config
    UNION ALL
    SELECT hour_idx + 1, ADDTIME(target_date, SEC_TO_TIME((hour_idx + 1) * 3600)), ADDTIME(target_date, SEC_TO_TIME((hour_idx + 2) * 3600))
    FROM Hours, Config WHERE hour_idx < 16
),
DeviceConfig AS (
    %s
),
Breaks AS (
    %s
),
StatusStats AS (
    SELECT dh.hour_idx, dh.device_id,
        SUM(CASE WHEN s.status = 1 THEN
            TIMESTAMPDIFF(SECOND, GREATEST(s.start_time, dh.hour_start), LEAST(COALESCE(s.end_time, NOW()), dh.hour_end))
        ELSE 0 END) as running_sec
    FROM (SELECT h.hour_idx, h.hour_start, h.hour_end, d.device_id FROM Hours h CROSS JOIN DeviceConfig d) dh
    CROSS JOIN Config c
    LEFT JOIN sys_device_status s ON dh.device_id = s.device_id AND s.start_time < dh.hour_end AND COALESCE(s.end_time, NOW()) > dh.hour_start
    AND COALESCE(s.end_time, NOW()) >= (SELECT work_start FROM Config)
    GROUP BY dh.hour_idx, dh.device_id
),
ProductionRaw AS (
    SELECT d.created_at, d.val, d.var_id, dc.device_id, dc.var_ok, dc.var_ng_add, dc.var_ng_sub,
        CASE WHEN d.var_id IN (dc.var_ok) THEN
            LAG(d.val) OVER (PARTITION BY d.var_id ORDER BY d.created_at)
        END as prev_val
    FROM sys_data_history d
    JOIN DeviceConfig dc ON d.var_id IN (dc.var_ok, dc.var_ng_add, dc.var_ng_sub)
    CROSS JOIN Config c
    WHERE (
        d.created_at >= ADDTIME(c.target_date, '07:00:00')
        AND d.created_at <= ADDTIME(c.target_date, '17:00:00')
    ) OR (
        d.id IN (
            SELECT MAX(id) FROM sys_data_history
            WHERE var_id IN (%s)
              AND created_at < ADDTIME(CURDATE(), '07:00:00')
            GROUP BY var_id
        )
    )
),
ProductionStats AS (
    SELECT HOUR(created_at) as hour_idx, device_id,
        SUM(CASE WHEN var_id = var_ok THEN
            CASE WHEN prev_val IS NULL THEN val WHEN val >= prev_val THEN val - prev_val ELSE val END
        ELSE 0 END) as total_qty,
        GREATEST(SUM(CASE
            WHEN var_id = var_ng_add AND val = 1 THEN 1
            WHEN var_id = var_ng_sub AND val = 1 THEN -1
            ELSE 0 END), 0) as ng_qty
    FROM ProductionRaw
    CROSS JOIN Config c
    WHERE created_at >= ADDTIME(c.target_date, '07:40:00')
      AND created_at <= ADDTIME(c.target_date, '16:20:00')
    GROUP BY HOUR(created_at), device_id
),
HourDevice AS (
    SELECT h.hour_idx, h.hour_start, h.hour_end, dc.device_id, dc.device_name, dc.cycle_time
    FROM Hours h CROSS JOIN DeviceConfig dc
),
BreakOverlap AS (
    SELECT hd.hour_idx, hd.device_id,
        COALESCE(SUM(GREATEST(TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, b.b_start), LEAST(hd.hour_end, b.b_end)), 0)), 0) as break_sec
    FROM HourDevice hd
    LEFT JOIN Breaks b ON b.b_start < hd.hour_end AND b.b_end > hd.hour_start
    GROUP BY hd.hour_idx, hd.device_id
),
CombinedMetrics AS (
    SELECT 
        hd.hour_idx, hd.device_id, hd.device_name, hd.cycle_time,
        hd.hour_start, hd.hour_end,
        GREATEST(COALESCE(s.running_sec, 0) - bo.break_sec, 0) as t_run,
        COALESCE(p.total_qty, 0) - COALESCE(p.ng_qty, 0) as q_ok,
        COALESCE(p.ng_qty, 0) as q_ng,
        COALESCE(p.total_qty, 0) as q_total,
        GREATEST(
            (CASE WHEN hd.hour_start > NOW() THEN 0
                  WHEN hd.hour_end > NOW() THEN TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, (SELECT work_start FROM Config)), LEAST(NOW(), (SELECT work_end FROM Config)))
                  ELSE TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, (SELECT work_start FROM Config)), LEAST(hd.hour_end, (SELECT work_end FROM Config)))
            END) - bo.break_sec,
        0) as t_plan
    FROM HourDevice hd
    LEFT JOIN BreakOverlap bo ON hd.hour_idx = bo.hour_idx AND hd.device_id = bo.device_id
    LEFT JOIN StatusStats s ON hd.hour_idx = s.hour_idx AND hd.device_id = s.device_id
    LEFT JOIN ProductionStats p ON hd.hour_idx = p.hour_idx AND hd.device_id = p.device_id
)
SELECT CASE WHEN hour_idx IS NULL THEN '合计' ELSE CONCAT(hour_idx, ':00-', hour_idx+1, ':00') END as time_period,
    MAX(device_name) as device_name,
    SUM(t_run) as total_run_sec,
    SUM(t_plan) as total_plan_sec,
    SUM(q_total) as total_products,
    GREATEST(SUM(q_ok), 0) as ok_qty,
    GREATEST(SUM(q_ng), 0) as ng_qty,
    ROUND(CASE WHEN SUM(t_plan) = 0 THEN 0 ELSE SUM(t_run) * 100.0 / SUM(t_plan) END, 2) as availability_pct,
    ROUND(CASE WHEN SUM(t_run) = 0 THEN 0 ELSE (SUM(q_total) * MAX(cycle_time) * 100.0) / SUM(t_run) END, 2) as performance_pct,
    ROUND(CASE WHEN SUM(q_total) = 0 THEN 100 ELSE GREATEST(SUM(q_ok), 0) * 100.0 / SUM(q_total) END, 2) as quality_pct,
    ROUND((CASE WHEN SUM(t_plan) = 0 THEN 0 ELSE SUM(t_run) * 1.0 / SUM(t_plan) END)
        * (CASE WHEN SUM(t_run) = 0 THEN 0 ELSE SUM(q_total) * MAX(cycle_time) * 1.0 / SUM(t_run) END)
        * (CASE WHEN SUM(q_total) = 0 THEN 1 ELSE GREATEST(SUM(q_ok), 0) * 1.0 / SUM(q_total) END) * 100, 2) as oee_pct
FROM CombinedMetrics
GROUP BY device_id, hour_idx WITH ROLLUP
HAVING device_id IS NOT NULL
ORDER BY device_id, hour_idx`, deviceConfigSQL, breaksSQL, varOKList)

	var results []HourlyOEEDebug
	err := DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, query, fmt.Errorf("GetHourlyOEEWithSQL查询失败: %w", err)
	}
	return results, query, nil
}
