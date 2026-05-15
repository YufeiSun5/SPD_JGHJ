package service

import (
	"encoding/json"
	"fmt"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
)

// DeviceStatusData is the operator-facing device status DTO.
// CN: 面向页面/小屏的设备状态 DTO，隐藏底层状态表与班次人员表的拼接细节。
// EN: Device status DTO for UI/small screens, hiding joins across status, shift and staff tables.
// JP: UI/小画面向け設備状態 DTO。状態表・班次・人員表の結合詳細を隠す。
type DeviceStatusData struct {
	DeviceID      int        `json:"device_id"`
	DeviceName    string     `json:"device_name"`
	DeviceCode    string     `json:"device_code"`
	CurrentStatus int8       `json:"current_status"`
	StatusName    string     `json:"status_name"`
	StartTime     *time.Time `json:"start_time"`
	DurationMin   int        `json:"duration_min"`
	RunningMin    int        `json:"running_min"`
	IdleMin       int        `json:"idle_min"`
	FaultMin      int        `json:"fault_min"`
	Utilization   float64    `json:"utilization"`
	Operators     string     `json:"operators"`
	RecordTime    string     `json:"record_time"`
	Temperature   string     `json:"temperature"`
	Humidity      string     `json:"humidity"`
	Remark        string     `json:"remark"`
}

type ExtraData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Speed       int     `json:"speed"`
}

type DeviceStatusHistoryData struct {
	models.SysDeviceStatus
	TeamName  string `json:"team_name"`
	Operators string `json:"operators"`
}

func GetAllDevicesStatus() ([]DeviceStatusData, error) {
	summaries, err := database.GetAllDevicesStatusSummary()
	if err != nil {
		return nil, fmt.Errorf("获取设备状态失败: %v", err)
	}

	result := make([]DeviceStatusData, 0, len(summaries))
	for _, summary := range summaries {
		data := DeviceStatusData{
			DeviceID:      summary.DeviceID,
			DeviceName:    summary.DeviceName,
			CurrentStatus: summary.CurrentStatus,
			StatusName:    summary.StatusName,
			StartTime:     summary.StartTime,
			DurationMin:   summary.DurationMin,
			RunningMin:    summary.RunningMin,
			IdleMin:       summary.IdleMin,
			FaultMin:      summary.FaultMin,
			Utilization:   summary.Utilization,
		}

		if device, err := database.GetDeviceByID(summary.DeviceID); err == nil {
			data.DeviceCode = device.DeviceCode
		}

		data.Operators = activeSessionOperators(summary.DeviceID)
		if data.Operators == "" {
			data.Operators = "-"
		}

		if data.StartTime != nil {
			data.RecordTime = data.StartTime.Format("2006-01-02 15:04")
		} else {
			data.RecordTime = "-"
		}

		if currentStatus, err := database.GetDeviceCurrentStatus(summary.DeviceID); err == nil && currentStatus != nil {
			if currentStatus.ExtraData != nil && *currentStatus.ExtraData != "" {
				var extraData ExtraData
				if err := json.Unmarshal([]byte(*currentStatus.ExtraData), &extraData); err == nil {
					if extraData.Temperature > 0 {
						data.Temperature = fmt.Sprintf("%.1f°C", extraData.Temperature)
					}
					if extraData.Humidity > 0 {
						data.Humidity = fmt.Sprintf("%.1f%%", extraData.Humidity)
					}
				}
			}
			if currentStatus.Remark != nil {
				data.Remark = *currentStatus.Remark
			}
		}

		if data.Temperature == "" {
			data.Temperature = "-"
		}
		if data.Humidity == "" {
			data.Humidity = "-"
		}
		if data.Remark == "" {
			data.Remark = "-"
		}

		result = append(result, data)
	}

	return result, nil
}

func GetDeviceStatusHistory(deviceID int) ([]*models.SysDeviceStatus, error) {
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	history, err := database.GetDeviceStatusHistory(deviceID, &startTime, &endTime)
	if err != nil {
		return nil, fmt.Errorf("获取状态历史失败: %v", err)
	}
	return history, nil
}

func GetDeviceStatusHistoryAll(deviceID *int, startTimeStr, endTimeStr string) ([]DeviceStatusHistoryData, error) {
	startTime := parseLocalDateTimePtr(startTimeStr)
	endTime := parseLocalDateTimePtr(endTimeStr)

	query := database.DB.Preload("Device").Order("start_time DESC")
	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}
	if startTime != nil && endTime != nil {
		query = query.Where("(end_time IS NULL OR end_time > ?) AND start_time <= ?", *startTime, *endTime)
	} else if startTime != nil {
		query = query.Where("end_time IS NULL OR end_time > ?", *startTime)
	} else if endTime != nil {
		query = query.Where("start_time <= ?", *endTime)
	}

	var records []*models.SysDeviceStatus
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("查询历史记录失败: %v", err)
	}

	result := make([]DeviceStatusHistoryData, 0, len(records))
	for _, record := range records {
		data := DeviceStatusHistoryData{SysDeviceStatus: *record}

		var session models.ProMachineSession
		err := database.DB.Preload("Team").
			Where("device_id = ? AND login_time <= ? AND (logout_time IS NULL OR logout_time >= ?)",
				record.DeviceID, record.StartTime, record.StartTime).
			Order("login_time DESC").
			First(&session).Error
		if err == nil {
			if session.Team != nil {
				data.TeamName = session.Team.TeamName
			}
			data.Operators = staffNamesFromJSON(session.StaffIDs)
		}

		result = append(result, data)
	}

	return result, nil
}

func GetDeviceStatusStats() (map[string]interface{}, error) {
	summaries, err := database.GetAllDevicesStatusSummary()
	if err != nil {
		return nil, err
	}

	runningCount := 0
	idleCount := 0
	faultCount := 0
	totalUtilization := 0.0
	for _, summary := range summaries {
		switch summary.CurrentStatus {
		case 1:
			runningCount++
		case 0:
			idleCount++
		case 2:
			faultCount++
		}
		totalUtilization += summary.Utilization
	}

	avgUtilization := 0.0
	if len(summaries) > 0 {
		avgUtilization = totalUtilization / float64(len(summaries))
	}

	return map[string]interface{}{
		"running_count":   runningCount,
		"idle_count":      idleCount,
		"fault_count":     faultCount,
		"total_count":     len(summaries),
		"avg_utilization": avgUtilization,
	}, nil
}

func parseLocalDateTimePtr(value string) *time.Time {
	if value == "" {
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02T15:04:05", value, time.Local)
	if err != nil {
		t, err = time.ParseInLocation("2006-01-02T15:04", value, time.Local)
	}
	if err != nil {
		return nil
	}
	return &t
}

func activeSessionOperators(deviceID int) string {
	session, err := database.GetActiveSession(deviceID)
	if err != nil || session == nil {
		return ""
	}
	return staffNamesFromJSON(session.StaffIDs)
}

func staffNamesFromJSON(staffIDsJSON string) string {
	var staffIDs []int
	if err := json.Unmarshal([]byte(staffIDsJSON), &staffIDs); err != nil {
		return ""
	}

	names := ""
	for _, staffID := range staffIDs {
		staff, err := database.GetStaffByID(staffID)
		if err != nil {
			continue
		}
		if names == "" {
			names = staff.Name
		} else {
			names += "、" + staff.Name
		}
	}
	return names
}
