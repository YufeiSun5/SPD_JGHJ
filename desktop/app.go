package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
	"gin-mqtt-pgsql/workers"
)

// App Wails应用结构
type App struct {
	ctx context.Context
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{}
}

// Startup Wails启动回调
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	// 启动班次快照定时器（班次结束时自动生成生产追溯快照）
	// Start shift snapshot ticker (auto-generate production traceability snapshots at shift end)
	// シフトスナップショットタイマーを起動（シフト終了時に生産追溯スナップショットを自動生成）
	StartSnapshotTicker(a)
}

// SyncErrorCode 同步错误码到MySQL
func (a *App) SyncErrorCode(errorCode int, errorMsg string) error {
	if database.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	var existingCode database.ErrorCode
	result := database.DB.Where("error_code = ?", errorCode).First(&existingCode)

	now := time.Now()

	if result.Error == nil {
		// 错误码已存在，更新
		existingCode.ErrorMsg = errorMsg
		existingCode.UpdatedAt = now
		if err := database.DB.Save(&existingCode).Error; err != nil {
			return fmt.Errorf("更新错误码失败: %v", err)
		}
		fmt.Printf("✅ [错误码同步] 更新错误码 %d\n", errorCode)
	} else {
		// 错误码不存在，插入
		newCode := database.ErrorCode{
			ErrorCode: errorCode,
			ErrorMsg:  errorMsg,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := database.DB.Create(&newCode).Error; err != nil {
			return fmt.Errorf("创建错误码失败: %v", err)
		}
		fmt.Printf("✅ [错误码同步] 创建错误码 %d\n", errorCode)
	}

	return nil
}

// GetAllErrorCodes 获取所有错误码
func (a *App) GetAllErrorCodes() ([]*database.ErrorCode, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var errorCodes []*database.ErrorCode
	result := database.DB.Order("error_code ASC").Find(&errorCodes)
	if result.Error != nil {
		return nil, fmt.Errorf("查询错误码失败: %v", result.Error)
	}

	return errorCodes, nil
}

// GetRealtimeData 获取所有测点实时数据 (线程安全)
func (a *App) GetRealtimeData() []TagData {
	tagManager := core.GetTagManager()
	allTags := tagManager.GetAllTags()

	result := make([]TagData, 0, len(allTags))
	for _, tag := range allTags {
		// 🔧 线程安全: 一次性读取所有值,避免中途被修改
		currentValue := tag.GetValue()
		currentStrValue := tag.GetStringValue()
		alarmState, _ := tag.GetAlarmState()

		data := TagData{
			VarName:     tag.VarName,
			DisplayName: tag.DisplayName,
			DataType:    tag.DataType,
			Unit:        tag.Unit,
			AlarmState:  alarmState,
		}

		// 根据类型格式化值
		if tag.DataType == "STRING" || tag.DataType == "TEXT" {
			if currentStrValue == "" {
				data.Value = "-" // 空字符串显示为横线
			} else {
				data.Value = currentStrValue
			}
		} else {
			// 数值类型: 直接显示值即可
			// 格式化数值: 整数不显示小数点
			if currentValue == float64(int64(currentValue)) {
				data.Value = fmt.Sprintf("%d", int64(currentValue))
			} else {
				data.Value = fmt.Sprintf("%.2f", currentValue)
			}
		}

		result = append(result, data)
	}

	return result
}

// GetChannelStats 获取通道状态
func (a *App) GetChannelStats() map[string]int {
	return core.GetChannelStats()
}

// GetSystemMonitor 获取系统监控数据 (通道队列长度、任务统计)
func (a *App) GetSystemMonitor() map[string]interface{} {
	// 获取通道状态
	channelStats := core.GetChannelStats()

	// 获取任务统计
	scheduler := workers.GetTaskScheduler()
	var taskStats map[string]int
	if scheduler != nil {
		taskStats = scheduler.GetTaskCount()
	} else {
		taskStats = map[string]int{
			"scheduled":   0,
			"data_change": 0,
			"condition":   0,
		}
	}

	// 计算通道容量和使用率
	channelCapacity := map[string]int{
		"logic_chan":   2000,
		"change_chan":  500,
		"cycle_chan":   500,
		"alarm_chan":   200,
		"sse_chan":     100,
		"event_chan":   300,
		"trigger_chan": 10000, // 🔥 超大缓冲，不丢弃任务
	}

	channelUsage := make(map[string]float64)
	for name, current := range channelStats {
		capacity := channelCapacity[name]
		if capacity > 0 {
			channelUsage[name] = float64(current) / float64(capacity) * 100
		}
	}

	// 使用核心模块的健康检查
	alerts := core.CheckChannelHealth()

	return map[string]interface{}{
		"channel_stats":    channelStats,
		"channel_capacity": channelCapacity,
		"channel_usage":    channelUsage,
		"task_stats":       taskStats,
		"alerts":           alerts,
		"timestamp":        time.Now(),
	}
}

// TagData 测点数据传输对象
type TagData struct {
	VarName     string `json:"var_name"`
	DisplayName string `json:"display_name"`
	DataType    string `json:"data_type"`
	Value       string `json:"value"`
	Unit        string `json:"unit"`
	AlarmState  string `json:"alarm_state"`
}

// ========================================================
// MES 工单管理接口
// ========================================================

// GetAllOrders 获取所有工单
func (a *App) GetAllOrders() ([]*models.ProOrder, error) {
	orders, err := database.GetAllOrders(nil, nil)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// CreateOrder 创建工单
func (a *App) CreateOrder(orderNo, productCode string, planQty int, targetDeviceID *int) (*models.ProOrder, error) {
	order := &models.ProOrder{
		OrderNo:        orderNo,
		ProductCode:    productCode,
		TargetDeviceID: targetDeviceID,
		PlanQty:        planQty,
		ActualQty:      0,
		OkQty:          0,
		NgQty:          0,
		Status:         0, // 待产
		Version:        0,
	}

	if err := database.CreateOrder(order); err != nil {
		return nil, err
	}

	return order, nil
}

// UpdateOrder 更新工单
func (a *App) UpdateOrder(id int64, productCode *string, planQty *int, status *int8, targetDeviceID *int) error {
	updates := make(map[string]interface{})

	if productCode != nil {
		updates["product_code"] = *productCode
	}
	if planQty != nil {
		updates["plan_qty"] = *planQty
	}
	if targetDeviceID != nil {
		updates["target_device_id"] = *targetDeviceID
	}
	if status != nil {
		// 先查询当前工单状态
		order, err := database.GetOrderByID(id)
		if err != nil {
			return err
		}

		updates["status"] = *status

		now := time.Now()

		// 开工/继续时：设置当前开始时间
		if *status == 1 {
			updates["current_start_time"] = now
			// 首次开工时：设置开工时间
			if order.StartTime == nil {
				updates["start_time"] = now
			}
		}

		// 暂停时：累加用时
		if *status == 2 && order.Status == 1 && order.CurrentStartTime != nil {
			elapsed := int(time.Since(*order.CurrentStartTime).Seconds())
			updates["used_seconds"] = order.UsedSeconds + elapsed
			updates["current_start_time"] = nil
		}

		// 完工时：累加用时并设置完工时间
		if *status == 3 {
			if order.CurrentStartTime != nil {
				elapsed := int(time.Since(*order.CurrentStartTime).Seconds())
				updates["used_seconds"] = order.UsedSeconds + elapsed
			}
			updates["end_time"] = now
			updates["current_start_time"] = nil
		}
	}

	return database.UpdateOrder(id, updates)
}

// DeleteOrder 删除工单
func (a *App) DeleteOrder(id int64) error {
	return database.DeleteOrder(id)
}

// StartProductionSmart 智能开工（自动暂停该设备的其他工单，自动获取班次信息）
func (a *App) StartProductionSmart(orderID int64) (*models.ProProductionRun, error) {
	// 1. 获取工单信息
	order, err := database.GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("工单不存在: %w", err)
	}

	if order.TargetDeviceID == nil {
		return nil, fmt.Errorf("工单未指定设备")
	}

	deviceID := *order.TargetDeviceID

	// 2. 获取该设备当前活动的班次信息
	session, err := database.GetActiveSession(deviceID)
	if err != nil {
		return nil, fmt.Errorf("设备%d没有活动班次，请先在\"人员管理\"页面进行班次登记", deviceID)
	}

	// 3. 解析班次中的员工ID
	var staffIDs []int
	if err := json.Unmarshal([]byte(session.StaffIDs), &staffIDs); err != nil {
		return nil, fmt.Errorf("解析班次员工信息失败: %w", err)
	}

	// 4. 构建开工请求
	req := &models.StartProductionRequest{
		OrderID:     orderID,
		DeviceID:    deviceID,
		TeamID:      session.TeamID,
		OperatorIDs: staffIDs,
		Remark:      nil,
	}

	// 5. 调用智能开工方法（会自动暂停该设备的其他工单）
	return database.StartProductionSmart(req)
}

// GetRealHourlyProduction 获取真实的每小时产量统计（调用数据访问层）
func (a *App) GetRealHourlyProduction() ([]database.HourlyProductionPulse, error) {
	// ⭐ 调用数据访问层方法
	return database.GetHourlyProductionPulse(nil)
}

// GetHourlyProductionAccurate 获取每小时精确产量统计（完全参数化，无硬编码）
// configs: 设备变量配置数组，每个设备指定产量ID、NG按钮ID和设备名称
// 如果传入nil，使用默认配置（设备1和设备2）
// GetHourlyProductionAccurate 获取逻辑日每小时精确产量统计
// CN: 自动从 GetShiftsForLogicalDay 获取逻辑日期，确保 7:00 前显示昨天同一时段的数据。
// EN: Automatically resolves the logical date from shift config; before first-shift start = yesterday.
// JP: シフト設定から論理日付を自動解決。第1シフト開始前は前日のデータを返す。
func (a *App) GetHourlyProductionAccurate(configs []database.DeviceVarConfig) ([]database.HourlyProductionAccurate, error) {
	ldShifts, err := a.GetShiftsForLogicalDay()
	logicalDate := ""
	if err == nil && len(ldShifts) > 0 {
		logicalDate = ldShifts[0].LogicalDate
	}
	return database.GetHourlyProductionAccurate(configs, logicalDate)
}

// GetMonthlyProductionAccurate 获取当月产量汇总统计（按设备，不走工单表）
func (a *App) GetMonthlyProductionAccurate(configs []database.DeviceVarConfig) ([]database.MonthlyProductionAccurate, error) {
	return database.GetMonthlyProductionAccurate(configs)
}

// GetMonthlyQualityByOrder 从工单表获取当月各设备良品率汇总
func (a *App) GetMonthlyQualityByOrder() ([]database.DeviceQualityStat, error) {
	return database.GetMonthlyQualityByOrder()
}

// GetMonthlyDailyQualityTrend 获取本月每日良品率趋势
func (a *App) GetMonthlyDailyQualityTrend() ([]database.DailyQualityTrend, error) {
	return database.GetMonthlyDailyQualityTrend()
}

// GetDailyQualityByRun 从生产运行记录获取今日各设备良品率
func (a *App) GetDailyQualityByRun() ([]database.DeviceQualityStat, error) {
	return database.GetDailyQualityByRun()
}

// GetActiveOrderQuality 获取当前在产工单（生产中+暂停）各设备良品率
func (a *App) GetActiveOrderQuality() ([]database.DeviceQualityStat, error) {
	return database.GetActiveOrderQuality()
}

// ========================================================
// MES 设备管理接口
// ========================================================

// GetAllDevices 获取所有设备
func (a *App) GetAllDevices() ([]*models.SysDevice, error) {
	devices, err := database.GetAllDevices()
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// ========================================================
// MES 人员管理接口
// ========================================================

// GetAllStaff 获取所有员工
func (a *App) GetAllStaff(teamID *int, isActive *int8) ([]*models.SysStaff, error) {
	staffList, err := database.GetAllStaff(teamID, isActive)
	if err != nil {
		return nil, err
	}
	return staffList, nil
}

// CreateStaff 创建员工
func (a *App) CreateStaff(staffCode, name string, currentTeamID *int) (*models.SysStaff, error) {
	staff := &models.SysStaff{
		StaffCode:     staffCode,
		Name:          name,
		CurrentTeamID: currentTeamID,
		IsActive:      1, // 默认在职
	}

	if err := database.CreateStaff(staff); err != nil {
		return nil, err
	}

	// 重新加载包含班组信息
	staff, _ = database.GetStaffByID(staff.ID)
	return staff, nil
}

// UpdateStaff 更新员工
// 特殊值：currentTeamID = -1 表示清空班组（设为 NULL）
func (a *App) UpdateStaff(id int, name *string, currentTeamID *int, isActive *int8) error {
	updates := make(map[string]interface{})

	if name != nil {
		updates["name"] = *name
	}
	if currentTeamID != nil {
		if *currentTeamID == -1 {
			// -1 表示清空班组
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

// DeleteStaff 删除员工
func (a *App) DeleteStaff(id int) error {
	return database.DeleteStaff(id)
}

// TransferStaff 调动员工到新班组
func (a *App) TransferStaff(staffID, newTeamID int, operatorName *string) error {
	req := &models.TransferStaffRequest{
		StaffID:      staffID,
		NewTeamID:    newTeamID,
		OperatorName: operatorName,
	}
	return database.TransferStaff(req)
}

// GetStaffHistory 获取员工调动历史
func (a *App) GetStaffHistory(staffID int) ([]*models.SysStaffHistory, error) {
	return database.GetStaffHistory(staffID)
}

// ========================================================
// MES 班组管理接口
// ========================================================

// GetAllTeams 获取所有班组
func (a *App) GetAllTeams(status *int8) ([]*models.SysTeam, error) {
	teams, err := database.GetAllTeams(status)
	if err != nil {
		return nil, err
	}
	return teams, nil
}

// CreateTeam 创建班组
func (a *App) CreateTeam(teamName string, leaderName *string) (*models.SysTeam, error) {
	team := &models.SysTeam{
		TeamName:   teamName,
		LeaderName: leaderName,
		Status:     1, // 默认启用
	}

	if err := database.CreateTeam(team); err != nil {
		return nil, err
	}

	return team, nil
}

// UpdateTeam 更新班组
func (a *App) UpdateTeam(id int, teamName *string, leaderName *string, status *int8) error {
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

// DeleteTeam 删除班组
func (a *App) DeleteTeam(id int) error {
	return database.DeleteTeam(id)
}

// ========================================================
// MES 设备登录/班次管理接口
// ========================================================

// DeviceLogin 设备登录/上班打卡
func (a *App) DeviceLogin(deviceID, teamID int, staffIDs []int) (*models.ProMachineSession, error) {
	fmt.Printf("🔐 DeviceLogin 被调用: deviceID=%d, teamID=%d, staffIDs=%v\n", deviceID, teamID, staffIDs)

	// 序列化员工ID列表
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

	// 返回包含班组信息的记录
	result, err := database.GetSessionByID(session.ID)
	if err != nil {
		fmt.Printf("❌ 获取班次详情失败: %v\n", err)
		return nil, err
	}

	fmt.Printf("✅ 返回班次信息: %+v\n", result)
	return result, nil
}

// DeviceLogout 设备登出/下班打卡
func (a *App) DeviceLogout(deviceID int) (*models.ProMachineSession, error) {
	return database.DeviceLogout(deviceID)
}

// GetActiveSession 获取设备当前活动班次
func (a *App) GetActiveSession(deviceID int) (*models.ProMachineSession, error) {
	return database.GetActiveSession(deviceID)
}

// GetAllActiveSessions 获取所有设备的活动班次
func (a *App) GetAllActiveSessions() ([]*models.ProMachineSession, error) {
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

// GetSessionHistory 获取班次历史记录
func (a *App) GetSessionHistory(deviceID *int, teamID *int, startDate, endDate string) ([]*models.ProMachineSession, error) {
	var start, end *time.Time

	if startDate != "" {
		t, err := time.ParseInLocation("2006-01-02", startDate, time.Local)
		if err == nil {
			start = &t
		}
	}

	if endDate != "" {
		t, err := time.ParseInLocation("2006-01-02", endDate, time.Local)
		if err == nil {
			endTime := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			end = &endTime
		}
	}

	return database.GetSessionHistory(deviceID, teamID, start, end)
}

// GetSessionStats 获取班次统计信息
func (a *App) GetSessionStats(sessionID int64) (*models.SessionStatusResponse, error) {
	return database.GetSessionStats(sessionID)
}

// GetStaffAttendance 获取员工出勤记录
func (a *App) GetStaffAttendance(staffID int, startDate, endDate string) ([]*models.ProMachineSession, error) {
	var start, end *time.Time

	if startDate != "" {
		t, err := time.ParseInLocation("2006-01-02", startDate, time.Local)
		if err == nil {
			start = &t
		}
	}

	if endDate != "" {
		t, err := time.ParseInLocation("2006-01-02", endDate, time.Local)
		if err == nil {
			endTime := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			end = &endTime
		}
	}

	return database.GetStaffAttendance(staffID, start, end)
}

// ========================================================
// IOT 历史查询接口
// ========================================================

// TagInfo 变量信息
type TagInfo struct {
	VarID       int64  `json:"var_id"`
	VarName     string `json:"var_name"`
	DisplayName string `json:"display_name"`
	Unit        string `json:"unit"`
	StoreMode   int    `json:"store_mode"`
	DataType    string `json:"data_type"`
}

// GetAllTags 获取所有变量配置（用于历史查询）
func (a *App) GetAllTags() []TagInfo {
	tagManager := core.GetTagManager()
	allTags := tagManager.GetAllTags()

	result := make([]TagInfo, 0, len(allTags))
	for _, tag := range allTags {
		result = append(result, TagInfo{
			VarID:       tag.VarID,
			VarName:     tag.VarName,
			DisplayName: tag.DisplayName,
			Unit:        tag.Unit,
			StoreMode:   tag.StoreMode,
			DataType:    tag.DataType,
		})
	}

	return result
}

// HistoryRecord 历史数据记录
type HistoryRecord struct {
	Timestamp string      `json:"timestamp"`
	Value     interface{} `json:"value"`
}

// HistoryDataResponse 历史数据响应结构
type HistoryDataResponse struct {
	Records []HistoryRecord `json:"records"`
	Total   int64           `json:"total"`
}

// GetHistoryData 获取变量历史数据（分页）
func (a *App) GetHistoryData(varID int64, startTime, endTime string, page, pageSize int) (HistoryDataResponse, error) {
	// 默认值处理
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	// 构建基础查询
	baseQuery := database.DB.Table("sys_data_history").
		Where("var_id = ?", varID)

	if startTime != "" {
		baseQuery = baseQuery.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		baseQuery = baseQuery.Where("created_at <= ?", endTime)
	}

	// 查询总数
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return HistoryDataResponse{}, fmt.Errorf("查询历史数据总数失败: %v", err)
	}

	// 分页查询数据
	var results []struct {
		Val       *float64  `gorm:"column:val"`
		StrVal    *string   `gorm:"column:str_val"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	offset := (page - 1) * pageSize
	query := baseQuery.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset)

	if err := query.Find(&results).Error; err != nil {
		return HistoryDataResponse{}, fmt.Errorf("查询历史数据失败: %v", err)
	}

	// 格式化返回数据
	data := make([]HistoryRecord, len(results))
	for i, row := range results {
		record := HistoryRecord{
			Timestamp: row.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if row.StrVal != nil {
			record.Value = *row.StrVal
		} else if row.Val != nil {
			record.Value = *row.Val
		}
		data[i] = record
	}

	return HistoryDataResponse{
		Records: data,
		Total:   total,
	}, nil
}

// ========================================================
// IOT 配置管理接口
// ========================================================

// GetAllVariables 获取所有变量配置
func (a *App) GetAllVariables() ([]database.VariableRow, error) {
	var variables []database.VariableRow

	result := database.DB.Table("sys_variables").Order("id").Find(&variables)
	if result.Error != nil {
		return nil, fmt.Errorf("查询变量配置失败: %v", result.Error)
	}

	return variables, nil
}

// UpdateVariable 更新变量配置
func (a *App) UpdateVariable(variable database.VariableRow) error {
	result := database.DB.Table("sys_variables").
		Where("id = ?", variable.ID).
		Updates(map[string]interface{}{
			"device_id":      variable.DeviceID,
			"var_name":       variable.VarName,
			"display_name":   variable.DisplayName,
			"json_path":      variable.JSONPath,
			"data_type":      variable.DataType,
			"rw_mode":        variable.RWMode,
			"unit":           variable.Unit,
			"scale_factor":   variable.ScaleFactor,
			"offset_val":     variable.OffsetVal,
			"store_mode":     variable.StoreMode,
			"store_cycle":    variable.StoreCycle,
			"store_deadband": variable.StoreDeadband,
			"alarm_enable":   variable.AlarmEnable,
			"limit_hh":       variable.LimitHH,
			"limit_h":        variable.LimitH,
			"limit_l":        variable.LimitL,
			"limit_ll":       variable.LimitLL,
			"deadband":       variable.Deadband,
			"alarm_msg":      variable.AlarmMsg,
		})

	if result.Error != nil {
		return fmt.Errorf("更新变量配置失败: %v", result.Error)
	}

	// 更新配置版本以触发热重载
	newVersion := fmt.Sprintf("v%d", time.Now().Unix())
	if err := database.UpdateConfigVersion(newVersion); err != nil {
		return fmt.Errorf("触发热重载失败: %v", err)
	}

	return nil
}

// BatchUpdateVariables 批量更新变量配置
func (a *App) BatchUpdateVariables(variables []database.VariableRow) error {
	if len(variables) == 0 {
		return fmt.Errorf("没有要更新的变量")
	}

	// 使用事务批量更新
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, variable := range variables {
		result := tx.Table("sys_variables").
			Where("id = ?", variable.ID).
			Updates(map[string]interface{}{
				"device_id":      variable.DeviceID,
				"var_name":       variable.VarName,
				"display_name":   variable.DisplayName,
				"json_path":      variable.JSONPath,
				"data_type":      variable.DataType,
				"rw_mode":        variable.RWMode,
				"unit":           variable.Unit,
				"scale_factor":   variable.ScaleFactor,
				"offset_val":     variable.OffsetVal,
				"store_mode":     variable.StoreMode,
				"store_cycle":    variable.StoreCycle,
				"store_deadband": variable.StoreDeadband,
				"alarm_enable":   variable.AlarmEnable,
				"limit_hh":       variable.LimitHH,
				"limit_h":        variable.LimitH,
				"limit_l":        variable.LimitL,
				"limit_ll":       variable.LimitLL,
				"deadband":       variable.Deadband,
				"alarm_msg":      variable.AlarmMsg,
			})

		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("批量更新失败: %v", result.Error)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 更新配置版本以触发热重载
	newVersion := fmt.Sprintf("v%d", time.Now().Unix())
	if err := database.UpdateConfigVersion(newVersion); err != nil {
		return fmt.Errorf("触发热重载失败: %v", err)
	}

	return nil
}

// CreateVariable 创建新变量配置
func (a *App) CreateVariable(variable database.VariableRow) error {
	// 验证必填字段
	if variable.VarName == "" {
		return fmt.Errorf("变量名不能为空")
	}
	if variable.JSONPath == "" {
		return fmt.Errorf("JSON路径不能为空")
	}

	// 设置默认值（处理指针类型）
	if variable.DataType == nil || *variable.DataType == "" {
		defaultDataType := "FLOAT"
		variable.DataType = &defaultDataType
	}
	if variable.RWMode == nil || *variable.RWMode == "" {
		defaultRWMode := "R"
		variable.RWMode = &defaultRWMode
	}
	if variable.ScaleFactor == 0 {
		variable.ScaleFactor = 1.0
	}

	result := database.DB.Table("sys_variables").Create(&variable)
	if result.Error != nil {
		return fmt.Errorf("创建变量配置失败: %v", result.Error)
	}

	// 更新配置版本以触发热重载
	newVersion := fmt.Sprintf("v%d", time.Now().Unix())
	if err := database.UpdateConfigVersion(newVersion); err != nil {
		return fmt.Errorf("触发热重载失败: %v", err)
	}

	return nil
}

// DeleteVariable 删除变量配置
func (a *App) DeleteVariable(id int64) error {
	if id <= 0 {
		return fmt.Errorf("无效的变量ID")
	}

	// 检查是否存在
	var count int64
	database.DB.Table("sys_variables").Where("id = ?", id).Count(&count)
	if count == 0 {
		return fmt.Errorf("变量不存在")
	}

	// 删除变量
	result := database.DB.Table("sys_variables").Where("id = ?", id).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除变量配置失败: %v", result.Error)
	}

	// 更新配置版本以触发热重载
	newVersion := fmt.Sprintf("v%d", time.Now().Unix())
	if err := database.UpdateConfigVersion(newVersion); err != nil {
		return fmt.Errorf("触发热重载失败: %v", err)
	}

	return nil
}

// BatchDeleteVariables 批量删除变量配置
func (a *App) BatchDeleteVariables(ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("没有要删除的变量")
	}

	// 使用事务批量删除
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Table("sys_variables").Where("id IN ?", ids).Delete(nil)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("批量删除失败: %v", result.Error)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 更新配置版本以触发热重载
	newVersion := fmt.Sprintf("v%d", time.Now().Unix())
	if err := database.UpdateConfigVersion(newVersion); err != nil {
		return fmt.Errorf("触发热重载失败: %v", err)
	}

	return nil
}

// ========================================================
// 设备状态管理接口
// ========================================================

// DeviceStatusData 设备状态数据传输对象
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
	Operators     string     `json:"operators"`   // 操作人员姓名列表
	RecordTime    string     `json:"record_time"` // 记录时间
	Temperature   string     `json:"temperature"` // 温度
	Humidity      string     `json:"humidity"`    // 湿度
	Remark        string     `json:"remark"`      // 备注
}

// ExtraData 扩展数据结构
type ExtraData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Speed       int     `json:"speed"`
}

// GetAllDevicesStatus 获取所有设备状态（含操作人员）
func (a *App) GetAllDevicesStatus() ([]DeviceStatusData, error) {
	// 1. 获取所有设备状态统计
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

		// 2. 获取设备编码
		device, err := database.GetDeviceByID(summary.DeviceID)
		if err == nil {
			data.DeviceCode = device.DeviceCode
		}

		// 3. 获取当前班次的操作人员
		session, err := database.GetActiveSession(summary.DeviceID)
		if err == nil && session != nil {
			// 解析员工ID列表
			var staffIDs []int
			if err := json.Unmarshal([]byte(session.StaffIDs), &staffIDs); err == nil {
				// 获取员工姓名
				var staffNames []string
				for _, staffID := range staffIDs {
					staff, err := database.GetStaffByID(staffID)
					if err == nil {
						staffNames = append(staffNames, staff.Name)
					}
				}
				if len(staffNames) > 0 {
					data.Operators = fmt.Sprintf("%v", staffNames[0])
					if len(staffNames) > 1 {
						for i := 1; i < len(staffNames); i++ {
							data.Operators += "、" + staffNames[i]
						}
					}
				}
			}
		}

		// 如果没有操作人员，显示"-"
		if data.Operators == "" {
			data.Operators = "-"
		}

		// 格式化记录时间
		if data.StartTime != nil {
			data.RecordTime = data.StartTime.Format("2006-01-02 15:04")
		} else {
			data.RecordTime = "-"
		}

		// 4. 获取当前状态记录的扩展数据（温度、湿度）
		currentStatus, err := database.GetDeviceCurrentStatus(summary.DeviceID)
		if err == nil && currentStatus != nil {
			// 解析 extra_data JSON字段
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
			// 获取备注
			if currentStatus.Remark != nil {
				data.Remark = *currentStatus.Remark
			}
		}

		// 默认值
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

// GetDeviceStatusHistory 获取设备24小时状态历史（用于甘特图）
func (a *App) GetDeviceStatusHistory(deviceID int) ([]*models.SysDeviceStatus, error) {
	// 获取最近24小时的数据
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	history, err := database.GetDeviceStatusHistory(deviceID, &startTime, &endTime)
	if err != nil {
		return nil, fmt.Errorf("获取状态历史失败: %v", err)
	}

	return history, nil
}

// DeviceStatusHistoryData 设备状态历史数据（含班次信息）
type DeviceStatusHistoryData struct {
	models.SysDeviceStatus
	TeamName  string `json:"team_name"`
	Operators string `json:"operators"`
}

// GetDeviceStatusHistoryAll 获取所有设备状态历史（支持筛选，含班次信息）
func (a *App) GetDeviceStatusHistoryAll(deviceID *int, startTimeStr, endTimeStr string) ([]DeviceStatusHistoryData, error) {
	var startTime, endTime *time.Time

	// 解析时间参数（使用本地时区，与数据库的 loc=Local 保持一致）
	if startTimeStr != "" {
		// 尝试带秒的格式，如果失败则尝试不带秒的格式
		t, err := time.ParseInLocation("2006-01-02T15:04:05", startTimeStr, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02T15:04", startTimeStr, time.Local)
		}
		if err == nil {
			startTime = &t
		}
	}

	if endTimeStr != "" {
		// 尝试带秒的格式，如果失败则尝试不带秒的格式
		t, err := time.ParseInLocation("2006-01-02T15:04:05", endTimeStr, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02T15:04", endTimeStr, time.Local)
		}
		if err == nil {
			endTime = &t
		}
	}

	// 构建查询
	query := database.DB.Preload("Device").Order("start_time DESC")

	// 设备筛选
	if deviceID != nil {
		query = query.Where("device_id = ?", *deviceID)
	}

	// 时间范围筛选
	// 查询与时间范围有重叠的记录：
	// 1. 在范围内开始的记录
	// 2. 在范围前开始但在范围内结束的记录
	// 3. 在范围前开始且还未结束的记录（end_time IS NULL）
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

	var records []*models.SysDeviceStatus
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("查询历史记录失败: %v", err)
	}

	// 为每条记录查找对应时间的班次信息
	result := make([]DeviceStatusHistoryData, 0, len(records))
	for _, record := range records {
		data := DeviceStatusHistoryData{
			SysDeviceStatus: *record,
			TeamName:        "",
			Operators:       "",
		}

		// 查找该时间段内的班次（login_time <= record.start_time AND (logout_time IS NULL OR logout_time >= record.start_time)）
		var session models.ProMachineSession
		err := database.DB.Preload("Team").
			Where("device_id = ? AND login_time <= ? AND (logout_time IS NULL OR logout_time >= ?)",
				record.DeviceID, record.StartTime, record.StartTime).
			Order("login_time DESC").
			First(&session).Error

		if err == nil {
			// 找到对应的班次
			if session.Team != nil {
				data.TeamName = session.Team.TeamName
			}

			// 解析操作人员
			var staffIDs []int
			if err := json.Unmarshal([]byte(session.StaffIDs), &staffIDs); err == nil {
				var staffNames []string
				for _, staffID := range staffIDs {
					staff, err := database.GetStaffByID(staffID)
					if err == nil {
						staffNames = append(staffNames, staff.Name)
					}
				}
				if len(staffNames) > 0 {
					data.Operators = staffNames[0]
					for i := 1; i < len(staffNames); i++ {
						data.Operators += "、" + staffNames[i]
					}
				}
			}
		}

		result = append(result, data)
	}

	return result, nil
}

// GetDeviceStatusStats 获取设备状态统计汇总
func (a *App) GetDeviceStatusStats() (map[string]interface{}, error) {
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
		case 1: // 运行
			runningCount++
		case 0: // 空闲
			idleCount++
		case 2: // 故障
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

// ========================================================
// 统计数据接口 (用于驾驶舱)
// ========================================================

// GetHourlyProduction 获取今日按小时统计的产量
func (a *App) GetHourlyProduction(deviceID *int) ([]database.HourlyProduction, error) {
	return database.GetHourlyProduction(deviceID)
}

// GetStaffEfficiency 获取员工绩效统计
func (a *App) GetStaffEfficiency(startTime, endTime *time.Time) ([]database.StaffEfficiency, error) {
	return database.GetStaffEfficiency(startTime, endTime)
}

// GetDeviceUtilizationTrend 获取设备利用率趋势
func (a *App) GetDeviceUtilizationTrend(deviceID *int) ([]database.DeviceUtilizationTrend, error) {
	return database.GetDeviceUtilizationTrend(deviceID)
}

// ========================================================
// 驾驶舱多班次汇总接口（Cockpit Multi-Shift Summary）
// 大画面マルチシフト集計インターフェース
// ========================================================

// ShiftDeviceOEE 单班次单设备OEE汇总（驾驶舱多班对比用）
// Per-device OEE summary for one shift (cockpit multi-shift comparison).
// 1シフト・1設備のOEE集計（コックピット多シフト比較用）。
type ShiftDeviceOEE struct {
	DeviceName      string  `json:"device_name"`
	AvailabilityPct float64 `json:"availability_pct"`
	PerformancePct  float64 `json:"performance_pct"`
	QualityPct      float64 `json:"quality_pct"`
	OEEPct          float64 `json:"oee_pct"`
}

// ShiftOEESummary 单班次OEE汇总（驾驶舱多班对比用）
// OEE summary for one shift (cockpit multi-shift comparison).
// 1シフト分のOEE集計（コックピット多シフト比較用）。
type ShiftOEESummary struct {
	ShiftName  string           `json:"shift_name"`
	StartLabel string           `json:"start_label"` // "07:40"
	EndLabel   string           `json:"end_label"`   // "16:20"
	IsCurrent  bool             `json:"is_current"`
	HasArrived bool             `json:"has_arrived"`
	Devices    []ShiftDeviceOEE `json:"devices"`
}

// ShiftQualitySummary 单班次良品率汇总（驾驶舱多班对比用）
// Quality summary for one shift (cockpit multi-shift comparison).
// 1シフト分の良品率集計（コックピット多シフト比較用）。
type ShiftQualitySummary struct {
	ShiftName  string                       `json:"shift_name"`
	StartLabel string                       `json:"start_label"`
	EndLabel   string                       `json:"end_label"`
	IsCurrent  bool                         `json:"is_current"`
	HasArrived bool                         `json:"has_arrived"`
	Devices    []database.DeviceQualityStat `json:"devices"`
}

// GetAllShiftsOEESummary 返回当前逻辑日所有已到达班次的OEE汇总（含当班）
//
// CN: 遍历 GetShiftsForLogicalDay 中 has_arrived==true 的班次，
//
//	对每个班次复用 hardcodedOEEConfig + buildShiftOEEWindow 构建独立时间窗口，
//	调用 database.GetHourlyOEE 取数后提取各设备"合计"行构成 ShiftDeviceOEE。
//	前端可从数组中取 is_current 项作为大字，取其余最近2条作为小字历史对比。
//
// EN: Iterates arrived shifts for the logical day, computes per-shift OEE using the existing
//
//	hardcoded config + shift window builders, and returns one ShiftOEESummary per shift.
//
// JP: 論理日の到達済シフトを走査し、既存のコンフィグ+ウィンドウビルダでシフト単位OEEを計算して返す。
func (a *App) GetAllShiftsOEESummary() ([]ShiftOEESummary, error) {
	ldShifts, err := a.GetShiftsForLogicalDay()
	if err != nil {
		return nil, fmt.Errorf("获取逻辑日班次失败: %w", err)
	}

	// 全局 CT（各设备无单独配置时 fallback）
	ct, err := a.GetProductionCoefficient()
	if err != nil || ct <= 0 {
		ct = 100.0
	}

	// 读取所有设备（含 cycle_time）
	var allDevices []models.SysDevice
	if err := database.DB.Select("id, device_name, schedule_id, cycle_time").Find(&allDevices).Error; err != nil {
		allDevices = nil
	}

	result := make([]ShiftOEESummary, 0, len(ldShifts))
	for _, s := range ldShifts {
		if !s.HasArrived {
			continue
		}

		// 构建该班次的 OEE 配置
		var devCfgs []database.DeviceOEEConfig
		if len(allDevices) > 0 {
			for _, d := range allDevices {
				devCT := ct
				if d.CycleTime != nil && *d.CycleTime > 0 {
					devCT = *d.CycleTime
				}
				if cfg := hardcodedOEEConfig(d.ID, devCT); cfg != nil {
					devCfgs = append(devCfgs, *cfg)
				}
			}
		}
		if len(devCfgs) == 0 {
			devCfgs = []database.DeviceOEEConfig{
				hardcodedOEEConfigDefault(1, ct),
				hardcodedOEEConfigDefault(2, ct),
			}
		}

		window, breaks := buildShiftOEEWindow(s)
		rows, err := database.GetHourlyOEE(devCfgs, breaks, window)
		if err != nil {
			rows = nil
		}

		// 提取各设备"合计"行
		devices := make([]ShiftDeviceOEE, 0, 2)
		for _, row := range rows {
			if row.TimePeriod != "" && containsStr(row.TimePeriod, "合计") {
				devices = append(devices, ShiftDeviceOEE{
					DeviceName:      row.DeviceName,
					AvailabilityPct: row.Availability,
					PerformancePct:  row.Performance,
					QualityPct:      row.Quality,
					OEEPct:          row.OEE,
				})
			}
		}

		result = append(result, ShiftOEESummary{
			ShiftName:  s.Name,
			StartLabel: fmt.Sprintf("%02d:%02d", s.StartHour, s.StartMin),
			EndLabel:   fmt.Sprintf("%02d:%02d", s.EndHour, s.EndMin),
			IsCurrent:  s.IsCurrent,
			HasArrived: s.HasArrived,
			Devices:    devices,
		})
	}
	return result, nil
}

// GetAllShiftsQualitySummary 返回当前逻辑日所有已到达班次的良品率汇总（含当班）
//
// CN: 遍历 has_arrived==true 的班次，对每台有 OEE 配置的设备调用 GetShiftWindowProduction，
//     口径与班次快照（pro_shift_snapshots.quality_pct）完全一致，均来自 sys_data_history。
//     之前的实现调用 GetShiftQualityByRun（pro_production_runs 工单运行记录），
//     与快照数据源不同导致驾驶舱良品率和班组追溯页面数值对不上，现已修正。
//     跨零点班次：endMin <= startMin 时 shiftEnd 加 24h。
//
// EN: For each arrived shift, calls GetShiftWindowProduction per configured device.
//     Data source is sys_data_history, identical to snapshot quality_pct calculation.
//     Previous impl used pro_production_runs (GetShiftQualityByRun), causing mismatch with
//     shift report; now unified with snapshot source.
//
// JP: 到達済シフトごとに設備別 GetShiftWindowProduction を呼び出す。
//     データソースは sys_data_history で、スナップショットの quality_pct と完全に一致する。
//     以前は pro_production_runs を参照していたため班組追跡画面と乖離していた。修正済み。
func (a *App) GetAllShiftsQualitySummary() ([]ShiftQualitySummary, error) {
	ldShifts, err := a.GetShiftsForLogicalDay()
	if err != nil {
		return nil, fmt.Errorf("获取逻辑日班次失败: %w", err)
	}

	// 枚举所有设备中有 OEE 配置的设备，构建变量映射列表（CT 对良品率无影响，传 1 即可）
	var allDevices []models.SysDevice
	if err := database.DB.Find(&allDevices).Error; err != nil {
		return nil, fmt.Errorf("查询设备列表失败: %w", err)
	}
	type configuredDev struct {
		Dev models.SysDevice
		Cfg database.DeviceVarConfig
	}
	var confDevs []configuredDev
	for _, dev := range allDevices {
		oee := hardcodedOEEConfig(dev.ID, 1)
		if oee == nil {
			continue
		}
		confDevs = append(confDevs, configuredDev{
			Dev: dev,
			Cfg: database.DeviceVarConfig{
				DeviceName:      dev.DeviceName,
				ProductionVarID: oee.VarOK,
				NgAddVarID:      oee.VarNGAdd,
				NgSubVarID:      oee.VarNGSub,
			},
		})
	}

	result := make([]ShiftQualitySummary, 0, len(ldShifts))
	for _, s := range ldShifts {
		if !s.HasArrived {
			continue
		}

		// 解析逻辑日基准时间
		logicalBase, err := time.ParseInLocation("2006-01-02", s.LogicalDate, time.Local)
		if err != nil {
			continue
		}
		startMin := s.StartHour*60 + s.StartMin
		endMin := s.EndHour*60 + s.EndMin
		shiftStart := logicalBase.Add(time.Duration(startMin) * time.Minute)
		var shiftEnd time.Time
		if endMin > startMin {
			shiftEnd = logicalBase.Add(time.Duration(endMin) * time.Minute)
		} else {
			shiftEnd = logicalBase.Add(time.Duration(endMin+24*60) * time.Minute)
		}

		// 对每台配置设备调用 sys_data_history 窗口聚合，口径与快照完全一致
		devStats := make([]database.DeviceQualityStat, 0, len(confDevs))
		for _, cd := range confDevs {
			pr, err := database.GetShiftWindowProduction(cd.Cfg, shiftStart, shiftEnd)
			if err != nil {
				pr = database.ShiftWindowProdResult{}
			}
			qRate := 100.0
			if pr.TotalQty > 0 {
				qRate = math.Round(float64(pr.OkQty)*10000.0/float64(pr.TotalQty)) / 100
			}
			devStats = append(devStats, database.DeviceQualityStat{
				DeviceID:    cd.Dev.ID,
				DeviceName:  cd.Dev.DeviceName,
				TotalQty:    pr.TotalQty,
				OkQty:       pr.OkQty,
				NgQty:       pr.NgQty,
				QualityRate: qRate,
			})
		}

		result = append(result, ShiftQualitySummary{
			ShiftName:  s.Name,
			StartLabel: fmt.Sprintf("%02d:%02d", s.StartHour, s.StartMin),
			EndLabel:   fmt.Sprintf("%02d:%02d", s.EndHour, s.EndMin),
			IsCurrent:  s.IsCurrent,
			HasArrived: s.HasArrived,
			Devices:    devStats,
		})
	}
	return result, nil
}

// hardcodedOEEConfigDefault 返回 hardcodedOEEConfig 的非指针安全版本（兜底用）
func hardcodedOEEConfigDefault(deviceID int, ct float64) database.DeviceOEEConfig {
	if cfg := hardcodedOEEConfig(deviceID, ct); cfg != nil {
		return *cfg
	}
	return database.DeviceOEEConfig{DeviceID: deviceID, CycleTime: ct}
}

// containsStr 简单子串检查（替代 strings.Contains 避免额外 import）
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}

// getOEEConfigs 读取OEE所需的配置（设备配置和休息时间），供内部复用
func (a *App) getOEEConfigs() ([]database.DeviceOEEConfig, []database.BreakTimeConfig, error) {
	// 读取休息时间配置
	breakTimes, err := a.GetBreakTimes()
	if err != nil {
		fmt.Printf("⚠️ 读取休息时间配置失败，使用默认配置: %v\n", err)
		breakTimes = getDefaultBreakTimes()
	}
	dbBreakTimes := make([]database.BreakTimeConfig, len(breakTimes))
	for i, bt := range breakTimes {
		dbBreakTimes[i] = database.BreakTimeConfig{
			Name:      bt.Name,
			StartHour: bt.StartHour,
			StartMin:  bt.StartMin,
			EndHour:   bt.EndHour,
			EndMin:    bt.EndMin,
		}
	}

	// 读取理论节拍（CT），用于计算性能稼动率
	ct, err := a.GetProductionCoefficient()
	if err != nil || ct <= 0 {
		ct = 100.0
	}
	deviceConfigs := []database.DeviceOEEConfig{
		{DeviceID: 1, DeviceName: "设备#1", VarOK: 1, VarNGAdd: 72, VarNGSub: 71, CycleTime: ct},
		{DeviceID: 2, DeviceName: "设备#2", VarOK: 95, VarNGAdd: 97, VarNGSub: 96, CycleTime: ct},
	}
	return deviceConfigs, dbBreakTimes, nil
}

// hardcodedOEEConfig 返回指定设备 ID 的 OEE 变量配置（当前为固定值，后续可改为 DB 配置）
// Returns the hardcoded OEE variable mapping for a device ID.
// デバイス ID に対応するハードコードされた OEE 変数マッピングを返す。
func hardcodedOEEConfig(deviceID int, ct float64) *database.DeviceOEEConfig {
	switch deviceID {
	case 1:
		return &database.DeviceOEEConfig{DeviceID: 1, DeviceName: "设备#1", VarOK: 1, VarNGAdd: 72, VarNGSub: 71, CycleTime: ct}
	case 2:
		return &database.DeviceOEEConfig{DeviceID: 2, DeviceName: "设备#2", VarOK: 95, VarNGAdd: 97, VarNGSub: 96, CycleTime: ct}
	}
	return nil
}

// buildScheduleOEEWindow 根据一组逻辑日班次（同属一个时间安排组）构建宽时间窗口 + 合并休息段
//
// CN: 宽窗口 = 首班开始 → 末班结束（跨日用 24+ 小时表示）。
//
//	班次内休息段直接沿用 buildShiftOEEWindow 的跨日修正。
//	相邻班次间的间隔也作为休息段插入，防止计划时间虚高。
//
// EN: Wide window = first-shift-start → last-shift-end (24+ hour for cross-midnight).
//
//	Intra-shift breaks via buildShiftOEEWindow; inter-shift gaps also added as breaks.
//
// JP: 広ウィンドウ = 最初シフト開始 → 最後シフト終了（深夜跨ぎは 24+ 時間表記）。
//
//	シフト内休憩は buildShiftOEEWindow で処理し、シフト間ギャップも休憩として追加。
func buildScheduleOEEWindow(shifts []LogicalDayShift, logicalDate string) (*database.ShiftWindow, []database.BreakTimeConfig) {
	if len(shifts) == 0 {
		return nil, nil
	}
	if len(shifts) == 1 {
		return buildShiftOEEWindow(shifts[0])
	}

	firstShift := shifts[0]
	lastShift := shifts[len(shifts)-1]

	workStartMin := firstShift.StartHour*60 + firstShift.StartMin

	lastEndMin := lastShift.EndHour*60 + lastShift.EndMin
	lastStartMin := lastShift.StartHour*60 + lastShift.StartMin
	if lastEndMin < lastStartMin {
		lastEndMin += 24 * 60 // 末班跨零点
	}

	lastEndH := lastEndMin / 60
	lastEndM := lastEndMin % 60
	hourEndLimit := lastEndH
	if lastEndM > 0 {
		hourEndLimit++
	}

	window := &database.ShiftWindow{
		WorkStart:   fmt.Sprintf("%02d:%02d", workStartMin/60, workStartMin%60),
		WorkEnd:     fmt.Sprintf("%02d:%02d", lastEndMin/60, lastEndMin%60),
		LogicalDate: logicalDate,
		HourStart:   workStartMin / 60,
		HourEnd:     hourEndLimit,
	}

	allBreaks := make([]database.BreakTimeConfig, 0)
	prevSpanEnd := -1
	for _, s := range shifts {
		_, bkCfgs := buildShiftOEEWindow(s)
		sStartMin := s.StartHour*60 + s.StartMin
		sEndRaw := s.EndHour*60 + s.EndMin
		sSpanEnd := sEndRaw
		if sEndRaw < sStartMin {
			sSpanEnd = sEndRaw + 24*60
		}
		if prevSpanEnd > 0 && sStartMin > prevSpanEnd {
			allBreaks = append(allBreaks, database.BreakTimeConfig{
				Name: "班间间隔", StartHour: prevSpanEnd / 60, StartMin: prevSpanEnd % 60,
				EndHour: sStartMin / 60, EndMin: sStartMin % 60,
			})
		}
		prevSpanEnd = sSpanEnd
		allBreaks = append(allBreaks, bkCfgs...)
	}
	return window, allBreaks
}

// buildShiftOEEWindow 根据单个逻辑日班次构建 ShiftWindow 和休息段列表
//
// CN: 跨零点班次（EndHour < StartHour，如 22:00→06:00）处理规则：
//   - WorkEnd 使用 "30:00" 形式（EndHour+24），MySQL ADDTIME 支持 >24h 的时间偏移
//   - HourEnd 同步上调（24 + EndHour），使 Hours CTE 能生成 24/25/26…等桶
//   - 落在次日（bStartMin < shiftStartMin）的休息段小时同样加 24
//
// EN: For cross-midnight shifts, WorkEnd / HourEnd / break hours are offset by +24
//
//	so that MySQL ADDTIME and Hours CTE correctly point to the next calendar day.
//
// JP: 深夜跨ぎシフトは WorkEnd/HourEnd/休憩時間帯に +24 を加算し、
//
//	MySQL の ADDTIME と Hours CTE が正しく翌日を指すようにする。
func buildShiftOEEWindow(s LogicalDayShift) (*database.ShiftWindow, []database.BreakTimeConfig) {
	startMinTotal := s.StartHour*60 + s.StartMin
	endMinTotal := s.EndHour*60 + s.EndMin

	// 跨零点判定：结束分钟数 < 开始分钟数（例如 22:00→06:00，360 < 1320）
	crossMidnight := endMinTotal < startMinTotal

	// 跨日时，WorkEnd 用 "30:00" 等超 24 格式，使 ADDTIME 指向次日
	endHourSQL := s.EndHour
	if crossMidnight {
		endHourSQL = s.EndHour + 24
	}

	// HourEnd = 最后时间桶 hour_idx 的上限（WHERE hour_idx < HourEnd）
	hourEndLimit := endHourSQL
	if s.EndMin > 0 {
		hourEndLimit = endHourSQL + 1
	}
	if hourEndLimit <= s.StartHour {
		hourEndLimit = s.StartHour + 1
	}

	window := &database.ShiftWindow{
		WorkStart:   fmt.Sprintf("%02d:%02d", s.StartHour, s.StartMin),
		WorkEnd:     fmt.Sprintf("%02d:%02d", endHourSQL, s.EndMin),
		LogicalDate: s.LogicalDate,
		HourStart:   s.StartHour,
		HourEnd:     hourEndLimit,
	}

	breaks := make([]database.BreakTimeConfig, 0, len(s.Breaks))
	for _, b := range s.Breaks {
		bStartMin := b.StartHour*60 + b.StartMin
		bEndMin := b.EndHour*60 + b.EndMin
		startH, endH := b.StartHour, b.EndHour

		if crossMidnight {
			// 休息段开始落在次日（时钟时间 < 班次开始时间）
			if bStartMin < startMinTotal {
				startH += 24
			}
			// 休息段结束落在次日：
			//   (a) 结束时间 < 班次开始时间（次日早上）
			//   (b) 休息段本身跨零点（结束时间 < 开始时间，如 23:30→00:30）
			// CN: 两种情况分别判断，避免"开始在今日深夜、结束在次日凌晨"的休息段漏加偏移
			// EN: Check (a) and (b) separately so cross-midnight breaks (e.g., 23:30→00:30) get correct endH.
			// JP: (a)(b) を個別に判定し、深夜跨ぎ休憩（23:30→00:30 等）の endH が正しくオフセットされるようにする。
			if bEndMin < startMinTotal || bEndMin < bStartMin {
				endH += 24
			}
		}
		breaks = append(breaks, database.BreakTimeConfig{
			Name: b.Name, StartHour: startH, StartMin: b.StartMin,
			EndHour: endH, EndMin: b.EndMin,
		})
	}
	return window, breaks
}

// GetHourlyOEE 获取当前逻辑日每小时 OEE 统计
//
// CN: 数据流：
//  1. GetShiftsForLogicalDay → 逻辑日期 + 各活动班次（含休息段）
//  2. 从 DB 读取设备的 shift_id 关联
//  3. 按 shift_id 将设备分组
//  4. 每组单独调用 database.GetHourlyOEE（使用各自班次的时间窗口 + 休息段）
//  5. 若设备无 shift_id，回退到合并所有班次的宽窗口查询
//
// EN: Groups devices by their assigned shift, runs per-shift OEE queries, merges results.
// JP: 設備をシフト別にグループ化し、シフトごとに OEE クエリを実行して結果をマージする。
func (a *App) GetHourlyOEE(configs []database.DeviceOEEConfig) ([]database.HourlyOEE, error) {
	ct, err := a.GetProductionCoefficient()
	if err != nil || ct <= 0 {
		ct = 100.0
	}

	// 1. 获取逻辑日班次列表
	ldShifts, err := a.GetShiftsForLogicalDay()
	if err != nil || len(ldShifts) == 0 {
		fmt.Printf("⚠️ 无活动班次，OEE使用默认配置: %v\n", err)
		return database.GetHourlyOEE(nil, nil, nil) // 全部用 GetHourlyOEE 内置默认值
	}

	// 2. 读取 DB 中所有设备（含 schedule_id 和 cycle_time）
	var allDevices []models.SysDevice
	if err := database.DB.Select("id, device_name, schedule_id, cycle_time").Find(&allDevices).Error; err != nil {
		fmt.Printf("⚠️ 读取设备列表失败，OEE使用默认配置: %v\n", err)
		return database.GetHourlyOEE(nil, nil, nil)
	}

	// 3. 构建 scheduleID → []DeviceOEEConfig 映射
	// CN: 设备按所属时间安排组（ScheduleID）分组，同组的设备共享同一套班次时间窗口。
	//     CT 优先使用设备级别配置（SysDevice.CycleTime），NULL 时 fallback 到全局默认值。
	// EN: Group devices by their schedule (ScheduleID); devices in the same schedule share the same shift windows.
	//     CT prefers device-level config (SysDevice.CycleTime), falls back to global default if NULL.
	// JP: 設備をスケジュール（ScheduleID）でグループ化。同グループの設備は同一シフトウィンドウを共有する。
	//     CT はデバイス単位設定を優先し、NULL の場合はグローバルデフォルトにフォールバック。
	scheduleDeviceMap := map[int][]database.DeviceOEEConfig{}
	unassigned := []database.DeviceOEEConfig{}

	for _, d := range allDevices {
		devCT := ct
		if d.CycleTime != nil && *d.CycleTime > 0 {
			devCT = *d.CycleTime
		}
		cfg := hardcodedOEEConfig(d.ID, devCT)
		if cfg == nil {
			continue
		}
		if d.ScheduleID != nil && *d.ScheduleID > 0 {
			scheduleDeviceMap[*d.ScheduleID] = append(scheduleDeviceMap[*d.ScheduleID], *cfg)
		} else {
			unassigned = append(unassigned, *cfg)
		}
	}

	allResults := make([]database.HourlyOEE, 0, 64)
	logicalDate := ldShifts[0].LogicalDate

	// 4. 按时间安排组（ScheduleID）计算 OEE
	// CN: 对每个 ScheduleID，取该组内所有逻辑日班次，以整组窗口（首班开始→末班结束）调用 OEE 查询。
	//     组内班次间的间隔作为 Break 处理，确保非工作时段不计入计划时间。
	// EN: For each ScheduleID, collect its logical-day shifts, run OEE over the full group window,
	//     treating inter-shift gaps as break periods.
	// JP: ScheduleID ごとに、論理日内の全シフトをまとめ、グループ全体のウィンドウで OEE クエリを実行。
	//     シフト間のギャップは休憩として扱い、計画時間に含めない。

	// scheduleID → その組に属する ldShifts
	scheduleShiftMap := map[int][]LogicalDayShift{}
	for _, s := range ldShifts {
		scheduleShiftMap[s.ScheduleID] = append(scheduleShiftMap[s.ScheduleID], s)
	}

	for schedID, devCfgs := range scheduleDeviceMap {
		if len(devCfgs) == 0 {
			continue
		}
		groupShifts := scheduleShiftMap[schedID]
		if len(groupShifts) == 0 {
			// 该 schedule 当前逻辑日没有活动班次，用所有 ldShifts 兜底
			groupShifts = ldShifts
		}

		// 构建该组的宽窗口 + 合并休息段
		window, allBreaks := buildScheduleOEEWindow(groupShifts, logicalDate)
		rows, err := database.GetHourlyOEE(devCfgs, allBreaks, window)
		if err != nil {
			fmt.Printf("⚠️ 时间安排组(id=%d) OEE查询失败: %v\n", schedID, err)
			continue
		}
		allResults = append(allResults, rows...)
	}

	// 5. 未绑定时间安排组的设备：用全局宽窗口（所有活动班次）兜底
	if len(unassigned) > 0 {
		window, allBreaks := buildScheduleOEEWindow(ldShifts, logicalDate)
		rows, err := database.GetHourlyOEE(unassigned, allBreaks, window)
		if err == nil {
			allResults = append(allResults, rows...)
		}
	}

	if len(allResults) == 0 {
		return database.GetHourlyOEE(nil, nil, nil)
	}
	return allResults, nil
}

// DebugOEEDirect 直接执行OEE SQL并返回结果（使用逻辑日窗口，与GetHourlyOEE保持一致）
//
// CN: 先调 GetShiftsForLogicalDay 确定逻辑日期和班次时间窗口，而非自然日 CURDATE()。
//
//	无活动班次时兜底使用 CURDATE()（window=nil）。
//
// EN: Uses logical day (GetShiftsForLogicalDay) instead of calendar CURDATE(), matching GetHourlyOEE.
//
//	Falls back to CURDATE() when no active shifts are found.
//
// JP: GetShiftsForLogicalDay で論理日ウィンドウを決定し、CURDATE() 自然日は使わない。
//
//	アクティブシフトがない場合は CURDATE() にフォールバック。
func (a *App) DebugOEEDirect() (map[string]interface{}, error) {
	dbDeviceConfigs, _, err := a.getOEEConfigs()
	if err != nil {
		return nil, fmt.Errorf("DebugOEEDirect读取配置失败: %w", err)
	}

	// 构建逻辑日窗口（与 GetHourlyOEE 相同路径）
	var window *database.ShiftWindow
	var allBreaks []database.BreakTimeConfig
	ldShifts, shiftErr := a.GetShiftsForLogicalDay()
	if shiftErr != nil || len(ldShifts) == 0 {
		fmt.Printf("⚠️ DebugOEEDirect: 无活动班次，使用默认窗口(CURDATE): %v\n", shiftErr)
		// window = nil → GetHourlyOEEWithSQL 内部使用 CURDATE() 兜底
	} else {
		window, allBreaks = buildScheduleOEEWindow(ldShifts, ldShifts[0].LogicalDate)
	}

	rows, query, err := database.GetHourlyOEEWithSQL(dbDeviceConfigs, allBreaks, window)
	if err != nil {
		return nil, fmt.Errorf("DebugOEEDirect失败: %w", err)
	}
	result := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		result[i] = map[string]interface{}{
			"time_period":      r.TimePeriod,
			"device_name":      r.DeviceName,
			"t_run":            r.TotalRunSec,
			"t_plan":           r.TotalPlanSec,
			"total_products":   r.TotalProducts,
			"ok_qty":           r.OKQty,
			"ng_qty":           r.NGQty,
			"availability_pct": r.Availability,
			"performance_pct":  r.Performance,
			"quality_pct":      r.Quality,
			"oee_pct":          r.OEE,
		}
	}
	return map[string]interface{}{
		"sql":  query,
		"rows": result,
	}, nil
}

// DebugOEEByShift 按班次分组返回逻辑日 OEE 数据，每班单独查询并附带班次元信息。
//
// CN: 对每个活动班次调用 buildShiftOEEWindow + GetHourlyOEEWithSQL，保留各班次自己的时间窗口和休息段。
//
//	结果按班次顺序（sort_order）排列，前端可直接按组渲染。
//
// EN: Queries OEE for each active shift independently using its own time window and break configuration.
//
//	Results are ordered by shift sort_order for direct per-group rendering on the frontend.
//
// JP: 各アクティブシフトを独立したウィンドウで OEE クエリし、sort_order 順に返す。フロントがグループ単位で描画できる。
func (a *App) DebugOEEByShift() ([]map[string]interface{}, error) {
	dbDeviceConfigs, _, err := a.getOEEConfigs()
	if err != nil {
		return nil, fmt.Errorf("DebugOEEByShift读取配置失败: %w", err)
	}

	ldShifts, shiftErr := a.GetShiftsForLogicalDay()
	if shiftErr != nil || len(ldShifts) == 0 {
		// 无活动班次：降级为单次全天查询，包装成一个伪班次组返回
		rows, _, err2 := database.GetHourlyOEEWithSQL(dbDeviceConfigs, nil, nil, true)
		if err2 != nil {
			return nil, fmt.Errorf("DebugOEEByShift降级查询失败: %w", err2)
		}
		return []map[string]interface{}{{
			"shift_id":    0,
			"shift_name":  "全天（无班次）",
			"shift_range": "",
			"rows":        oeeDebugRowsToMaps(rows),
		}}, nil
	}

	result := make([]map[string]interface{}, 0, len(ldShifts))
	for _, s := range ldShifts {
		window, breaks := buildShiftOEEWindow(s)
		rows, _, qErr := database.GetHourlyOEEWithSQL(dbDeviceConfigs, breaks, window, true)
		if qErr != nil {
			return nil, fmt.Errorf("班次%q OEE查询失败: %w", s.Name, qErr)
		}
		rangeStr := fmt.Sprintf("%02d:%02d — %02d:%02d", s.StartHour, s.StartMin, s.EndHour, s.EndMin)
		result = append(result, map[string]interface{}{
			"shift_id":     s.ID,
			"shift_name":   s.Name,
			"shift_range":  rangeStr,
			"logical_date": s.LogicalDate,
			"has_arrived":  s.HasArrived,
			"is_current":   s.IsCurrent,
			"rows":         oeeDebugRowsToMaps(rows),
		})
	}
	return result, nil
}

// oeeDebugRowsToMaps 将 HourlyOEEDebug 切片转换为前端 map 格式
// Converts HourlyOEEDebug slice to frontend-friendly map slice.
// HourlyOEEDebug スライスをフロントエンド向け map スライスに変換する。
func oeeDebugRowsToMaps(rows []database.HourlyOEEDebug) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rows))
	for i, r := range rows {
		result[i] = map[string]interface{}{
			"time_period":      r.TimePeriod,
			"device_name":      r.DeviceName,
			"t_run":            r.TotalRunSec,
			"t_plan":           r.TotalPlanSec,
			"total_products":   r.TotalProducts,
			"ok_qty":           r.OKQty,
			"ng_qty":           r.NGQty,
			"availability_pct": r.Availability,
			"performance_pct":  r.Performance,
			"quality_pct":      r.Quality,
			"oee_pct":          r.OEE,
		}
	}
	return result
}

// DebugOEEProductionRaw 直接查询ProductionRaw中间层，用于调试产量计算
func (a *App) DebugOEEProductionRaw() ([]map[string]interface{}, error) {
	query := `
WITH Config AS (SELECT CURDATE() as target_date),
DeviceConfig AS (
    SELECT 1 as device_id, 1 as var_ok, 72 as var_ng_add, 71 as var_ng_sub
    UNION ALL
    SELECT 2, 95, 97, 96
),
ProductionRaw AS (
    SELECT 
        d.created_at, d.val, d.var_id, dc.device_id, dc.var_ok,
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
            WHERE var_id IN (1, 95)
              AND created_at < ADDTIME(CURDATE(), '07:00:00')
            GROUP BY var_id
        )
    )
)
SELECT 
    HOUR(created_at) as hour_idx,
    created_at,
    var_id,
    val,
    prev_val,
    CASE WHEN prev_val IS NULL THEN val WHEN val >= prev_val THEN val - prev_val ELSE val END as delta
FROM ProductionRaw
WHERE var_id IN (1, 95)
ORDER BY var_id, created_at`

	var results []map[string]interface{}
	err := database.DB.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("DebugOEEProductionRaw失败: %w", err)
	}
	return results, nil
}

// ========================================================
// 报警管理接口
// ========================================================

// AlarmRecordData 报警记录数据
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
	Duration   string     `json:"duration"` // 持续时长
}

// GetAlarmRecords 获取报警记录（支持筛选）
// 🔥 改造：联表查询报警表+错误码表，组合显示
func (a *App) GetAlarmRecords(ackStatus *int, startTimeStr, endTimeStr string, varID *int64) ([]AlarmRecordData, error) {
	var startTime, endTime *time.Time

	// 解析时间参数（使用本地时区）
	if startTimeStr != "" {
		// 尝试带秒的格式，如果失败则尝试不带秒的格式
		t, err := time.ParseInLocation("2006-01-02T15:04:05", startTimeStr, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02T15:04", startTimeStr, time.Local)
		}
		if err == nil {
			startTime = &t
		}
	}

	if endTimeStr != "" {
		// 尝试带秒的格式，如果失败则尝试不带秒的格式
		t, err := time.ParseInLocation("2006-01-02T15:04:05", endTimeStr, time.Local)
		if err != nil {
			t, err = time.ParseInLocation("2006-01-02T15:04", endTimeStr, time.Local)
		}
		if err == nil {
			endTime = &t
		}
	}

	// 🔥 构建联表查询 - LEFT JOIN 错误码表
	// 对于系统报警(alarm_type='SYS')，通过 val 字段关联错误码表
	query := database.DB.Table("sys_alarm_records AS a").
		Select(`a.id, a.var_id, a.var_name, a.val, a.alarm_type, a.limit_value, 
		        CASE 
		            WHEN a.alarm_type = 'SYS' THEN e.error_msg 
		            ELSE a.msg 
		        END as msg,
		        a.start_time, a.end_time, a.ack_status`).
		Joins("LEFT JOIN sys_error_codes AS e ON a.alarm_type = 'SYS' AND CAST(a.val AS SIGNED) = e.error_code").
		Order("a.id DESC")

	// 确认状态筛选
	if ackStatus != nil {
		query = query.Where("a.ack_status = ?", *ackStatus)
	}

	// 变量ID筛选
	if varID != nil && *varID > 0 {
		query = query.Where("a.var_id = ?", *varID)
	}

	// 时间范围筛选
	if startTime != nil {
		query = query.Where("a.start_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("a.start_time <= ?", *endTime)
	}

	// 🔥 使用自定义结构接收联表查询结果
	type AlarmRecordWithErrorMsg struct {
		ID         int64      `gorm:"column:id"`
		VarID      int64      `gorm:"column:var_id"`
		VarName    *string    `gorm:"column:var_name"`
		Val        *float64   `gorm:"column:val"`
		AlarmType  string     `gorm:"column:alarm_type"`
		LimitValue *float64   `gorm:"column:limit_value"`
		Msg        *string    `gorm:"column:msg"` // 🔥 联表后的错误信息
		StartTime  time.Time  `gorm:"column:start_time"`
		EndTime    *time.Time `gorm:"column:end_time"`
		AckStatus  int        `gorm:"column:ack_status"`
	}

	var records []AlarmRecordWithErrorMsg
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("查询报警记录失败: %v", err)
	}

	// 转换为前端数据格式
	result := make([]AlarmRecordData, 0, len(records))
	for _, record := range records {
		data := AlarmRecordData{
			ID:         record.ID,
			VarID:      record.VarID,
			VarName:    "",
			Val:        0,
			AlarmType:  record.AlarmType,
			LimitValue: 0,
			Msg:        "",
			StartTime:  record.StartTime,
			EndTime:    record.EndTime,
			AckStatus:  record.AckStatus,
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

		// 计算持续时长
		if record.EndTime != nil {
			duration := record.EndTime.Sub(record.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			if hours > 0 {
				data.Duration = fmt.Sprintf("%dh%dm", hours, minutes)
			} else {
				data.Duration = fmt.Sprintf("%d分", minutes)
			}
		} else {
			// 未恢复，计算到现在的时长
			duration := time.Since(record.StartTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60
			if hours > 0 {
				data.Duration = fmt.Sprintf("%dh%dm (报警中)", hours, minutes)
			} else {
				data.Duration = fmt.Sprintf("%d分 (报警中)", minutes)
			}
		}

		result = append(result, data)
	}

	return result, nil
}

// AckAlarm 确认报警
func (a *App) AckAlarm(alarmID int64) error {
	result := database.DB.Table("sys_alarm_records").
		Where("id = ?", alarmID).
		Update("ack_status", 1)

	if result.Error != nil {
		return fmt.Errorf("确认报警失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("报警记录不存在")
	}

	return nil
}

// GetTodayUnacknowledgedAlarmCount 获取今日未确认的报警数（ack_status=0）
func (a *App) GetTodayUnacknowledgedAlarmCount() (int64, error) {
	// 获取今天的开始时间（凌晨0点）
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var count int64
	err := database.DB.Table("sys_alarm_records").
		Where("ack_status = ?", 0).           // 未确认的报警
		Where("start_time >= ?", todayStart). // 今天的报警
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("查询今日未确认报警数失败: %v", err)
	}

	return count, nil
}

// HourlyAlarmCount 每小时报警数统计
type HourlyAlarmCount struct {
	Hour       int    `json:"hour"`        // 小时 (0-23)
	AlarmCount int64  `json:"alarm_count"` // 报警数量
	TimeSlot   string `json:"time_slot"`   // 时间段标签，如 "7:00"
}

// GetTodayHourlyAlarmCount 获取今日每小时的报警数（用于迷你图）
func (a *App) GetTodayHourlyAlarmCount() ([]HourlyAlarmCount, error) {
	// 获取今天的开始和结束时间
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24 * time.Hour)

	// 查询今日报警记录，按小时分组统计
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

	// 转换为前端格式
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

// GetActiveAlarmCount 获取当前进行中工单相关的报警数（保留旧接口以兼容）
func (a *App) GetActiveAlarmCount() (int64, error) {
	// 1. 查询所有进行中的工单（status=1）
	var activeOrders []models.ProOrder
	err := database.DB.Where("status = ?", 1).Find(&activeOrders).Error
	if err != nil {
		return 0, fmt.Errorf("查询活动工单失败: %v", err)
	}

	// 如果没有进行中的工单，返回0
	if len(activeOrders) == 0 {
		return 0, nil
	}

	// 2. 提取设备ID列表
	deviceIDs := make([]int, 0)
	for _, order := range activeOrders {
		if order.TargetDeviceID != nil {
			deviceIDs = append(deviceIDs, *order.TargetDeviceID)
		}
	}

	if len(deviceIDs) == 0 {
		return 0, nil
	}

	// 3. 查询这些设备对应的变量ID
	var varIDs []int64
	err = database.DB.Table("sys_variables").
		Where("device_id IN ?", deviceIDs).
		Pluck("id", &varIDs).Error
	if err != nil {
		return 0, fmt.Errorf("查询设备变量失败: %v", err)
	}

	if len(varIDs) == 0 {
		return 0, nil
	}

	// 4. 统计这些变量的未恢复报警数（在工单开始时间之后触发的）
	// 获取最早的工单开始时间
	var earliestStartTime *time.Time
	for _, order := range activeOrders {
		if order.StartTime != nil {
			if earliestStartTime == nil || order.StartTime.Before(*earliestStartTime) {
				earliestStartTime = order.StartTime
			}
		}
	}

	// 如果没有工单开始时间，使用今日0点
	if earliestStartTime == nil {
		todayStart := time.Now().Truncate(24 * time.Hour)
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

// GetAlarmStats 获取报警统计
func (a *App) GetAlarmStats() (map[string]interface{}, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.Add(-24 * time.Hour)

	// 今日报警总数
	var todayTotal int64
	database.DB.Table("sys_alarm_records").
		Where("start_time >= ?", todayStart).
		Count(&todayTotal)

	// 今日已确认
	var todayAcked int64
	database.DB.Table("sys_alarm_records").
		Where("start_time >= ? AND ack_status = 1", todayStart).
		Count(&todayAcked)

	// 今日待确认
	var todayPending int64
	database.DB.Table("sys_alarm_records").
		Where("start_time >= ? AND ack_status = 0", todayStart).
		Count(&todayPending)

	// 昨日报警总数
	var yesterdayTotal int64
	database.DB.Table("sys_alarm_records").
		Where("start_time >= ? AND start_time < ?", yesterdayStart, todayStart).
		Count(&yesterdayTotal)

	// 计算同比
	var comparison string
	var comparisonValue int64
	if yesterdayTotal == 0 {
		if todayTotal > 0 {
			comparison = "up"
			comparisonValue = todayTotal
		} else {
			comparison = "equal"
			comparisonValue = 0
		}
	} else {
		diff := todayTotal - yesterdayTotal
		if diff > 0 {
			comparison = "up"
			comparisonValue = diff
		} else if diff < 0 {
			comparison = "down"
			comparisonValue = -diff
		} else {
			comparison = "equal"
			comparisonValue = 0
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

// VariableOption 变量选项（用于筛选）
type VariableOption struct {
	VarID       int64  `json:"var_id"`
	VarName     string `json:"var_name"`
	DisplayName string `json:"display_name"`
	DeviceName  string `json:"device_name"`
	GatewayName string `json:"gateway_name"`
}

// GetVariableOptions 获取所有变量选项（用于报警筛选）
func (a *App) GetVariableOptions() ([]VariableOption, error) {
	var options []VariableOption

	// 查询所有变量，关联设备和网关信息
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

// ========================================================
// 任务管理接口
// ========================================================

// GetAllTasks 获取所有任务
func (a *App) GetAllTasks() ([]*models.Task, error) {
	var tasks []*models.Task
	err := database.DB.Order("task_id DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// CreateTask 创建任务
func (a *App) CreateTask(task *models.Task) error {
	return database.DB.Create(task).Error
}

// UpdateTask 更新任务
func (a *App) UpdateTask(taskID int64, updates *models.Task) error {
	return database.DB.Model(&models.Task{}).
		Where("task_id = ?", taskID).
		Updates(updates).Error
}

// TriggerTaskManually 手动触发任务（前端按钮触发）
func (a *App) TriggerTaskManually(taskID int64) error {
	// 导入 workers 包
	scheduler := workers.GetTaskScheduler()
	if scheduler == nil {
		return fmt.Errorf("任务调度器未初始化")
	}

	// 手动触发任务，网关ID使用0，触发数据为nil（会自动填充）
	err := scheduler.ManualTriggerTask(taskID, 0, nil)
	if err != nil {
		return fmt.Errorf("触发任务失败: %v", err)
	}

	return nil
}

// DeleteTask 删除任务
func (a *App) DeleteTask(taskID int64) error {
	return database.DB.Delete(&models.Task{}, taskID).Error
}

// EnableTask 启用任务
func (a *App) EnableTask(taskID int64) error {
	return database.DB.Model(&models.Task{}).
		Where("task_id = ?", taskID).
		Update("is_enabled", true).Error
}

// DisableTask 禁用任务
func (a *App) DisableTask(taskID int64) error {
	return database.DB.Model(&models.Task{}).
		Where("task_id = ?", taskID).
		Update("is_enabled", false).Error
}

// ========================================================
// 网关管理接口
// ========================================================

// Gateway 网关结构
type Gateway struct {
	ID     int    `json:"id"`
	GwName string `json:"gw_name"`
	Status int    `json:"status"`
}

// GetAllGateways 获取所有网关
func (a *App) GetAllGateways() ([]Gateway, error) {
	var gateways []Gateway
	err := database.DB.Table("sys_gateways").
		Select("id, gw_name, status").
		Order("id").
		Find(&gateways).Error
	if err != nil {
		return nil, fmt.Errorf("查询网关失败: %v", err)
	}
	return gateways, nil
}

// ========================================================
// 理论节拍配置管理接口
// ========================================================

// UserConfig 用户配置结构
// BreakTime 休息时间段
type BreakTime struct {
	ID        int    `json:"id"`         // 唯一标识
	Name      string `json:"name"`       // 名称（如"上午茶歇"）
	StartHour int    `json:"start_hour"` // 开始小时
	StartMin  int    `json:"start_min"`  // 开始分钟
	EndHour   int    `json:"end_hour"`   // 结束小时
	EndMin    int    `json:"end_min"`    // 结束分钟
}

type UserConfig struct {
	ProductionCoefficient float64     `json:"production_coefficient"` // 单件加工时间（秒/件）
	DailyWorkMinutes      int         `json:"daily_work_minutes"`     // 每日应工作分钟数（扣除休息后）
	BreakTimes            []BreakTime `json:"break_times"`            // 休息时间段列表
}

// getConfigDir 获取配置目录
func (a *App) getConfigDir() (string, error) {
	// 获取用户配置目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %v", err)
	}

	configDir := filepath.Join(homeDir, ".spd_jghj")

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("创建配置目录失败: %v", err)
	}

	return configDir, nil
}

// getConfigFilePath 获取配置文件路径
func (a *App) getConfigFilePath() (string, error) {
	configDir, err := a.getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "user_config.json"), nil
}

// GetProductionCoefficient 获取理论节拍系数
func (a *App) GetProductionCoefficient() (float64, error) {
	configPath, err := a.getConfigFilePath()
	if err != nil {
		return 10.0, err // 默认值
	}

	// 如果文件不存在，返回默认值
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return 10.0, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 10.0, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config UserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return 10.0, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证值的合理性
	if config.ProductionCoefficient <= 0 {
		return 10.0, nil
	}

	return config.ProductionCoefficient, nil
}

// SetProductionCoefficient 设置理论节拍系数
func (a *App) SetProductionCoefficient(coefficient float64) error {
	if coefficient <= 0 {
		return fmt.Errorf("系数必须大于0")
	}

	configPath, err := a.getConfigFilePath()
	if err != nil {
		return err
	}

	// 读取现有配置（如果存在）
	var config UserConfig
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &config) // 忽略错误，使用默认值
	}

	// 更新系数
	config.ProductionCoefficient = coefficient

	// 保存配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

// GetProductionCoefficientFromEnv 从环境变量获取理论节拍（备用方案）
func (a *App) GetProductionCoefficientFromEnv() float64 {
	if envVal := os.Getenv("PRODUCTION_COEFFICIENT"); envVal != "" {
		if val, err := strconv.ParseFloat(envVal, 64); err == nil && val > 0 {
			return val
		}
	}
	return 10.0 // 默认值
}

// GetDailyWorkMinutes 获取每日应工作分钟数（扣除休息后）
func (a *App) GetDailyWorkMinutes() (int, error) {
	configPath, err := a.getConfigFilePath()
	if err != nil {
		return 460, err // 默认值：7小时40分钟
	}

	// 如果文件不存在，返回默认值
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return 460, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 460, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config UserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return 460, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证值的合理性（应在60-1440分钟之间，即1小时到24小时）
	if config.DailyWorkMinutes <= 0 || config.DailyWorkMinutes > 1440 {
		return 460, nil
	}

	return config.DailyWorkMinutes, nil
}

// SetDailyWorkMinutes 设置每日应工作分钟数
func (a *App) SetDailyWorkMinutes(minutes int) error {
	if minutes <= 0 || minutes > 1440 {
		return fmt.Errorf("每日工作分钟数必须在1-1440之间")
	}

	configPath, err := a.getConfigFilePath()
	if err != nil {
		return err
	}

	// 读取现有配置（如果存在）
	var config UserConfig
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &config) // 忽略错误，使用默认值
	}

	// 更新分钟数
	config.DailyWorkMinutes = minutes

	// 保存配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

// GetBreakTimes 获取休息时间段列表
func (a *App) GetBreakTimes() ([]BreakTime, error) {
	configPath, err := a.getConfigFilePath()
	if err != nil {
		return getDefaultBreakTimes(), err
	}

	// 如果文件不存在，返回默认值
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultBreakTimes(), nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return getDefaultBreakTimes(), fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config UserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return getDefaultBreakTimes(), fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 如果没有配置，返回默认值
	if len(config.BreakTimes) == 0 {
		return getDefaultBreakTimes(), nil
	}

	return config.BreakTimes, nil
}

// getDefaultBreakTimes 获取默认休息时间段
func getDefaultBreakTimes() []BreakTime {
	return []BreakTime{
		{ID: 1, Name: "上午茶歇", StartHour: 9, StartMin: 40, EndHour: 9, EndMin: 50},
		{ID: 2, Name: "午餐休息", StartHour: 11, StartMin: 40, EndHour: 12, EndMin: 20},
		{ID: 3, Name: "下午茶歇", StartHour: 14, StartMin: 20, EndHour: 14, EndMin: 30},
	}
}

// SetBreakTimes 设置休息时间段列表
func (a *App) SetBreakTimes(breakTimes []BreakTime) error {
	// 验证时间段的合理性
	for _, bt := range breakTimes {
		if bt.StartHour < 0 || bt.StartHour > 23 || bt.EndHour < 0 || bt.EndHour > 23 {
			return fmt.Errorf("小时必须在0-23之间")
		}
		if bt.StartMin < 0 || bt.StartMin > 59 || bt.EndMin < 0 || bt.EndMin > 59 {
			return fmt.Errorf("分钟必须在0-59之间")
		}
		startInMin := bt.StartHour*60 + bt.StartMin
		endInMin := bt.EndHour*60 + bt.EndMin
		if startInMin >= endInMin {
			return fmt.Errorf("结束时间必须晚于开始时间")
		}
	}

	configPath, err := a.getConfigFilePath()
	if err != nil {
		return err
	}

	// 读取现有配置（如果存在）
	var config UserConfig
	if data, err := os.ReadFile(configPath); err == nil {
		json.Unmarshal(data, &config) // 忽略错误，使用默认值
	}

	// 更新休息时间段
	config.BreakTimes = breakTimes

	// 保存配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

// GetSystemConfig 获取完整的系统配置
func (a *App) GetSystemConfig() (*UserConfig, error) {
	configPath, err := a.getConfigFilePath()
	if err != nil {
		return getDefaultConfig(), err
	}

	// 如果文件不存在，返回默认值
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return getDefaultConfig(), fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config UserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return getDefaultConfig(), fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 填充默认值
	if config.ProductionCoefficient <= 0 {
		config.ProductionCoefficient = 100.0
	}
	if config.DailyWorkMinutes <= 0 {
		config.DailyWorkMinutes = 460
	}
	if len(config.BreakTimes) == 0 {
		config.BreakTimes = getDefaultBreakTimes()
	}

	return &config, nil
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *UserConfig {
	return &UserConfig{
		ProductionCoefficient: 100.0,
		DailyWorkMinutes:      460,
		BreakTimes:            getDefaultBreakTimes(),
	}
}

// SetSystemConfig 设置完整的系统配置
func (a *App) SetSystemConfig(config *UserConfig) error {
	// 验证配置
	if config.ProductionCoefficient <= 0 {
		return fmt.Errorf("单件加工时间必须大于0")
	}
	if config.DailyWorkMinutes <= 0 || config.DailyWorkMinutes > 1440 {
		return fmt.Errorf("每日工作分钟数必须在1-1440之间")
	}

	// 验证休息时间段
	for _, bt := range config.BreakTimes {
		if bt.StartHour < 0 || bt.StartHour > 23 || bt.EndHour < 0 || bt.EndHour > 23 {
			return fmt.Errorf("小时必须在0-23之间")
		}
		if bt.StartMin < 0 || bt.StartMin > 59 || bt.EndMin < 0 || bt.EndMin > 59 {
			return fmt.Errorf("分钟必须在0-59之间")
		}
		startInMin := bt.StartHour*60 + bt.StartMin
		endInMin := bt.EndHour*60 + bt.EndMin
		if startInMin >= endInMin {
			return fmt.Errorf("结束时间必须晚于开始时间")
		}
	}

	configPath, err := a.getConfigFilePath()
	if err != nil {
		return err
	}

	// 保存配置
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失败: %v", err)
	}

	return nil
}

// ========================================================
// 能耗数据接口
// ========================================================

// GetTodayEnergyConsumption 获取今日电能消耗（最大值-最小值）
func (a *App) GetTodayEnergyConsumption(varID int64) (float64, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var result struct {
		MaxVal *float64 `gorm:"column:max_val"`
		MinVal *float64 `gorm:"column:min_val"`
	}

	err := database.DB.Table("sys_data_history").
		Select("MAX(val) as max_val, MIN(val) as min_val").
		Where("var_id = ? AND created_at >= ? AND val IS NOT NULL", varID, todayStart).
		Scan(&result).Error

	if err != nil {
		return 0, fmt.Errorf("查询今日电能失败: %v", err)
	}

	if result.MaxVal == nil || result.MinVal == nil {
		return 0, nil
	}

	consumption := *result.MaxVal - *result.MinVal
	if consumption < 0 {
		consumption = 0
	}

	return consumption, nil
}

// DeviceEnergyData 设备能耗数据
type DeviceEnergyData struct {
	DeviceID         int     `json:"device_id"`
	DeviceName       string  `json:"device_name"`
	RealTimePower    float64 `json:"real_time_power"`
	TodayConsumption float64 `json:"today_consumption"`
	PowerUnit        string  `json:"power_unit"`
	EnergyUnit       string  `json:"energy_unit"`
}

// GetAllDevicesEnergyData 获取所有设备能耗数据
func (a *App) GetAllDevicesEnergyData() ([]*DeviceEnergyData, error) {
	// 配置：设备ID -> (功率变量ID, 电能变量ID)
	config := map[int]struct {
		PowerVarID  int64
		EnergyVarID int64
	}{
		1: {PowerVarID: 86, EnergyVarID: 81},
		2: {PowerVarID: 110, EnergyVarID: 107},
	}

	tagManager := core.GetTagManager()
	results := make([]*DeviceEnergyData, 0, len(config))

	for deviceID, cfg := range config {
		data := &DeviceEnergyData{
			DeviceID:   deviceID,
			PowerUnit:  "kW",
			EnergyUnit: "kWh",
		}

		// 获取设备名称
		if device, err := database.GetDeviceByID(deviceID); err == nil {
			data.DeviceName = device.DeviceName
		} else {
			data.DeviceName = fmt.Sprintf("设备%d", deviceID)
		}

		// 从内存获取实时功率
		if powerTag, ok := tagManager.GetTag(cfg.PowerVarID); ok && powerTag != nil {
			data.RealTimePower = powerTag.GetValue()
			if powerTag.Unit != "" {
				data.PowerUnit = powerTag.Unit
			}
		}

		// 从历史表计算今日电能
		if consumption, err := a.GetTodayEnergyConsumption(cfg.EnergyVarID); err == nil {
			data.TodayConsumption = consumption
		}

		results = append(results, data)
	}

	return results, nil
}
