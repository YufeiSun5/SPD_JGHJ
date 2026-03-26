package database

import (
	"encoding/json"
	"fmt"
	"time"

	"gin-mqtt-pgsql/models"
)

// ========================================================
// 设备登录/班次管理 (Session Management)
// ========================================================

// DeviceLogin 设备登录/上班打卡
func DeviceLogin(session *models.ProMachineSession) error {
	fmt.Printf("[DeviceLogin] 检查设备 %d 是否有活动班次...\n", session.DeviceID)

	// 1. 检查该设备是否已有正在进行的班次
	var activeSession models.ProMachineSession
	err := DB.Where("device_id = ? AND logout_time IS NULL", session.DeviceID).First(&activeSession).Error
	if err == nil {
		fmt.Printf("[DeviceLogin] ❌ 设备 %d 已有活动班次 ID=%d\n", session.DeviceID, activeSession.ID)
		return fmt.Errorf("设备 %d 已有正在进行的班次（Session ID: %d），请先下班签退", session.DeviceID, activeSession.ID)
	}

	fmt.Printf("[DeviceLogin] ✅ 设备空闲，可以登录\n")

	// 2. 插入新的班次记录
	session.LoginTime = time.Now()
	session.LogoutTime = nil
	session.DurationMin = 0

	fmt.Printf("[DeviceLogin] 准备插入数据: DeviceID=%d, TeamID=%d, StaffIDs=%s\n",
		session.DeviceID, session.TeamID, session.StaffIDs)

	err = DB.Create(session).Error
	if err != nil {
		fmt.Printf("[DeviceLogin] ❌ 插入失败: %v\n", err)
		return err
	}

	fmt.Printf("[DeviceLogin] ✅ 插入成功，Session ID=%d\n", session.ID)
	return nil
}

// DeviceLogout 设备登出/下班打卡
func DeviceLogout(deviceID int) (*models.ProMachineSession, error) {
	// 1. 查找当前正在进行的班次
	var session models.ProMachineSession
	err := DB.Where("device_id = ? AND logout_time IS NULL", deviceID).First(&session).Error
	if err != nil {
		return nil, fmt.Errorf("设备 %d 没有正在进行的班次，无需签退", deviceID)
	}

	// 2. 更新登出时间和时长
	now := time.Now()
	duration := int(now.Sub(session.LoginTime).Minutes())

	err = DB.Model(&session).Updates(map[string]interface{}{
		"logout_time":  now,
		"duration_min": duration,
	}).Error

	if err != nil {
		return nil, err
	}

	// 3. 返回更新后的记录
	session.LogoutTime = &now
	session.DurationMin = duration
	return &session, nil
}

// GetActiveSession 获取设备当前活动的班次
func GetActiveSession(deviceID int) (*models.ProMachineSession, error) {
	var session models.ProMachineSession
	err := DB.Preload("Team").
		Where("device_id = ? AND logout_time IS NULL", deviceID).
		First(&session).Error

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetSessionByID 根据ID获取班次记录
func GetSessionByID(id int64) (*models.ProMachineSession, error) {
	var session models.ProMachineSession
	err := DB.Preload("Team").First(&session, id).Error
	return &session, err
}

// GetSessionHistory 获取设备的班次历史记录
func GetSessionHistory(deviceID *int, teamID *int, startDate, endDate *time.Time) ([]*models.ProMachineSession, error) {
	query := DB.Preload("Team").Order("login_time DESC")

	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}

	if teamID != nil {
		query = query.Where("team_id = ?", *teamID)
	}

	if startDate != nil {
		query = query.Where("login_time >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("login_time <= ?", *endDate)
	}

	var sessions []*models.ProMachineSession
	err := query.Find(&sessions).Error
	return sessions, err
}

// GetStaffAttendance 获取员工出勤记录
func GetStaffAttendance(staffID int, startDate, endDate *time.Time) ([]*models.ProMachineSession, error) {
	query := DB.Preload("Team").Order("login_time DESC")

	// 使用 JSON 查询功能查找包含该员工ID的记录
	// MySQL: JSON_CONTAINS(staff_ids, '101')
	query = query.Where("JSON_CONTAINS(staff_ids, ?)", fmt.Sprintf("%d", staffID))

	if startDate != nil {
		query = query.Where("login_time >= ?", *startDate)
	}

	if endDate != nil {
		query = query.Where("login_time <= ?", *endDate)
	}

	var sessions []*models.ProMachineSession
	err := query.Find(&sessions).Error
	return sessions, err
}

// GetSessionStats 获取班次统计信息
func GetSessionStats(sessionID int64) (*models.SessionStatusResponse, error) {
	// 1. 获取班次基本信息
	session, err := GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	// 2. 解析员工ID列表
	var staffIDs []int
	if err := json.Unmarshal([]byte(session.StaffIDs), &staffIDs); err != nil {
		return nil, fmt.Errorf("解析员工列表失败: %v", err)
	}

	// 3. 查询员工详情
	var staffList []*models.SysStaff
	if len(staffIDs) > 0 {
		DB.Where("id IN ?", staffIDs).Find(&staffList)
	}

	// 4. 计算工作时长
	isActive := session.LogoutTime == nil
	var currentDuration int
	if isActive {
		currentDuration = int(time.Since(session.LoginTime).Minutes())
	} else {
		currentDuration = session.DurationMin
	}

	// 5. 统计该班次期间的工单运行时长
	var totalWorkedMin int
	DB.Model(&models.ProProductionRun{}).
		Select("COALESCE(SUM(TIMESTAMPDIFF(MINUTE, start_time, COALESCE(end_time, NOW()))), 0)").
		Where("device_id = ? AND start_time >= ? AND (end_time IS NULL OR end_time <= ?)",
			session.DeviceID,
			session.LoginTime,
			func() time.Time {
				if session.LogoutTime != nil {
					return *session.LogoutTime
				}
				return time.Now()
			}()).
		Scan(&totalWorkedMin)

	// 6. 计算效率
	idleMin := currentDuration - totalWorkedMin
	if idleMin < 0 {
		idleMin = 0
	}

	efficiency := 0.0
	if currentDuration > 0 {
		efficiency = float64(totalWorkedMin) / float64(currentDuration) * 100
	}

	return &models.SessionStatusResponse{
		ProMachineSession: *session,
		IsActive:          isActive,
		StaffList:         staffList,
		WorkedMin:         totalWorkedMin,
		IdleMin:           idleMin,
		Efficiency:        efficiency,
	}, nil
}

// UpdateSessionStaff 更新班次的员工列表（用于临时调整人员）
func UpdateSessionStaff(sessionID int64, staffIDs []int) error {
	staffIDsJSON, err := json.Marshal(staffIDs)
	if err != nil {
		return fmt.Errorf("序列化员工列表失败: %v", err)
	}

	return DB.Model(&models.ProMachineSession{}).
		Where("id = ?", sessionID).
		Update("staff_ids", string(staffIDsJSON)).Error
}
