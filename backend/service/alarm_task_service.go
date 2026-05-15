package service

import (
	"fmt"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
	"gin-mqtt-pgsql/workers"
)

type AlarmRecordData struct {
	ID         int64      `json:"id"`
	VarID      int64      `json:"var_id"`
	VarName    string     `json:"var_name"`
	Val        float64    `json:"val"`
	AlarmType  string     `json:"alarm_type"`
	LimitValue float64    `json:"limit_value"`
	Msg        string     `json:"msg"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	AckStatus  int        `json:"ack_status"`
	Duration   string     `json:"duration"`
}

type HourlyAlarmCount struct {
	Hour       int    `json:"hour"`
	AlarmCount int64  `json:"alarm_count"`
	TimeSlot   string `json:"time_slot"`
}

type VariableOption struct {
	VarID       int64  `json:"var_id"`
	VarName     string `json:"var_name"`
	DisplayName string `json:"display_name"`
	DeviceName  string `json:"device_name"`
	GatewayName string `json:"gateway_name"`
}

type Gateway struct {
	ID     int    `json:"id"`
	GwName string `json:"gw_name"`
	Status int    `json:"status"`
}

func GetAlarmRecords(ackStatus *int, startTimeStr, endTimeStr string, varID *int64) ([]AlarmRecordData, error) {
	startTime := parseLocalDateTimePtr(startTimeStr)
	endTime := parseLocalDateTimePtr(endTimeStr)

	query := database.DB.Table("sys_alarm_records AS a").
		Select(`a.id, a.var_id, a.var_name, a.val, a.alarm_type, a.limit_value, 
		        CASE 
		            WHEN a.alarm_type = 'SYS' THEN e.error_msg 
		            ELSE a.msg 
		        END as msg,
		        a.start_time, a.end_time, a.ack_status`).
		Joins("LEFT JOIN sys_error_codes AS e ON a.alarm_type = 'SYS' AND CAST(a.val AS SIGNED) = e.error_code").
		Order("a.id DESC")

	if ackStatus != nil {
		query = query.Where("a.ack_status = ?", *ackStatus)
	}
	if varID != nil && *varID > 0 {
		query = query.Where("a.var_id = ?", *varID)
	}
	if startTime != nil {
		query = query.Where("a.start_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("a.start_time <= ?", *endTime)
	}

	type alarmRecordWithErrorMsg struct {
		ID         int64      `gorm:"column:id"`
		VarID      int64      `gorm:"column:var_id"`
		VarName    *string    `gorm:"column:var_name"`
		Val        *float64   `gorm:"column:val"`
		AlarmType  string     `gorm:"column:alarm_type"`
		LimitValue *float64   `gorm:"column:limit_value"`
		Msg        *string    `gorm:"column:msg"`
		StartTime  time.Time  `gorm:"column:start_time"`
		EndTime    *time.Time `gorm:"column:end_time"`
		AckStatus  int        `gorm:"column:ack_status"`
	}

	var records []alarmRecordWithErrorMsg
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("查询报警记录失败: %v", err)
	}

	result := make([]AlarmRecordData, 0, len(records))
	for _, record := range records {
		data := AlarmRecordData{
			ID:        record.ID,
			VarID:     record.VarID,
			AlarmType: record.AlarmType,
			StartTime: record.StartTime,
			EndTime:   record.EndTime,
			AckStatus: record.AckStatus,
		}
		if record.VarName != nil {
			data.VarName = *record.VarName
		}
		if record.Val != nil {
			data.Val = *record.Val
		}
		if record.LimitValue != nil {
			data.LimitValue = *record.LimitValue
		}
		if record.Msg != nil {
			data.Msg = *record.Msg
		}
		data.Duration = formatAlarmDuration(record.StartTime, record.EndTime)
		result = append(result, data)
	}

	return result, nil
}

func AckAlarm(alarmID int64) error {
	result := database.DB.Table("sys_alarm_records").Where("id = ?", alarmID).Update("ack_status", 1)
	if result.Error != nil {
		return fmt.Errorf("确认报警失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("报警记录不存在")
	}
	return nil
}

func GetTodayUnacknowledgedAlarmCount() (int64, error) {
	todayStart := beginningOfToday()
	var count int64
	err := database.DB.Table("sys_alarm_records").
		Where("ack_status = ?", 0).
		Where("start_time >= ?", todayStart).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询今日未确认报警数失败: %v", err)
	}
	return count, nil
}

func GetTodayHourlyAlarmCount() ([]HourlyAlarmCount, error) {
	todayStart := beginningOfToday()
	todayEnd := todayStart.Add(24 * time.Hour)

	var results []struct {
		Hour       int   `gorm:"column:hour"`
		AlarmCount int64 `gorm:"column:alarm_count"`
	}
	err := database.DB.Table("sys_alarm_records").
		Select("EXTRACT(HOUR FROM start_time) as hour, COUNT(*) as alarm_count").
		Where("start_time >= ? AND start_time < ?", todayStart, todayEnd).
		Group("EXTRACT(HOUR FROM start_time)").
		Order("hour ASC").
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("查询每小时报警数失败: %v", err)
	}

	hourlyData := make([]HourlyAlarmCount, 0, len(results))
	for _, r := range results {
		hourlyData = append(hourlyData, HourlyAlarmCount{
			Hour:       r.Hour,
			AlarmCount: r.AlarmCount,
			TimeSlot:   fmt.Sprintf("%d:00", r.Hour),
		})
	}
	return hourlyData, nil
}

func GetActiveAlarmCount() (int64, error) {
	var activeOrders []models.ProOrder
	err := database.DB.Where("status = ?", 1).Find(&activeOrders).Error
	if err != nil {
		return 0, fmt.Errorf("查询活动工单失败: %v", err)
	}
	if len(activeOrders) == 0 {
		return 0, nil
	}

	deviceIDs := make([]int, 0)
	for _, order := range activeOrders {
		if order.TargetDeviceID != nil {
			deviceIDs = append(deviceIDs, *order.TargetDeviceID)
		}
	}
	if len(deviceIDs) == 0 {
		return 0, nil
	}

	var varIDs []int64
	err = database.DB.Table("sys_variables").Where("device_id IN ?", deviceIDs).Pluck("id", &varIDs).Error
	if err != nil {
		return 0, fmt.Errorf("查询设备变量失败: %v", err)
	}
	if len(varIDs) == 0 {
		return 0, nil
	}

	var earliestStartTime *time.Time
	for _, order := range activeOrders {
		if order.StartTime != nil && (earliestStartTime == nil || order.StartTime.Before(*earliestStartTime)) {
			earliestStartTime = order.StartTime
		}
	}
	if earliestStartTime == nil {
		todayStart := beginningOfToday()
		earliestStartTime = &todayStart
	}

	var count int64
	err = database.DB.Table("sys_alarm_records").
		Where("var_id IN ? AND start_time >= ?", varIDs, earliestStartTime).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("查询报警记录失败: %v", err)
	}
	return count, nil
}

func GetAlarmStats() (map[string]interface{}, error) {
	todayStart := beginningOfToday()
	yesterdayStart := todayStart.Add(-24 * time.Hour)

	var todayTotal int64
	database.DB.Table("sys_alarm_records").Where("start_time >= ?", todayStart).Count(&todayTotal)

	var todayAcked int64
	database.DB.Table("sys_alarm_records").Where("start_time >= ? AND ack_status = 1", todayStart).Count(&todayAcked)

	var todayPending int64
	database.DB.Table("sys_alarm_records").Where("start_time >= ? AND ack_status = 0", todayStart).Count(&todayPending)

	var yesterdayTotal int64
	database.DB.Table("sys_alarm_records").Where("start_time >= ? AND start_time < ?", yesterdayStart, todayStart).Count(&yesterdayTotal)

	comparison := "equal"
	var comparisonValue int64
	if yesterdayTotal == 0 {
		if todayTotal > 0 {
			comparison = "up"
			comparisonValue = todayTotal
		}
	} else {
		diff := todayTotal - yesterdayTotal
		if diff > 0 {
			comparison = "up"
			comparisonValue = diff
		} else if diff < 0 {
			comparison = "down"
			comparisonValue = -diff
		}
	}

	return map[string]interface{}{
		"today_total":      todayTotal,
		"today_acked":      todayAcked,
		"today_pending":    todayPending,
		"yesterday_total":  yesterdayTotal,
		"comparison":       comparison,
		"comparison_value": comparisonValue,
	}, nil
}

func GetVariableOptions() ([]VariableOption, error) {
	var options []VariableOption
	err := database.DB.Table("sys_variables v").
		Select("v.id as var_id, v.var_name, COALESCE(v.display_name, v.var_name) as display_name, d.device_name, g.gw_name as gateway_name").
		Joins("INNER JOIN sys_devices d ON v.device_id = d.id").
		Joins("INNER JOIN sys_gateways g ON d.gateway_id = g.id").
		Where("g.status = ?", 1).
		Order("g.gw_name, d.device_name, v.var_name").
		Scan(&options).Error
	if err != nil {
		return nil, fmt.Errorf("查询变量列表失败: %v", err)
	}
	return options, nil
}

func GetAllTasks() ([]*models.Task, error) {
	var tasks []*models.Task
	err := database.DB.Order("task_id DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func CreateTask(task *models.Task) error {
	return database.DB.Create(task).Error
}

func UpdateTask(taskID int64, updates *models.Task) error {
	return database.DB.Model(&models.Task{}).Where("task_id = ?", taskID).Updates(updates).Error
}

func TriggerTaskManually(taskID int64) error {
	scheduler := workers.GetTaskScheduler()
	if scheduler == nil {
		return fmt.Errorf("任务调度器未初始化")
	}
	if err := scheduler.ManualTriggerTask(taskID, 0, nil); err != nil {
		return fmt.Errorf("触发任务失败: %v", err)
	}
	return nil
}

func DeleteTask(taskID int64) error {
	return database.DB.Delete(&models.Task{}, taskID).Error
}

func EnableTask(taskID int64) error {
	return database.DB.Model(&models.Task{}).Where("task_id = ?", taskID).Update("is_enabled", true).Error
}

func DisableTask(taskID int64) error {
	return database.DB.Model(&models.Task{}).Where("task_id = ?", taskID).Update("is_enabled", false).Error
}

func GetAllGateways() ([]Gateway, error) {
	var gateways []Gateway
	err := database.DB.Table("sys_gateways").Select("id, gw_name, status").Order("id").Find(&gateways).Error
	if err != nil {
		return nil, fmt.Errorf("查询网关失败: %v", err)
	}
	return gateways, nil
}

func formatAlarmDuration(start time.Time, end *time.Time) string {
	duration := time.Since(start)
	suffix := " (报警中)"
	if end != nil {
		duration = end.Sub(start)
		suffix = ""
	}
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%s", hours, minutes, suffix)
	}
	return fmt.Sprintf("%d分%s", minutes, suffix)
}

func beginningOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
