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

// GetHourlyProductionAccurate 获取指定逻辑日每小时精确产量统计（完全参数化，无硬编码）
// configs: 设备变量配置数组，每个设备指定产量ID、NG加减按钮ID和设备名称；nil 使用默认配置。
// logicalDate: "YYYY-MM-DD" 格式的逻辑日期；空字符串使用 CURDATE()（今天）。
//
// Returns hourly accurate production statistics for the given logical date.
// 指定論理日付の時間別精確生産統計を返す。
func GetHourlyProductionAccurate(configs []DeviceVarConfig, logicalDate string) ([]HourlyProductionAccurate, error) {
	// 默认配置：设备1和设备2
	if len(configs) == 0 {
		configs = []DeviceVarConfig{
			{DeviceName: "一号机", ProductionVarID: 1, NgAddVarID: 72, NgSubVarID: 71},
			{DeviceName: "二号机", ProductionVarID: 95, NgAddVarID: 97, NgSubVarID: 96},
		}
	}
	if logicalDate == "" {
		logicalDate = time.Now().Format("2006-01-02")
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

	// CN: 日期范围覆盖逻辑日当天 + 次日（跨零点班次如 14:20→07:40 next day 需要取次日记录）。
	//     time_slot 用 DATE_ADD 基于逻辑日午夜加整小时数来生成，保证夜班后半段的 time_slot
	//     与前端 hourLabelToIdx 的 nextDay 日期推导保持一致。
	// EN: Date range covers logical date + next day (cross-midnight shifts need next-day records).
	//     time_slot is built by adding whole-hour offsets from midnight of logicalDate,
	//     matching the frontend's nextDay date derivation for cross-midnight hour labels.
	// JP: 日付範囲は論理日 + 翌日をカバー（深夜跨ぎシフトには翌日レコードが必要）。
	//     time_slot は論理日深夜からの整時オフセットで生成し、フロントの nextDay 推論と一致させる。
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
			AND created_at >= CAST('%s' AS DATE)
			AND created_at <  DATE_ADD(CAST('%s' AS DATE), INTERVAL 2 DAY)
	),
	ProcessedData AS (
		SELECT 
			-- CN: 用逻辑日午夜到记录时刻的绝对小时数构造 time_slot，跨日时 hour_idx >= 24
			--     如 2026-04-11 01:00 对应 logicalDate=2026-04-10，hour_idx=25，
			--     time_slot = DATE_ADD('2026-04-10', INTERVAL 25 HOUR) = '2026-04-11 01:00:00'
			-- EN: Absolute hour offset from midnight of logicalDate; cross-midnight yields hour_idx>=24.
			-- JP: 論理日深夜からの絶対時間オフセットで time_slot を構成。深夜跨ぎ時は hour_idx>=24。
			DATE_ADD(CAST('%s' AS DATE), INTERVAL FLOOR(TIMESTAMPDIFF(SECOND, CAST('%s' AS DATE), created_at) / 3600) HOUR) AS time_slot,
			%s AS machine_name,
			%s AS production_delta,
			%s AS ng_add,
			%s AS ng_sub
		FROM RawData
	)
	SELECT 
		DATE_FORMAT(time_slot, '%%Y-%%m-%%d %%H:00:00') AS time_slot,
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
	`, varIDsStr, logicalDate, logicalDate, logicalDate, logicalDate, deviceNameSQL, productionSQL, ngAddSQL, ngSubSQL)

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

// GetShiftQualityByRun 从生产运行记录表获取指定班次时间窗口内各设备良品率
// CN: 逻辑与 GetDailyQualityByRun 完全相同，仅将日期过滤替换为精确时间范围，用于当班/历史班次数据展示。
// EN: Same logic as GetDailyQualityByRun, but filters by an exact datetime range for shift-scoped quality display.
// JP: GetDailyQualityByRun と同ロジック。日付フィルタを正確な時間範囲に置き換えてシフト単位の良品率を返す。
func GetShiftQualityByRun(start, end time.Time) ([]DeviceQualityStat, error) {
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
	WHERE r.start_time >= ? AND r.start_time < ?
	GROUP BY d.id, d.device_name
	ORDER BY d.id ASC
	`
	var results []DeviceQualityStat
	err := DB.Raw(query, start, end).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询班次运行良品率失败: %w", err)
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

// ShiftWindow 班次工作时间窗口（用于 OEE SQL 参数化）
// Shift work-time window used to parameterise the OEE SQL query.
// OEE SQL クエリをパラメータ化するためのシフト作業時間ウィンドウ。
//
// WorkStart / WorkEnd 格式为 "HH:MM"（如 "07:40" / "16:20"）。
// LogicalDate 为 "YYYY-MM-DD"，空字符串时 SQL 使用 CURDATE()（即今天）。
// HourStart / HourEnd 控制 Hours CTE 生成的小时范围：
//   - HourStart：第一个时间桶的 hour_idx（如 7 → 7:00-8:00）
//   - HourEnd：SQL 中 WHERE hour_idx < HourEnd 的上限（如 16 → 最后桶 16:00-17:00）
//
// 传 nil 时使用兜底默认值 07:40-16:20 / HourStart=7 / HourEnd=16 / 今天。
type ShiftWindow struct {
	WorkStart   string // "HH:MM"，如 "07:40"
	WorkEnd     string // "HH:MM"，如 "16:20"
	LogicalDate string // "YYYY-MM-DD"，逻辑日期；空 = CURDATE()
	HourStart   int    // Hours CTE 起始 hour_idx，0 = 默认 7
	HourEnd     int    // Hours CTE 上限（WHERE hour_idx < HourEnd），0 = 默认 16
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
// window: 班次时间窗口（工作起止时间），传nil时使用默认 07:40-16:20
//
// Returns hourly OEE statistics for today using dynamic shift time window.
// 動的なシフト時間ウィンドウを使用して本日の時間別OEE統計を返す。
func GetHourlyOEE(configs []DeviceOEEConfig, breakTimes []BreakTimeConfig, window *ShiftWindow) ([]HourlyOEE, error) {
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

	// 解析班次时间窗口；未传入时使用历史默认值 07:40-16:20
	workStart := "07:40"
	workEnd := "16:20"
	logicalDate := time.Now().Format("2006-01-02") // 默认今天
	startHour := 7                                 // Hours CTE 起始 hour_idx
	hourEndLimit := 16                             // Hours CTE WHERE hour_idx < N 的 N

	if window != nil {
		if window.WorkStart != "" {
			workStart = window.WorkStart
		}
		if window.WorkEnd != "" {
			workEnd = window.WorkEnd
		}
		if window.LogicalDate != "" {
			logicalDate = window.LogicalDate
		}
		if window.HourStart > 0 {
			startHour = window.HourStart
		}
		if window.HourEnd > 0 {
			hourEndLimit = window.HourEnd
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

	// CN: Hours CTE 结束小时标签（用于 time_period 输出）
	// EN: End-hour label shown in time_period output (= hourEndLimit, the last bucket's hour_idx)
	// JP: time_period 出力に使う終了時間ラベル（最後のバケットの hour_idx = hourEndLimit）
	query := fmt.Sprintf(`
WITH RECURSIVE 
-- 1. 配置（逻辑日期 + 工作时间窗口，均由班次动态传入，兜底默认今天 07:40-16:20）
Config AS ( 
    SELECT CAST('%s' AS DATE) as target_date,
           ADDTIME(CAST('%s' AS DATE), '%s:00') as work_start,
           ADDTIME(CAST('%s' AS DATE), '%s:00') as work_end
),
Hours AS (
    SELECT %d as hour_idx,
           ADDTIME(target_date, '%02d:00:00') as hour_start,
           ADDTIME(target_date, '%02d:00:00') as hour_end
    FROM Config
    UNION ALL
    SELECT hour_idx + 1,
           ADDTIME(target_date, SEC_TO_TIME((hour_idx + 1) * 3600)),
           ADDTIME(target_date, SEC_TO_TIME((hour_idx + 2) * 3600))
    FROM Hours, Config WHERE hour_idx < %d
),
DeviceConfig AS (
    %s
),

-- 2. 定义休息时间段（动态生成，含班次内休息 + 班次间间隔）
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

-- 4. 生产流 (查班次数据，额外拉一条 work_start 之前的最后记录用于 LAG 基准)
-- CN: work_start/work_end 均为完整 DATETIME，跨日班次（如22:00→06:00）work_end 会指向次日，ADDTIME 直接使用无需截断。
-- EN: work_start/work_end are full DATETIMEs; for cross-midnight shifts, work_end points to next day, ADDTIME handles it correctly.
-- JP: work_start/work_end は完全な DATETIME。深夜跨ぎシフトでは work_end が翌日を指し、ADDTIME がそのまま正しく処理する。
-- CN: ProductionRaw 加入 last_nonzero_val 防回跳子查询，消除计数器 reset bounce（如 65→0→65 产生的虚增产量）。
-- EN: ProductionRaw includes last_nonzero_val subquery to prevent counter-reset bounce (e.g. 65→0→65 ghost delta).
-- JP: ProductionRaw に last_nonzero_val サブクエリを追加し、カウンタリセットバウンス（例: 65→0→65）による虚偽増分を防止。
ProductionRaw AS (
    SELECT 
        d.created_at, d.val, d.var_id, dc.device_id, dc.var_ok, dc.var_ng_add, dc.var_ng_sub,
        CASE WHEN d.var_id IN (dc.var_ok) THEN 
            LAG(d.val) OVER (PARTITION BY d.var_id ORDER BY d.created_at) 
        END as prev_val,
        CASE WHEN d.var_id IN (dc.var_ok) THEN (
            SELECT h2.val FROM sys_data_history h2
            WHERE h2.var_id = d.var_id AND h2.created_at < d.created_at AND h2.val > 0
            ORDER BY h2.created_at DESC LIMIT 1
        ) END as last_nonzero_val
    FROM sys_data_history d
    JOIN DeviceConfig dc ON d.var_id IN (dc.var_ok, dc.var_ng_add, dc.var_ng_sub)
    CROSS JOIN Config c
    WHERE (
        -- 班次窗口前后各 1 小时，确保 LAG 有足够数据；直接用 DATETIME 加减，兼容跨日
        d.created_at >= ADDTIME((SELECT work_start FROM Config), '-01:00:00')
        AND d.created_at <= ADDTIME((SELECT work_end FROM Config), '01:00:00')
    ) OR (
        -- 每个 var_id 在 work_start 之前的最后一条，用于 LAG 基准
        d.id IN (
            SELECT MAX(id) FROM sys_data_history
            WHERE var_id IN (%s)
              AND created_at < (SELECT work_start FROM Config)
            GROUP BY var_id
        )
    )
),

-- CN: 三档防回跳 CASE：①prev=0且有last_nonzero→用last_nonzero ②正常递增→差值 ③其余→取当前值
-- EN: Three-branch anti-bounce CASE: ①prev=0 with last_nonzero→use it ②normal increment→delta ③else→current val
-- JP: 3段防バウンスCASE：①prev=0+last_nonzero有り→使用 ②正常増加→差分 ③その他→現在値
ProductionStats AS (
    SELECT 
        FLOOR(TIMESTAMPDIFF(SECOND, c.target_date, created_at) / 3600) as hour_idx,
        device_id,
        SUM(CASE 
            WHEN var_id = var_ok THEN 
                CASE 
                    WHEN prev_val IS NULL THEN val
                    WHEN prev_val = 0 AND last_nonzero_val IS NOT NULL AND val >= last_nonzero_val
                        THEN val - last_nonzero_val
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
    WHERE created_at >= (SELECT work_start FROM Config)
      AND created_at <= (SELECT work_end FROM Config)
    GROUP BY FLOOR(TIMESTAMPDIFF(SECOND, c.target_date, created_at) / 3600), device_id
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
	`, logicalDate, logicalDate, workStart, logicalDate, workEnd,
		startHour, startHour, startHour+1, hourEndLimit,
		deviceConfigSQL, breaksSQL, varOKList)

	// log.Printf("[GetHourlyOEE] 执行OEE查询，逻辑日期=%s, 时间窗=%s~%s, 小时范围=%d~%d", logicalDate, workStart, workEnd, startHour, hourEndLimit)
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
// fullDay=true 时 t_plan 使用完整时间窗口（不受 NOW() 截断），适用于全天计划视图；
// fullDay=false（默认）时 t_plan 截断到当前时刻，适用于实时驾驶舱模式。
// fullDay=true: t_plan uses full window (no NOW() cap), for full-day plan view.
// fullDay=true: t_plan はウィンドウ全体を使用（NOW() でカットしない）、全日計画ビュー用。
func GetHourlyOEEWithSQL(configs []DeviceOEEConfig, breakTimes []BreakTimeConfig, window *ShiftWindow, fullDay ...bool) ([]HourlyOEEDebug, string, error) {
	useFullDay := len(fullDay) > 0 && fullDay[0]
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

	workStart := "07:40"
	workEnd := "16:20"
	logicalDate2 := time.Now().Format("2006-01-02")
	startHour2 := 7
	hourEndLimit2 := 16

	if window != nil {
		if window.WorkStart != "" {
			workStart = window.WorkStart
		}
		if window.WorkEnd != "" {
			workEnd = window.WorkEnd
		}
		if window.LogicalDate != "" {
			logicalDate2 = window.LogicalDate
		}
		if window.HourStart > 0 {
			startHour2 = window.HourStart
		}
		if window.HourEnd > 0 {
			hourEndLimit2 = window.HourEnd
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

	// CN: fullDay=true 时 t_plan 使用完整班次窗口（不受 NOW() 截断），适用于全天计划视图。
	// EN: fullDay=true uses the complete shift window for t_plan (no NOW() cap), for full-day plan view.
	// JP: fullDay=true は t_plan にシフト全体を使用（NOW() カットなし）、全日計画ビュー用。
	var tPlanSQL string
	if useFullDay {
		tPlanSQL = `GREATEST(
            TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, (SELECT work_start FROM Config)), LEAST(hd.hour_end, (SELECT work_end FROM Config)))
            - bo.break_sec,
        0) as t_plan`
	} else {
		tPlanSQL = `GREATEST(
            (CASE WHEN hd.hour_start > NOW() THEN 0
                  WHEN hd.hour_end > NOW() THEN TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, (SELECT work_start FROM Config)), LEAST(NOW(), (SELECT work_end FROM Config)))
                  ELSE TIMESTAMPDIFF(SECOND, GREATEST(hd.hour_start, (SELECT work_start FROM Config)), LEAST(hd.hour_end, (SELECT work_end FROM Config)))
            END) - bo.break_sec,
        0) as t_plan`
	}

	query := fmt.Sprintf(`WITH RECURSIVE 
Config AS (
    SELECT CAST('%s' AS DATE) as target_date,
           ADDTIME(CAST('%s' AS DATE), '%s:00') as work_start,
           ADDTIME(CAST('%s' AS DATE), '%s:00') as work_end
),
Hours AS (
    SELECT %d as hour_idx,
           ADDTIME(target_date, '%02d:00:00') as hour_start,
           ADDTIME(target_date, '%02d:00:00') as hour_end
    FROM Config
    UNION ALL
    SELECT hour_idx + 1,
           ADDTIME(target_date, SEC_TO_TIME((hour_idx + 1) * 3600)),
           ADDTIME(target_date, SEC_TO_TIME((hour_idx + 2) * 3600))
    FROM Hours, Config WHERE hour_idx < %d
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
-- CN: ProductionRaw 加入 last_nonzero_val 防回跳（与 GetHourlyOEE 同步修复）。
-- EN: ProductionRaw includes last_nonzero_val (synced fix with GetHourlyOEE).
-- JP: ProductionRaw に last_nonzero_val を追加（GetHourlyOEE と同期修正）。
ProductionRaw AS (
    SELECT d.created_at, d.val, d.var_id, dc.device_id, dc.var_ok, dc.var_ng_add, dc.var_ng_sub,
        CASE WHEN d.var_id IN (dc.var_ok) THEN
            LAG(d.val) OVER (PARTITION BY d.var_id ORDER BY d.created_at)
        END as prev_val,
        CASE WHEN d.var_id IN (dc.var_ok) THEN (
            SELECT h2.val FROM sys_data_history h2
            WHERE h2.var_id = d.var_id AND h2.created_at < d.created_at AND h2.val > 0
            ORDER BY h2.created_at DESC LIMIT 1
        ) END as last_nonzero_val
    FROM sys_data_history d
    JOIN DeviceConfig dc ON d.var_id IN (dc.var_ok, dc.var_ng_add, dc.var_ng_sub)
    CROSS JOIN Config c
    WHERE (
        -- 班次窗口前后各1小时，用完整 DATETIME 加减，兼容跨日班次
        d.created_at >= ADDTIME((SELECT work_start FROM Config), '-01:00:00')
        AND d.created_at <= ADDTIME((SELECT work_end FROM Config), '01:00:00')
    ) OR (
        d.id IN (
            SELECT MAX(id) FROM sys_data_history
            WHERE var_id IN (%s)
              AND created_at < (SELECT work_start FROM Config)
            GROUP BY var_id
        )
    )
),
ProductionStats AS (
    SELECT FLOOR(TIMESTAMPDIFF(SECOND, c.target_date, created_at) / 3600) as hour_idx, device_id,
        SUM(CASE WHEN var_id = var_ok THEN
            CASE
                WHEN prev_val IS NULL THEN val
                WHEN prev_val = 0 AND last_nonzero_val IS NOT NULL AND val >= last_nonzero_val
                    THEN val - last_nonzero_val
                WHEN val >= prev_val THEN val - prev_val
                ELSE val
            END
        ELSE 0 END) as total_qty,
        GREATEST(SUM(CASE
            WHEN var_id = var_ng_add AND val = 1 THEN 1
            WHEN var_id = var_ng_sub AND val = 1 THEN -1
            ELSE 0 END), 0) as ng_qty
    FROM ProductionRaw
    CROSS JOIN Config c
    WHERE created_at >= (SELECT work_start FROM Config)
      AND created_at <= (SELECT work_end FROM Config)
    GROUP BY FLOOR(TIMESTAMPDIFF(SECOND, c.target_date, created_at) / 3600), device_id
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
        %s
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
ORDER BY device_id, hour_idx`,
		logicalDate2, logicalDate2, workStart, logicalDate2, workEnd,
		startHour2, startHour2, startHour2+1, hourEndLimit2,
		deviceConfigSQL, breaksSQL, varOKList, tPlanSQL)

	var results []HourlyOEEDebug
	err := DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, query, fmt.Errorf("GetHourlyOEEWithSQL查询失败: %w", err)
	}
	return results, query, nil
}

// ========================================================
// 班次窗口产量聚合（防复位回跳版，供快照生成复用）
// Shift-window production aggregation with anti-reset-bounce logic (shared by snapshot generation).
// 班次ウィンドウ産量集計（リセット跳ね返り防止版、スナップショット生成共用）。
// ========================================================

// ShiftWindowProdResult 单设备班次窗口产量聚合结果。
// Result of shift-window production aggregation for one device.
// 1設備分の班次ウィンドウ産量集計結果。
type ShiftWindowProdResult struct {
	TotalQty int `json:"total_qty"` // 窗口内总产量（防复位口径）
	NgQty    int `json:"ng_qty"`    // NG 净增量（ng_add 脉冲 - ng_sub 脉冲）
	OkQty    int `json:"ok_qty"`    // 良品数量（TotalQty - NgQty）
}

// GetShiftWindowProduction 聚合指定时间窗口内单个设备的产量、NG、良品数。
// CN: 产量计数器采用与 GetHourlyProductionAccurate/GetMonthlyProductionAccurate 完全一致的
//
//	"防复位回跳"规则：
//	  1. prev_val = 0 且 last_nonzero_val IS NOT NULL 且 val >= last_nonzero_val
//	     → delta = val - last_nonzero_val（设备复位后首次上报，相对于复位前最后有效值）
//	  2. val >= prev_val → delta = val - prev_val（正常累加）
//	  3. 其余（如清零后 val < last_nonzero_val）→ delta = val（当次绝对值当增量，即设备被设置为新基点）
//	NG 加/减按钮保持现状：val=1 脉冲计数，不做边沿去重。
//
// EN: Production counter uses the same anti-reset-bounce rule as GetHourlyProductionAccurate.
//
//	NG add/sub buttons remain as pulse-counts (no edge-dedup this round).
//
// JP: 産量カウンタは GetHourlyProductionAccurate と同一のリセット跳ね返り防止規則を適用。
//
//	NG ボタンは今回未変更（パルスカウント方式継続）。
//
// cfg: 设备变量映射（生产变量 ID / NG 加减变量 ID）。
// shiftStart / shiftEnd: 班次时间窗口（含首端，不含末端，与 SQL BETWEEN 语义一致）。
func GetShiftWindowProduction(cfg DeviceVarConfig, shiftStart, shiftEnd time.Time) (ShiftWindowProdResult, error) {
	startStr := shiftStart.Format("2006-01-02 15:04:05")
	endStr := shiftEnd.Format("2006-01-02 15:04:05")

	// CN: productionSQL 与月度/小时精确统计完全相同的三档 CASE：
	//     ① 复位后首条（prev=0, last_nonzero_val 存在，新值>=旧非零值）→ val - last_nonzero_val
	//     ② 正常累加（val >= prev_val）→ val - prev_val
	//     ③ 其余情况（新值<旧非零值，即清零重置）→ val
	// EN: Three-branch CASE identical to monthly/hourly accurate stats.
	// JP: 月次・時間別精確統計と完全同一の3分岐 CASE 文。
	productionSQL := fmt.Sprintf(
		"WHEN var_id = %d THEN CASE "+
			"WHEN prev_val = 0 AND last_nonzero_val IS NOT NULL AND val >= last_nonzero_val THEN val - last_nonzero_val "+
			"WHEN val >= prev_val THEN val - prev_val "+
			"ELSE val END ",
		cfg.ProductionVarID)

	ngAddSQL := fmt.Sprintf("WHEN var_id = %d AND val = 1 THEN 1 ", cfg.NgAddVarID)
	ngSubSQL := fmt.Sprintf("WHEN var_id = %d AND val = 1 THEN 1 ", cfg.NgSubVarID)

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
			WHERE var_id IN (%d, %d, %d)
				AND created_at >= '%s'
				AND created_at <  '%s'
		),
		Calc AS (
			SELECT
				CASE %s ELSE 0 END AS prod,
				CASE %s ELSE 0 END AS ng_add,
				CASE %s ELSE 0 END AS ng_sub
			FROM RawData
		)
		SELECT
			COALESCE(SUM(prod), 0)                                       AS total_qty,
			GREATEST(0, COALESCE(SUM(ng_add) - SUM(ng_sub), 0))         AS ng_qty,
			COALESCE(SUM(prod), 0) - GREATEST(0, COALESCE(SUM(ng_add) - SUM(ng_sub), 0)) AS ok_qty
		FROM Calc`,
		cfg.ProductionVarID, cfg.NgAddVarID, cfg.NgSubVarID,
		startStr, endStr,
		productionSQL, ngAddSQL, ngSubSQL,
	)

	var result ShiftWindowProdResult
	if err := DB.Raw(query).Scan(&result).Error; err != nil {
		return ShiftWindowProdResult{}, fmt.Errorf("GetShiftWindowProduction 查询失败: %w", err)
	}
	return result, nil
}
