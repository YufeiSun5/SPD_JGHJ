package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"gin-mqtt-pgsql/backend/service"
	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
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
	return service.SyncErrorCode(errorCode, errorMsg)
}

// GetAllErrorCodes 获取所有错误码
func (a *App) GetAllErrorCodes() ([]*database.ErrorCode, error) {
	return service.GetAllErrorCodes()
}

// GetRealtimeData 获取所有测点实时数据 (线程安全)
func (a *App) GetRealtimeData() []TagData {
	return service.GetRealtimeData()
}

// GetChannelStats 获取通道状态
func (a *App) GetChannelStats() map[string]int {
	return core.GetChannelStats()
}

// GetSystemMonitor 获取系统监控数据 (通道队列长度、任务统计)
func (a *App) GetSystemMonitor() map[string]interface{} {
	return service.GetSystemMonitor()
}

// TagData 测点数据传输对象
type TagData = service.TagData

// ========================================================
// MES 工单管理接口
// ========================================================

// GetAllOrders 获取所有工单
func (a *App) GetAllOrders() ([]*models.ProOrder, error) {
	return service.GetAllOrders()
}

// CreateOrder 创建工单
func (a *App) CreateOrder(orderNo, productCode string, planQty int, targetDeviceID *int) (*models.ProOrder, error) {
	return service.CreateOrder(orderNo, productCode, planQty, targetDeviceID)
}

// UpdateOrder 更新工单
func (a *App) UpdateOrder(id int64, productCode *string, planQty *int, status *int8, targetDeviceID *int) error {
	return service.UpdateOrder(id, productCode, planQty, status, targetDeviceID)
}

// DeleteOrder 删除工单
func (a *App) DeleteOrder(id int64) error {
	return service.DeleteOrder(id)
}

// StartProductionSmart 智能开工（自动暂停该设备的其他工单，自动获取班次信息）
func (a *App) StartProductionSmart(orderID int64) (*models.ProProductionRun, error) {
	return service.StartProductionSmart(orderID)
}

// GetRealHourlyProduction 获取真实的每小时产量统计（调用数据访问层）
func (a *App) GetRealHourlyProduction() ([]database.HourlyProductionPulse, error) {
	return service.GetRealHourlyProduction()
}

// GetHourlyProductionAccurate 获取每小时精确产量统计（完全参数化，无硬编码）
// configs: 设备变量配置数组，每个设备指定产量ID、NG按钮ID和设备名称
// 如果传入nil，使用默认配置（设备1和设备2）
// GetHourlyProductionAccurate 获取逻辑日每小时精确产量统计
// CN: 自动从 GetShiftsForLogicalDay 获取逻辑日期，确保 7:00 前显示昨天同一时段的数据。
// EN: Automatically resolves the logical date from shift config; before first-shift start = yesterday.
// JP: シフト設定から論理日付を自動解決。第1シフト開始前は前日のデータを返す。
func (a *App) GetHourlyProductionAccurate(configs []database.DeviceVarConfig) ([]database.HourlyProductionAccurate, error) {
	return service.GetHourlyProductionAccurate(configs)
}

// GetMonthlyProductionAccurate 获取当月产量汇总统计（按设备，不走工单表）
func (a *App) GetMonthlyProductionAccurate(configs []database.DeviceVarConfig) ([]database.MonthlyProductionAccurate, error) {
	return service.GetMonthlyProductionAccurate(configs)
}

// GetMonthlyQualityByOrder 从工单表获取当月各设备良品率汇总
func (a *App) GetMonthlyQualityByOrder() ([]database.DeviceQualityStat, error) {
	return service.GetMonthlyQualityByOrder()
}

// GetMonthlyDailyQualityTrend 获取本月每日良品率趋势
func (a *App) GetMonthlyDailyQualityTrend() ([]database.DailyQualityTrend, error) {
	return service.GetMonthlyDailyQualityTrend()
}

// GetDailyQualityByRun 从生产运行记录获取今日各设备良品率
func (a *App) GetDailyQualityByRun() ([]database.DeviceQualityStat, error) {
	return service.GetDailyQualityByRun()
}

// GetActiveOrderQuality 获取当前在产工单（生产中+暂停）各设备良品率
func (a *App) GetActiveOrderQuality() ([]database.DeviceQualityStat, error) {
	return service.GetActiveOrderQuality()
}

// ========================================================
// MES 设备管理接口
// ========================================================

// GetAllDevices 获取所有设备
func (a *App) GetAllDevices() ([]*models.SysDevice, error) {
	return service.GetAllDevices()
}

// ========================================================
// MES 人员管理接口
// ========================================================

// GetAllStaff 获取所有员工
func (a *App) GetAllStaff(teamID *int, isActive *int8) ([]*models.SysStaff, error) {
	return service.GetAllStaff(teamID, isActive)
}

// CreateStaff 创建员工
func (a *App) CreateStaff(staffCode, name string, currentTeamID *int) (*models.SysStaff, error) {
	return service.CreateStaff(staffCode, name, currentTeamID)
}

// UpdateStaff 更新员工
// 特殊值：currentTeamID = -1 表示清空班组（设为 NULL）
func (a *App) UpdateStaff(id int, name *string, currentTeamID *int, isActive *int8) error {
	return service.UpdateStaff(id, name, currentTeamID, isActive)
}

// DeleteStaff 删除员工
func (a *App) DeleteStaff(id int) error {
	return service.DeleteStaff(id)
}

// TransferStaff 调动员工到新班组
func (a *App) TransferStaff(staffID, newTeamID int, operatorName *string) error {
	return service.TransferStaff(staffID, newTeamID, operatorName)
}

// GetStaffHistory 获取员工调动历史
func (a *App) GetStaffHistory(staffID int) ([]*models.SysStaffHistory, error) {
	return service.GetStaffHistory(staffID)
}

// ========================================================
// MES 班组管理接口
// ========================================================

// GetAllTeams 获取所有班组
func (a *App) GetAllTeams(status *int8) ([]*models.SysTeam, error) {
	return service.GetAllTeams(status)
}

// CreateTeam 创建班组
func (a *App) CreateTeam(teamName string, leaderName *string) (*models.SysTeam, error) {
	return service.CreateTeam(teamName, leaderName)
}

// UpdateTeam 更新班组
func (a *App) UpdateTeam(id int, teamName *string, leaderName *string, status *int8) error {
	return service.UpdateTeam(id, teamName, leaderName, status)
}

// DeleteTeam 删除班组
func (a *App) DeleteTeam(id int) error {
	return service.DeleteTeam(id)
}

// ========================================================
// MES 设备登录/班次管理接口
// ========================================================

// DeviceLogin 设备登录/上班打卡
func (a *App) DeviceLogin(deviceID, teamID int, staffIDs []int) (*models.ProMachineSession, error) {
	return service.DeviceLogin(deviceID, teamID, staffIDs)
}

// DeviceLogout 设备登出/下班打卡
func (a *App) DeviceLogout(deviceID int) (*models.ProMachineSession, error) {
	return service.DeviceLogout(deviceID)
}

// GetActiveSession 获取设备当前活动班次
func (a *App) GetActiveSession(deviceID int) (*models.ProMachineSession, error) {
	return service.GetActiveSession(deviceID)
}

// GetAllActiveSessions 获取所有设备的活动班次
func (a *App) GetAllActiveSessions() ([]*models.ProMachineSession, error) {
	return service.GetAllActiveSessions()
}

// GetSessionHistory 获取班次历史记录
func (a *App) GetSessionHistory(deviceID *int, teamID *int, startDate, endDate string) ([]*models.ProMachineSession, error) {
	return service.GetSessionHistory(deviceID, teamID, startDate, endDate)
}

// GetSessionStats 获取班次统计信息
func (a *App) GetSessionStats(sessionID int64) (*models.SessionStatusResponse, error) {
	return service.GetSessionStats(sessionID)
}

// GetStaffAttendance 获取员工出勤记录
func (a *App) GetStaffAttendance(staffID int, startDate, endDate string) ([]*models.ProMachineSession, error) {
	return service.GetStaffAttendance(staffID, startDate, endDate)
}

// ========================================================
// IOT 历史查询接口
// ========================================================

// TagInfo 变量信息
type TagInfo = service.TagInfo

// GetAllTags 获取所有变量配置（用于历史查询）
func (a *App) GetAllTags() []TagInfo {
	return service.GetAllTags()
}

// HistoryRecord 历史数据记录
type HistoryRecord = service.HistoryRecord

// HistoryDataResponse 历史数据响应结构
type HistoryDataResponse = service.HistoryDataResponse

// GetHistoryData 获取变量历史数据（分页）
func (a *App) GetHistoryData(varID int64, startTime, endTime string, page, pageSize int) (HistoryDataResponse, error) {
	return service.GetHistoryData(varID, startTime, endTime, page, pageSize)
}

// ========================================================
// IOT 配置管理接口
// ========================================================

// GetAllVariables 获取所有变量配置
func (a *App) GetAllVariables() ([]database.VariableRow, error) {
	return service.GetAllVariables()
}

// UpdateVariable 更新变量配置
func (a *App) UpdateVariable(variable database.VariableRow) error {
	return service.UpdateVariable(variable)
}

// BatchUpdateVariables 批量更新变量配置
func (a *App) BatchUpdateVariables(variables []database.VariableRow) error {
	return service.BatchUpdateVariables(variables)
}

// CreateVariable 创建新变量配置
func (a *App) CreateVariable(variable database.VariableRow) error {
	return service.CreateVariable(variable)
}

// DeleteVariable 删除变量配置
func (a *App) DeleteVariable(id int64) error {
	return service.DeleteVariable(id)
}

// BatchDeleteVariables 批量删除变量配置
func (a *App) BatchDeleteVariables(ids []int64) error {
	return service.BatchDeleteVariables(ids)
}

// ========================================================
// 设备状态管理接口
// ========================================================

// DeviceStatusData 设备状态数据传输对象
type DeviceStatusData = service.DeviceStatusData

// ExtraData 扩展数据结构
type ExtraData = service.ExtraData

// GetAllDevicesStatus 获取所有设备状态（含操作人员）
func (a *App) GetAllDevicesStatus() ([]DeviceStatusData, error) {
	return service.GetAllDevicesStatus()
}

// GetDeviceStatusHistory 获取设备24小时状态历史（用于甘特图）
func (a *App) GetDeviceStatusHistory(deviceID int) ([]*models.SysDeviceStatus, error) {
	return service.GetDeviceStatusHistory(deviceID)
}

// DeviceStatusHistoryData 设备状态历史数据（含班次信息）
type DeviceStatusHistoryData = service.DeviceStatusHistoryData

// GetDeviceStatusHistoryAll 获取所有设备状态历史（支持筛选，含班次信息）
func (a *App) GetDeviceStatusHistoryAll(deviceID *int, startTimeStr, endTimeStr string) ([]DeviceStatusHistoryData, error) {
	return service.GetDeviceStatusHistoryAll(deviceID, startTimeStr, endTimeStr)
}

// GetDeviceStatusStats 获取设备状态统计汇总
func (a *App) GetDeviceStatusStats() (map[string]interface{}, error) {
	return service.GetDeviceStatusStats()
}

// ========================================================
// 统计数据接口 (用于驾驶舱)
// ========================================================

// GetHourlyProduction 获取今日按小时统计的产量
func (a *App) GetHourlyProduction(deviceID *int) ([]database.HourlyProduction, error) {
	return service.GetHourlyProduction(deviceID)
}

// GetStaffEfficiency 获取员工绩效统计
func (a *App) GetStaffEfficiency(startTime, endTime *time.Time) ([]database.StaffEfficiency, error) {
	return service.GetStaffEfficiency(startTime, endTime)
}

// GetDeviceUtilizationTrend 获取设备利用率趋势
func (a *App) GetDeviceUtilizationTrend(deviceID *int) ([]database.DeviceUtilizationTrend, error) {
	return service.GetDeviceUtilizationTrend(deviceID)
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

	cfgSet, cfgErr := a.getOEEDeviceConfigSet()
	if cfgErr != nil {
		fmt.Printf("⚠️ 读取设备级OEE配置失败，使用默认配置: %v\n", cfgErr)
	}

	result := make([]ShiftOEESummary, 0, len(ldShifts))
	for _, s := range ldShifts {
		if !s.HasArrived {
			continue
		}

		devCfgs := cfgSet.configsForSchedule(s.ScheduleID)

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
//
//	口径与班次快照（pro_shift_snapshots.quality_pct）完全一致，均来自 sys_data_history。
//	之前的实现调用 GetShiftQualityByRun（pro_production_runs 工单运行记录），
//	与快照数据源不同导致驾驶舱良品率和班组追溯页面数值对不上，现已修正。
//	跨零点班次：endMin <= startMin 时 shiftEnd 加 24h。
//
// EN: For each arrived shift, calls GetShiftWindowProduction per configured device.
//
//	Data source is sys_data_history, identical to snapshot quality_pct calculation.
//	Previous impl used pro_production_runs (GetShiftQualityByRun), causing mismatch with
//	shift report; now unified with snapshot source.
//
// JP: 到達済シフトごとに設備別 GetShiftWindowProduction を呼び出す。
//
//	データソースは sys_data_history で、スナップショットの quality_pct と完全に一致する。
//	以前は pro_production_runs を参照していたため班組追跡画面と乖離していた。修正済み。
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

type oeeDeviceConfigSet struct {
	All        []database.DeviceOEEConfig
	BySchedule map[int][]database.DeviceOEEConfig
	Unassigned []database.DeviceOEEConfig
}

func (s *oeeDeviceConfigSet) configsForSchedule(scheduleID int) []database.DeviceOEEConfig {
	if s == nil {
		return nil
	}
	cfgs := make([]database.DeviceOEEConfig, 0, len(s.BySchedule[scheduleID])+len(s.Unassigned))
	cfgs = append(cfgs, s.BySchedule[scheduleID]...)
	cfgs = append(cfgs, s.Unassigned...)
	if len(cfgs) == 0 {
		cfgs = append(cfgs, s.All...)
	}
	return cfgs
}

func (a *App) getGlobalCycleTimeFallback() float64 {
	config, err := service.GetSystemConfig()
	if err != nil || config == nil || config.ProductionCoefficient <= 0 {
		return 100.0
	}
	return config.ProductionCoefficient
}

func defaultOEEDeviceConfigSet(ct float64) *oeeDeviceConfigSet {
	cfgs := []database.DeviceOEEConfig{
		hardcodedOEEConfigDefault(1, ct),
		hardcodedOEEConfigDefault(2, ct),
	}
	return &oeeDeviceConfigSet{
		All:        cfgs,
		BySchedule: map[int][]database.DeviceOEEConfig{},
		Unassigned: cfgs,
	}
}

// getOEEDeviceConfigSet 统一读取设备级 OEE 配置。
//
// CN: OEE 只从 sys_devices.cycle_time 读取设备级 CT，未配置时才回退到全局默认 CT。
//
//	前端不再传入 OEE configs，避免两台设备被同一个旧节拍覆盖。
//
// EN: OEE reads per-device CT from sys_devices.cycle_time, falling back to the global default only when empty.
//
//	The frontend no longer supplies OEE configs, so two devices cannot be overwritten by one legacy CT.
//
// JP: OEE は sys_devices.cycle_time から設備別 CT を読み、未設定時のみグローバル既定値へ戻す。
//
//	フロントから OEE configs は渡さず、2台設備が旧単一 CT で上書きされることを防ぐ。
func (a *App) getOEEDeviceConfigSet() (*oeeDeviceConfigSet, error) {
	fallbackCT := a.getGlobalCycleTimeFallback()
	fallback := defaultOEEDeviceConfigSet(fallbackCT)
	if database.DB == nil {
		return fallback, fmt.Errorf("数据库未连接")
	}

	var devices []models.SysDevice
	if err := database.DB.Select("id, device_name, schedule_id, cycle_time").Order("id").Find(&devices).Error; err != nil {
		return fallback, err
	}

	result := &oeeDeviceConfigSet{
		All:        []database.DeviceOEEConfig{},
		BySchedule: map[int][]database.DeviceOEEConfig{},
		Unassigned: []database.DeviceOEEConfig{},
	}
	for _, dev := range devices {
		ct := fallbackCT
		if dev.CycleTime != nil && *dev.CycleTime > 0 {
			ct = *dev.CycleTime
		}
		cfg := hardcodedOEEConfig(dev.ID, ct)
		if cfg == nil {
			continue
		}
		result.All = append(result.All, *cfg)
		if dev.ScheduleID != nil && *dev.ScheduleID > 0 {
			result.BySchedule[*dev.ScheduleID] = append(result.BySchedule[*dev.ScheduleID], *cfg)
		} else {
			result.Unassigned = append(result.Unassigned, *cfg)
		}
	}
	if len(result.All) == 0 {
		return fallback, nil
	}
	return result, nil
}

// getOEEConfigs 读取OEE所需的配置（设备配置和休息时间），供内部复用。
// CN: 设备配置来自 sys_devices.cycle_time；全局配置只作为未配置设备的 CT 兜底。
// EN: Device config comes from sys_devices.cycle_time; the global config is only a CT fallback.
// JP: 設備設定は sys_devices.cycle_time から取得し、グローバル設定は未設定設備の CT フォールバックのみ。
func (a *App) getOEEConfigs() ([]database.DeviceOEEConfig, []database.BreakTimeConfig, error) {
	// 读取休息时间配置
	breakTimes, err := a.GetBreakTimes()
	if err != nil {
		fmt.Printf("⚠️ 读取休息时间配置失败，使用默认配置: %v\n", err)
		breakTimes = service.DefaultBreakTimes()
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

	cfgSet, cfgErr := a.getOEEDeviceConfigSet()
	if cfgErr != nil {
		fmt.Printf("⚠️ 读取设备级OEE配置失败，使用默认配置: %v\n", cfgErr)
	}
	return cfgSet.All, dbBreakTimes, nil
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

	workStartMin := firstShift.StartHour*60 + firstShift.StartMin + firstShift.CalendarDayOffset*24*60

	lastEndMin := lastShift.EndHour*60 + lastShift.EndMin + lastShift.CalendarDayOffset*24*60
	lastStartMin := lastShift.StartHour*60 + lastShift.StartMin + lastShift.CalendarDayOffset*24*60
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
		sStartMin := s.StartHour*60 + s.StartMin + s.CalendarDayOffset*24*60
		sEndRaw := s.EndHour*60 + s.EndMin + s.CalendarDayOffset*24*60
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

	// calendarDayOffset 偏移：三班等在逻辑日+1 自然日运行的班次，全段时间均平移 24h
	// EN: Shifts running on logical_date+1 calendar day (e.g., 三班 0:00-7:40) get all hours offset by +24.
	// JP: 論理日+1日目で動くシフト（三班 0:00-7:40 等）は全時刻に +24h を加算する。
	hourOffset := s.CalendarDayOffset * 24

	// 跨零点判定：结束分钟数 < 开始分钟数（例如 22:00→06:00，360 < 1320）
	crossMidnight := endMinTotal < startMinTotal

	startHourSQL := s.StartHour + hourOffset
	// 跨日时，WorkEnd 用 "30:00" 等超 24 格式，使 ADDTIME 指向次日
	endHourSQL := s.EndHour + hourOffset
	if crossMidnight {
		endHourSQL += 24
	}

	// HourEnd = 最后时间桶 hour_idx 的上限（WHERE hour_idx < HourEnd）
	hourEndLimit := endHourSQL
	if s.EndMin > 0 {
		hourEndLimit = endHourSQL + 1
	}
	if hourEndLimit <= startHourSQL {
		hourEndLimit = startHourSQL + 1
	}

	window := &database.ShiftWindow{
		WorkStart:   fmt.Sprintf("%02d:%02d", startHourSQL, s.StartMin),
		WorkEnd:     fmt.Sprintf("%02d:%02d", endHourSQL, s.EndMin),
		LogicalDate: s.LogicalDate,
		HourStart:   startHourSQL,
		HourEnd:     hourEndLimit,
	}

	breaks := make([]database.BreakTimeConfig, 0, len(s.Breaks))
	for _, b := range s.Breaks {
		bStartMin := b.StartHour*60 + b.StartMin
		bEndMin := b.EndHour*60 + b.EndMin
		startH := b.StartHour + hourOffset
		endH := b.EndHour + hourOffset

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
//  2. 从 DB 读取设备的 schedule_id + cycle_time
//  3. 按 schedule_id 将设备分组，CT 使用设备级配置
//  4. 每组单独调用 database.GetHourlyOEE（使用各自班次的时间窗口 + 休息段）
//  5. 若设备无 schedule_id，回退到合并所有班次的宽窗口查询
//
// EN: Groups devices by schedule_id and reads per-device CT from DB before querying OEE.
// JP: 設備を schedule_id でグループ化し、DB の設備別 CT を使って OEE を照会する。
func (a *App) GetHourlyOEE() ([]database.HourlyOEE, error) {
	cfgSet, cfgErr := a.getOEEDeviceConfigSet()
	if cfgErr != nil {
		fmt.Printf("⚠️ 读取设备级OEE配置失败，使用默认配置: %v\n", cfgErr)
	}

	// 1. 获取逻辑日班次列表
	ldShifts, err := a.GetShiftsForLogicalDay()
	if err != nil || len(ldShifts) == 0 {
		fmt.Printf("⚠️ 无活动班次，OEE使用默认配置: %v\n", err)
		return database.GetHourlyOEE(cfgSet.All, nil, nil)
	}

	// 3. 构建 scheduleID → []DeviceOEEConfig 映射
	// CN: 设备按所属时间安排组（ScheduleID）分组，同组的设备共享同一套班次时间窗口。
	//     CT 优先使用设备级别配置（SysDevice.CycleTime），NULL 时 fallback 到全局默认值。
	// EN: Group devices by their schedule (ScheduleID); devices in the same schedule share the same shift windows.
	//     CT prefers device-level config (SysDevice.CycleTime), falls back to global default if NULL.
	// JP: 設備をスケジュール（ScheduleID）でグループ化。同グループの設備は同一シフトウィンドウを共有する。
	//     CT はデバイス単位設定を優先し、NULL の場合はグローバルデフォルトにフォールバック。
	scheduleDeviceMap := cfgSet.BySchedule
	unassigned := cfgSet.Unassigned

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
		return database.GetHourlyOEE(cfgSet.All, nil, nil)
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
			"cycle_time":       r.CycleTime,
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
	cfgSet, cfgErr := a.getOEEDeviceConfigSet()
	if cfgErr != nil {
		fmt.Printf("⚠️ DebugOEEByShift读取设备级配置失败，使用默认配置: %v\n", cfgErr)
	}

	ldShifts, shiftErr := a.GetShiftsForLogicalDay()
	if shiftErr != nil || len(ldShifts) == 0 {
		// 无活动班次：降级为单次全天查询，包装成一个伪班次组返回
		rows, _, err2 := database.GetHourlyOEEWithSQL(cfgSet.All, nil, nil)
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
		if !s.HasArrived {
			continue
		}
		window, breaks := buildShiftOEEWindow(s)
		rows, _, qErr := database.GetHourlyOEEWithSQL(cfgSet.configsForSchedule(s.ScheduleID), breaks, window)
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
	if len(result) == 0 {
		rows, _, err2 := database.GetHourlyOEEWithSQL(cfgSet.All, nil, nil)
		if err2 != nil {
			return nil, fmt.Errorf("DebugOEEByShift兜底查询失败: %w", err2)
		}
		return []map[string]interface{}{{
			"shift_id":    0,
			"shift_name":  "当前逻辑日（无已到达班次）",
			"shift_range": "",
			"rows":        oeeDebugRowsToMaps(rows),
		}}, nil
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
			"cycle_time":       r.CycleTime,
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
type AlarmRecordData = service.AlarmRecordData

// GetAlarmRecords 获取报警记录（支持筛选）
// 🔥 改造：联表查询报警表+错误码表，组合显示
func (a *App) GetAlarmRecords(ackStatus *int, startTimeStr, endTimeStr string, varID *int64) ([]AlarmRecordData, error) {
	return service.GetAlarmRecords(ackStatus, startTimeStr, endTimeStr, varID)
}

// AckAlarm 确认报警
func (a *App) AckAlarm(alarmID int64) error {
	return service.AckAlarm(alarmID)
}

// GetTodayUnacknowledgedAlarmCount 获取今日未确认的报警数（ack_status=0）
func (a *App) GetTodayUnacknowledgedAlarmCount() (int64, error) {
	return service.GetTodayUnacknowledgedAlarmCount()
}

// HourlyAlarmCount 每小时报警数统计
type HourlyAlarmCount = service.HourlyAlarmCount

// GetTodayHourlyAlarmCount 获取今日每小时的报警数（用于迷你图）
func (a *App) GetTodayHourlyAlarmCount() ([]HourlyAlarmCount, error) {
	return service.GetTodayHourlyAlarmCount()
}

// GetActiveAlarmCount 获取当前进行中工单相关的报警数（保留旧接口以兼容）
func (a *App) GetActiveAlarmCount() (int64, error) {
	return service.GetActiveAlarmCount()
}

// GetAlarmStats 获取报警统计
func (a *App) GetAlarmStats() (map[string]interface{}, error) {
	return service.GetAlarmStats()
}

// VariableOption 变量选项（用于筛选）
type VariableOption = service.VariableOption

// GetVariableOptions 获取所有变量选项（用于报警筛选）
func (a *App) GetVariableOptions() ([]VariableOption, error) {
	return service.GetVariableOptions()
}

// ========================================================
// 任务管理接口
// ========================================================

// GetAllTasks 获取所有任务
func (a *App) GetAllTasks() ([]*models.Task, error) {
	return service.GetAllTasks()
}

// CreateTask 创建任务
func (a *App) CreateTask(task *models.Task) error {
	return service.CreateTask(task)
}

// UpdateTask 更新任务
func (a *App) UpdateTask(taskID int64, updates *models.Task) error {
	return service.UpdateTask(taskID, updates)
}

// TriggerTaskManually 手动触发任务（前端按钮触发）
func (a *App) TriggerTaskManually(taskID int64) error {
	return service.TriggerTaskManually(taskID)
}

// DeleteTask 删除任务
func (a *App) DeleteTask(taskID int64) error {
	return service.DeleteTask(taskID)
}

// EnableTask 启用任务
func (a *App) EnableTask(taskID int64) error {
	return service.EnableTask(taskID)
}

// DisableTask 禁用任务
func (a *App) DisableTask(taskID int64) error {
	return service.DisableTask(taskID)
}

// ========================================================
// 网关管理接口
// ========================================================

// Gateway 网关结构
type Gateway = service.Gateway

// GetAllGateways 获取所有网关
func (a *App) GetAllGateways() ([]Gateway, error) {
	return service.GetAllGateways()
}

// ========================================================
// 理论节拍配置管理接口
// ========================================================

type BreakTime = service.BreakTime
type UserConfig = service.UserConfig

// GetDailyWorkMinutes 获取每日应工作分钟数（扣除休息后）
func (a *App) GetDailyWorkMinutes() (int, error) {
	return service.GetDailyWorkMinutes()
}

// SetDailyWorkMinutes 设置每日应工作分钟数
func (a *App) SetDailyWorkMinutes(minutes int) error {
	return service.SetDailyWorkMinutes(minutes)
}

// GetBreakTimes 获取休息时间段列表
func (a *App) GetBreakTimes() ([]BreakTime, error) {
	return service.GetBreakTimes()
}

// SetBreakTimes 设置休息时间段列表
func (a *App) SetBreakTimes(breakTimes []BreakTime) error {
	return service.SetBreakTimes(breakTimes)
}

// GetSystemConfig 获取完整的系统配置
func (a *App) GetSystemConfig() (*UserConfig, error) {
	return service.GetSystemConfig()
}

// SetSystemConfig 设置完整的系统配置
func (a *App) SetSystemConfig(config *UserConfig) error {
	return service.SetSystemConfig(config)
}

// ========================================================
// 能耗数据接口
// ========================================================

// GetTodayEnergyConsumption 获取今日电能消耗（最大值-最小值）
func (a *App) GetTodayEnergyConsumption(varID int64) (float64, error) {
	return service.GetTodayEnergyConsumption(varID)
}

type DeviceEnergyData = service.DeviceEnergyData

// GetAllDevicesEnergyData 获取所有设备能耗数据
func (a *App) GetAllDevicesEnergyData() ([]*DeviceEnergyData, error) {
	return service.GetAllDevicesEnergyData()
}
