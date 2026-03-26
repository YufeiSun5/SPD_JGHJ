// ============================================================================
// 设备状态管理 (Device Status Management) - 设备运行状态跟踪
// ============================================================================
// 职责: 管理设备的运行状态记录（运行/停机/故障）
//
// 核心功能:
//  1. UpdateDeviceStatus: 更新设备状态（自动结束上一状态，创建新状态）
//  2. EndDeviceStatus: 结束当前状态并创建停机状态 ✅ 互斥管理
//  3. GetDeviceCurrentStatus: 获取设备当前状态
//  4. GetDeviceStatusHistory: 获取设备状态历史
//  5. GetAllDevicesStatusSummary: 获取所有设备状态统计
//
// 状态类型 (Status):
//
//	0 - 停机/空闲: 设备未运行
//	1 - 运行: 设备正在生产
//	2 - 故障: 设备故障停机
//
// 互斥状态管理 ⭐ 重要:
//
//	UpdateDeviceStatus: 创建新状态前，自动结束当前活动状态
//	EndDeviceStatus: 结束当前状态后，自动创建停机状态(status=0)
//
//	这样设计的好处:
//	  - 设备始终有一个活动状态（end_time=NULL）
//	  - 开机自动结束停机状态
//	  - 停机自动结束运行状态
//	  - 避免状态冲突和遗漏
//
// 使用场景:
//
//	任务22: #1停机状态 FALSE_TO_TRUE → EndDeviceStatus(1, "停机")
//	任务23: #1开机状态 FALSE_TO_TRUE → UpdateDeviceStatus(1, 1, "运行")
//
// 何时修改此文件:
//   - 需要添加新的状态类型（如维修、调试）
//   - 需要修改状态切换逻辑
//   - 需要添加状态统计功能
//
// ============================================================================
package database

import (
	"fmt"
	"time"

	"gin-mqtt-pgsql/models"
)

// ========================================================
// 设备状态管理
// ========================================================

// UpdateDeviceStatus 更新设备状态（自动结束上一个状态，创建新状态）
func UpdateDeviceStatus(deviceID int, newStatus int8, remark *string) (*models.SysDeviceStatus, error) {
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// 1. 查询所有未结束的活动记录，加 FOR UPDATE 行级锁
	// 防止多个并发事务同时读到相同快照后各自插入，造成重复的 end_time IS NULL 记录
	var activeRecords []models.SysDeviceStatus
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("device_id = ? AND end_time IS NULL", deviceID).
		Order("id ASC").Find(&activeRecords).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("查询活动状态失败: %w", err)
	}

	// 如果只有一条且状态相同，直接返回，不重复插入
	if len(activeRecords) == 1 && activeRecords[0].Status == newStatus {
		tx.Rollback()
		return &activeRecords[0], nil
	}

	// 批量结束所有未结束的记录（兜底处理已有的多条脏数据）
	if len(activeRecords) > 0 {
		for _, rec := range activeRecords {
			duration := int(now.Sub(rec.StartTime).Minutes())
			if err := tx.Model(&models.SysDeviceStatus{}).
				Where("id = ?", rec.ID).
				Updates(map[string]interface{}{
					"end_time":     now,
					"duration_min": duration,
				}).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("结束当前状态失败: %w", err)
			}
		}
	}

	// 2. 创建新状态记录
	newStatusRecord := &models.SysDeviceStatus{
		DeviceID:    deviceID,
		Status:      newStatus,
		StartTime:   now,
		EndTime:     nil,
		DurationMin: 0,
		Remark:      remark,
	}

	if err := tx.Create(newStatusRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建新状态失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	DB.Preload("Device").First(newStatusRecord, newStatusRecord.ID)
	return newStatusRecord, nil
}

// GetDeviceCurrentStatus 获取设备当前状态
func GetDeviceCurrentStatus(deviceID int) (*models.SysDeviceStatus, error) {
	var status models.SysDeviceStatus
	err := DB.Preload("Device").
		Where("device_id = ? AND end_time IS NULL", deviceID).
		First(&status).Error
	if err != nil {
		return nil, fmt.Errorf("查询设备当前状态失败: %w", err)
	}
	return &status, nil
}

// GetDeviceStatusHistory 获取设备状态历史记录
func GetDeviceStatusHistory(deviceID int, startTime, endTime *time.Time) ([]*models.SysDeviceStatus, error) {
	var statusList []*models.SysDeviceStatus
	query := DB.Preload("Device").Where("device_id = ?", deviceID)

	// 时间范围筛选 - 查询与时间范围有重叠的记录
	if startTime != nil && endTime != nil {
		// 记录的结束时间 > 查询开始时间 AND 记录的开始时间 <= 查询结束时间
		// 使用 <= 而不是 < 以包含在结束时间点开始的记录
		query = query.Where("(end_time IS NULL OR end_time > ?) AND start_time <= ?", *startTime, *endTime)
	} else if startTime != nil {
		// 只有开始时间：记录的结束时间 > 查询开始时间 OR 记录还未结束
		query = query.Where("end_time IS NULL OR end_time > ?", *startTime)
	} else if endTime != nil {
		// 只有结束时间：记录的开始时间 <= 查询结束时间
		query = query.Where("start_time <= ?", *endTime)
	}

	result := query.Order("start_time DESC").Find(&statusList)
	if result.Error != nil {
		return nil, fmt.Errorf("查询设备状态历史失败: %w", result.Error)
	}
	return statusList, nil
}

// GetAllDevicesStatus 获取所有设备的当前状态
func GetAllDevicesStatus() ([]*models.SysDeviceStatus, error) {
	var statusList []*models.SysDeviceStatus

	// 查询所有end_time为NULL的记录（当前活动状态）
	err := DB.Preload("Device").
		Where("end_time IS NULL").
		Order("device_id").
		Find(&statusList).Error

	if err != nil {
		return nil, fmt.Errorf("查询所有设备状态失败: %w", err)
	}
	return statusList, nil
}

// GetDeviceStatusSummary 获取设备状态统计（含今日各状态时长）
func GetDeviceStatusSummary(deviceID int) (*models.DeviceStatusSummary, error) {
	// 1. 获取设备信息
	device, err := GetDeviceByID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("设备不存在: %w", err)
	}

	// 2. 获取当前状态
	currentStatus, err := GetDeviceCurrentStatus(deviceID)
	var currentStatusValue int8 = 0
	var currentStartTime *time.Time
	var currentDuration int = 0

	if err == nil && currentStatus != nil {
		currentStatusValue = currentStatus.Status
		currentStartTime = &currentStatus.StartTime
		currentDuration = int(time.Since(currentStatus.StartTime).Minutes())
	}

	// 3. 统计今日各状态时长
	todayStart := time.Now().Truncate(24 * time.Hour)
	var statusRecords []models.SysDeviceStatus
	DB.Where("device_id = ? AND start_time >= ?", deviceID, todayStart).
		Order("start_time ASC").
		Find(&statusRecords)

	runningMin := 0
	idleMin := 0
	faultMin := 0

	for _, record := range statusRecords {
		endTime := time.Now()
		if record.EndTime != nil {
			endTime = *record.EndTime
		}
		duration := int(endTime.Sub(record.StartTime).Minutes())

		switch record.Status {
		case 0: // 空闲
			idleMin += duration
		case 1: // 运行
			runningMin += duration
		case 2: // 故障
			faultMin += duration
		}
	}

	// 4. 计算利用率（运行时间 / 总时间）
	totalMin := runningMin + idleMin + faultMin
	utilization := 0.0
	if totalMin > 0 {
		utilization = float64(runningMin) / float64(totalMin) * 100
	}

	// 5. 状态名称映射
	statusNames := map[int8]string{
		0: "空闲",
		1: "运行",
		2: "故障",
	}

	summary := &models.DeviceStatusSummary{
		DeviceID:      deviceID,
		DeviceName:    device.DeviceName,
		CurrentStatus: currentStatusValue,
		StatusName:    statusNames[currentStatusValue],
		StartTime:     currentStartTime,
		DurationMin:   currentDuration,
		RunningMin:    runningMin,
		IdleMin:       idleMin,
		FaultMin:      faultMin,
		Utilization:   utilization,
	}

	return summary, nil
}

// GetAllDevicesStatusSummary 获取所有设备的状态统计（从设备表获取，确保返回所有设备）- 优化版
func GetAllDevicesStatusSummary() ([]*models.DeviceStatusSummary, error) {
	todayStart := time.Now().Truncate(24 * time.Hour)

	// 一次SQL查询获取所有数据（避免N+1问题）
	type DeviceStats struct {
		DeviceID      int
		DeviceName    string
		CurrentStatus *int8
		StartTime     *time.Time
		RunningMin    int
		IdleMin       int
		FaultMin      int
	}

	var stats []DeviceStats

	// 从设备表开始查询，确保返回所有设备（即使没有状态记录）
	// 使用子查询取每台设备最新的未结束状态（防止脏数据导致同一设备出现多行）
	err := DB.Raw(`
		SELECT 
			d.id as device_id,
			d.device_name,
			latest_curr.status as current_status,
			latest_curr.start_time,
			COALESCE(SUM(CASE WHEN hist.status = 1 THEN 
				TIMESTAMPDIFF(MINUTE, hist.start_time, COALESCE(hist.end_time, NOW())) 
			END), 0) as running_min,
			COALESCE(SUM(CASE WHEN hist.status = 0 THEN 
				TIMESTAMPDIFF(MINUTE, hist.start_time, COALESCE(hist.end_time, NOW())) 
			END), 0) as idle_min,
			COALESCE(SUM(CASE WHEN hist.status = 2 THEN 
				TIMESTAMPDIFF(MINUTE, hist.start_time, COALESCE(hist.end_time, NOW())) 
			END), 0) as fault_min
		FROM sys_devices d
		LEFT JOIN (
			SELECT device_id, status, start_time
			FROM sys_device_status
			WHERE end_time IS NULL
			AND id IN (
				SELECT MAX(id) FROM sys_device_status WHERE end_time IS NULL GROUP BY device_id
			)
		) latest_curr ON latest_curr.device_id = d.id
		LEFT JOIN sys_device_status hist ON hist.device_id = d.id AND hist.start_time >= ?
		GROUP BY d.id, d.device_name, latest_curr.status, latest_curr.start_time
		ORDER BY d.id
	`, todayStart).Scan(&stats).Error

	if err != nil {
		return nil, fmt.Errorf("查询设备状态失败: %w", err)
	}

	// 转换为响应格式
	statusNames := map[int8]string{0: "空闲", 1: "运行", 2: "故障"}
	summaries := make([]*models.DeviceStatusSummary, 0, len(stats))

	for _, stat := range stats {
		// 处理没有状态记录的设备
		currentStatus := int8(0) // 默认空闲
		var startTime *time.Time = nil
		currentDuration := 0

		if stat.CurrentStatus != nil {
			currentStatus = *stat.CurrentStatus
			startTime = stat.StartTime
			if startTime != nil {
				currentDuration = int(time.Since(*startTime).Minutes())
			}
		}

		totalMin := stat.RunningMin + stat.IdleMin + stat.FaultMin
		utilization := 0.0
		if totalMin > 0 {
			utilization = float64(stat.RunningMin) / float64(totalMin) * 100
		}

		summaries = append(summaries, &models.DeviceStatusSummary{
			DeviceID:      stat.DeviceID,
			DeviceName:    stat.DeviceName,
			CurrentStatus: currentStatus,
			StatusName:    statusNames[currentStatus],
			StartTime:     startTime,
			DurationMin:   currentDuration,
			RunningMin:    stat.RunningMin,
			IdleMin:       stat.IdleMin,
			FaultMin:      stat.FaultMin,
			Utilization:   utilization,
		})
	}

	return summaries, nil
}

// EndDeviceStatus 手动结束设备当前状态，并创建停机状态（用于停机场景）
// 结束当前状态后，自动创建一个"停机"状态(status=0)
func EndDeviceStatus(deviceID int, remark *string) error {
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	now := time.Now()

	// 1. 查询所有未结束的活动记录，加 FOR UPDATE 行级锁防止并发竞态
	var activeRecords []models.SysDeviceStatus
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("device_id = ? AND end_time IS NULL", deviceID).
		Order("id ASC").Find(&activeRecords).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("查询活动状态失败: %w", err)
	}

	// 如果所有活动记录都已经是停机状态(status=0)，直接返回，不重复插入
	allStopped := true
	for _, rec := range activeRecords {
		if rec.Status != 0 {
			allStopped = false
			break
		}
	}
	if len(activeRecords) == 1 && allStopped {
		tx.Rollback()
		return nil
	}

	// 2. 批量结束所有未结束的记录（修复竞态产生的多条脏数据）
	stopRemark := "自动停机"
	if remark != nil {
		stopRemark = *remark
	}
	for _, rec := range activeRecords {
		duration := int(now.Sub(rec.StartTime).Minutes())
		updates := map[string]interface{}{
			"end_time":     now,
			"duration_min": duration,
			"remark":       stopRemark,
		}
		if err := tx.Model(&models.SysDeviceStatus{}).
			Where("id = ?", rec.ID).
			Updates(updates).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("结束设备状态失败: %w", err)
		}
	}

	// 3. 创建新的停机状态记录 (status=0)
	newStatusRecord := &models.SysDeviceStatus{
		DeviceID:    deviceID,
		Status:      0,
		StartTime:   now,
		EndTime:     nil,
		DurationMin: 0,
		Remark:      &stopRemark,
	}

	if err := tx.Create(newStatusRecord).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建停机状态失败: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetStatusStatistics 获取设备状态统计（指定时间范围）
func GetStatusStatistics(deviceID int, startTime, endTime time.Time) (map[string]interface{}, error) {
	var records []models.SysDeviceStatus
	// 查询与时间范围有重叠的记录（使用 <= 包含结束时间点开始的记录）
	err := DB.Where("device_id = ? AND (end_time IS NULL OR end_time > ?) AND start_time <= ?", deviceID, startTime, endTime).
		Order("start_time ASC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("查询状态记录失败: %w", err)
	}

	runningMin := 0
	idleMin := 0
	faultMin := 0
	statusChanges := 0

	for _, record := range records {
		endTimeVal := time.Now()
		if record.EndTime != nil {
			endTimeVal = *record.EndTime
		}

		// 确保不超出查询范围
		if endTimeVal.After(endTime) {
			endTimeVal = endTime
		}

		duration := int(endTimeVal.Sub(record.StartTime).Minutes())

		switch record.Status {
		case 0:
			idleMin += duration
		case 1:
			runningMin += duration
		case 2:
			faultMin += duration
		}
		statusChanges++
	}

	totalMin := runningMin + idleMin + faultMin
	utilization := 0.0
	if totalMin > 0 {
		utilization = float64(runningMin) / float64(totalMin) * 100
	}

	return map[string]interface{}{
		"running_min":    runningMin,
		"idle_min":       idleMin,
		"fault_min":      faultMin,
		"total_min":      totalMin,
		"utilization":    utilization,
		"status_changes": statusChanges,
	}, nil
}
