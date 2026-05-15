package main

// ========================================================
// 时间安排组 & 班次配置管理接口（Wails 绑定）
// Shift Schedule & Shift Configuration Management (Wails Bindings)
// スケジュールグループ & シフト設定管理インターフェース（Wailsバインド）
//
// CN: 业务逻辑已下沉到 backend/service/shift_service.go，本文件只保留 Wails 方法名和类型别名。
// EN: Business logic lives in backend/service/shift_service.go; this file keeps only Wails method names and type aliases.
// JP: 業務ロジックは backend/service/shift_service.go に移し、このファイルは Wails メソッド名と型エイリアスのみ保持する。
// ========================================================

import "gin-mqtt-pgsql/backend/service"

type ShiftBreak = service.ShiftBreak
type ShiftConfig = service.ShiftConfig
type ShiftScheduleConfig = service.ShiftScheduleConfig
type LogicalDayShift = service.LogicalDayShift

// GetShiftSchedules 获取所有时间安排组（含各组的班次和休息段）。
// Returns all schedule groups with their shifts and break periods.
// 全スケジュールグループをシフト・休憩時間帯リストごと返す。
func (a *App) GetShiftSchedules() ([]ShiftScheduleConfig, error) {
	return service.GetShiftSchedules()
}

// GetShifts 获取所有活动班次（平铺，跨所有时间安排组；供内部逻辑日计算使用）。
// Returns all active shifts flattened across all schedules.
// 全スケジュールの全アクティブシフトをフラット化して返す。
func (a *App) GetShifts() ([]ShiftConfig, error) {
	return service.GetShifts()
}

// SaveShiftSchedules 保存全部时间安排组（upsert 语义）。
// Saves all schedule groups with upsert semantics.
// 全スケジュールグループを upsert セマンティクスで保存する。
func (a *App) SaveShiftSchedules(schedules []ShiftScheduleConfig) error {
	return service.SaveShiftSchedules(schedules)
}

// GetScheduleDeviceIDs 获取指定时间安排组关联的所有设备 ID 列表。
// Returns device IDs assigned to the given schedule group.
// 指定スケジュールグループに紐づく設備 ID 一覧を返す。
func (a *App) GetScheduleDeviceIDs(scheduleID int) ([]int, error) {
	return service.GetScheduleDeviceIDs(scheduleID)
}

// SetDeviceSchedule 设置设备关联的时间安排组（scheduleID <= 0 表示解除绑定）。
// Sets a device schedule group; scheduleID <= 0 clears the binding.
// デバイスのスケジュールグループを設定し、scheduleID <= 0 で解除する。
func (a *App) SetDeviceSchedule(deviceID int, scheduleID int) error {
	return service.SetDeviceSchedule(deviceID, scheduleID)
}

// GetDeviceCycleTime 获取指定设备的理论节拍（CT，秒/件）。
// Returns device-level cycle time in seconds per piece.
// 指定設備のサイクルタイム（秒/個）を返す。
func (a *App) GetDeviceCycleTime(deviceID int) (float64, error) {
	return service.GetDeviceCycleTime(deviceID)
}

// SetDeviceCycleTime 设置设备的理论节拍（CT，秒/件）。ct <= 0 时置 NULL。
// Sets device-level cycle time; ct <= 0 stores NULL.
// デバイスのサイクルタイムを設定し、ct <= 0 で NULL にする。
func (a *App) SetDeviceCycleTime(deviceID int, ct float64) error {
	return service.SetDeviceCycleTime(deviceID, ct)
}

// GetCurrentShift 根据当前时刻返回匹配的活动班次（含休息时间段）。
// Returns the currently active shift matching the current wall-clock time.
// 現在時刻に一致するアクティブなシフトを返す。
func (a *App) GetCurrentShift() (*ShiftConfig, error) {
	return service.GetCurrentShift()
}

// GetShiftsForLogicalDay 返回当前逻辑日的所有活动班次列表。
// Returns the active shifts for the current logical day.
// 現在の論理日のアクティブなシフト一覧を返す。
func (a *App) GetShiftsForLogicalDay() ([]LogicalDayShift, error) {
	return service.GetShiftsForLogicalDay()
}

// GetDefaultShiftWindow 获取兜底的工作时间窗口字符串（用于 OEE SQL）。
// Returns the fallback work window for OEE SQL.
// OEE SQL 用のフォールバック作業時間帯を返す。
func (a *App) GetDefaultShiftWindow() (workStart, workEnd string) {
	return service.GetDefaultShiftWindow()
}
