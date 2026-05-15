package service

import (
	"encoding/json"
	"fmt"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
)

// GetAllDevices 获取所有设备。
func GetAllDevices() ([]*models.SysDevice, error) {
	return database.GetAllDevices()
}

// GetAllStaff 获取所有员工。
func GetAllStaff(teamID *int, isActive *int8) ([]*models.SysStaff, error) {
	return database.GetAllStaff(teamID, isActive)
}

// CreateStaff 创建员工。
// CN: 员工默认在职；创建后重新加载班组关联，保持前端原返回结构不变。
// EN: Staff are active by default; reload after create to preserve the frontend response shape.
// JP: 従業員は既定で在職。作成後に班組関連を再読み込みし、フロントの戻り構造を維持する。
func CreateStaff(staffCode, name string, currentTeamID *int) (*models.SysStaff, error) {
	staff := &models.SysStaff{
		StaffCode:     staffCode,
		Name:          name,
		CurrentTeamID: currentTeamID,
		IsActive:      1,
	}
	if err := database.CreateStaff(staff); err != nil {
		return nil, err
	}
	staff, _ = database.GetStaffByID(staff.ID)
	return staff, nil
}

func UpdateStaff(id int, name *string, currentTeamID *int, isActive *int8) error {
	updates := make(map[string]interface{})
	if name != nil {
		updates["name"] = *name
	}
	if currentTeamID != nil {
		if *currentTeamID == -1 {
			updates["current_team_id"] = nil
		} else {
			updates["current_team_id"] = *currentTeamID
		}
	}
	if isActive != nil {
		updates["is_active"] = *isActive
	}
	return database.UpdateStaff(id, updates)
}

func DeleteStaff(id int) error {
	return database.DeleteStaff(id)
}

func TransferStaff(staffID, newTeamID int, operatorName *string) error {
	req := &models.TransferStaffRequest{
		StaffID:      staffID,
		NewTeamID:    newTeamID,
		OperatorName: operatorName,
	}
	return database.TransferStaff(req)
}

func GetStaffHistory(staffID int) ([]*models.SysStaffHistory, error) {
	return database.GetStaffHistory(staffID)
}

func GetAllTeams(status *int8) ([]*models.SysTeam, error) {
	return database.GetAllTeams(status)
}

func CreateTeam(teamName string, leaderName *string) (*models.SysTeam, error) {
	team := &models.SysTeam{
		TeamName:   teamName,
		LeaderName: leaderName,
		Status:     1,
	}
	if err := database.CreateTeam(team); err != nil {
		return nil, err
	}
	return team, nil
}

func UpdateTeam(id int, teamName *string, leaderName *string, status *int8) error {
	updates := make(map[string]interface{})
	if teamName != nil {
		updates["team_name"] = *teamName
	}
	if leaderName != nil {
		updates["leader_name"] = *leaderName
	}
	if status != nil {
		updates["status"] = *status
	}
	return database.UpdateTeam(id, updates)
}

func DeleteTeam(id int) error {
	return database.DeleteTeam(id)
}

// DeviceLogin 设备登录/上班打卡。
// CN: StaffIDs 仍按历史 JSON 字符串入库，避免改变 pro_machine_sessions 兼容格式。
// EN: StaffIDs remain stored as the historical JSON string to preserve pro_machine_sessions compatibility.
// JP: StaffIDs は既存互換のため従来通り JSON 文字列として pro_machine_sessions に保存する。
func DeviceLogin(deviceID, teamID int, staffIDs []int) (*models.ProMachineSession, error) {
	fmt.Printf("🔐 DeviceLogin 被调用: deviceID=%d, teamID=%d, staffIDs=%v\n", deviceID, teamID, staffIDs)

	staffIDsJSON, err := json.Marshal(staffIDs)
	if err != nil {
		fmt.Printf("❌ 序列化员工列表失败: %v\n", err)
		return nil, fmt.Errorf("序列化员工列表失败: %v", err)
	}

	session := &models.ProMachineSession{
		DeviceID: deviceID,
		TeamID:   teamID,
		StaffIDs: string(staffIDsJSON),
	}

	fmt.Printf("📝 准备写入班次记录: %+v\n", session)
	if err := database.DeviceLogin(session); err != nil {
		fmt.Printf("❌ DeviceLogin 失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ 班次记录创建成功，ID: %d\n", session.ID)
	result, err := database.GetSessionByID(session.ID)
	if err != nil {
		fmt.Printf("❌ 获取班次详情失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ 返回班次信息: %+v\n", result)
	return result, nil
}

func DeviceLogout(deviceID int) (*models.ProMachineSession, error) {
	return database.DeviceLogout(deviceID)
}

func GetActiveSession(deviceID int) (*models.ProMachineSession, error) {
	return database.GetActiveSession(deviceID)
}

func GetAllActiveSessions() ([]*models.ProMachineSession, error) {
	var sessions []*models.ProMachineSession
	err := database.DB.Preload("Team").
		Where("logout_time IS NULL").
		Order("device_id").
		Find(&sessions).Error
	if err != nil {
		return nil, fmt.Errorf("查询活动班次失败: %v", err)
	}
	return sessions, nil
}

func parseDateRange(startDate, endDate string) (*time.Time, *time.Time) {
	var start, end *time.Time
	if startDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", startDate, time.Local); err == nil {
			start = &t
		}
	}
	if endDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", endDate, time.Local); err == nil {
			endTime := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			end = &endTime
		}
	}
	return start, end
}

func GetSessionHistory(deviceID *int, teamID *int, startDate, endDate string) ([]*models.ProMachineSession, error) {
	start, end := parseDateRange(startDate, endDate)
	return database.GetSessionHistory(deviceID, teamID, start, end)
}

func GetSessionStats(sessionID int64) (*models.SessionStatusResponse, error) {
	return database.GetSessionStats(sessionID)
}

func GetStaffAttendance(staffID int, startDate, endDate string) ([]*models.ProMachineSession, error) {
	start, end := parseDateRange(startDate, endDate)
	return database.GetStaffAttendance(staffID, start, end)
}
