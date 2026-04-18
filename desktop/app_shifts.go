package main

// ========================================================
// 时间安排组 & 班次配置管理接口（Wails 绑定）
// Shift Schedule & Shift Configuration Management (Wails Bindings)
// スケジュールグループ & シフト設定管理インターフェース（Wailsバインド）
//
// 两层数据模型（Two-level data model / 二層データモデル）：
//   SysShiftSchedule（时间安排组 / Schedule group）
//     └── SysShift[]（班次 / Shifts），每班含 SysShiftBreak[]（休息段 / Breaks）
//
// 设备关联规则（Device assignment rule）：
//   一台设备 → 一个时间安排组（多台设备可共用同一个组）
//   SysDevice.ScheduleID → SysShiftSchedule.ID
//
// 数据流：
//   前端 SystemSettings.vue
//     → GetShiftSchedules / SaveShiftSchedules（本文件）
//     → DB.Transaction → sys_shift_schedules + sys_shifts + sys_shift_breaks
//
//   OEE 大屏（Cockpit.vue）
//     → GetShiftsForLogicalDay（本文件）→ 跨所有组的全部活动班次（平铺）
//     → GetHourlyOEE（app.go）→ 按设备的 ScheduleID 分组查询
// ========================================================

import (
	"fmt"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"

	"gorm.io/gorm"
)

// deleteShiftCascade 手动删除单个班次及其全部休息段。
// CN: 运行时关闭了 GORM 自动外键迁移，不能依赖数据库级 ON DELETE CASCADE，必须在业务层显式删除子记录。
// EN: Foreign-key auto migration is disabled, so database-level ON DELETE CASCADE cannot be relied on here.
// JP: 外部キー自動マイグレーションが無効なため、DB の ON DELETE CASCADE に依存せず業務層で子レコードを明示削除する。
func deleteShiftCascade(tx *gorm.DB, shiftID int) error {
	if err := tx.Where("shift_id = ?", shiftID).Delete(&models.SysShiftBreak{}).Error; err != nil {
		return fmt.Errorf("删除班次休息段(shift_id=%d)失败: %w", shiftID, err)
	}
	if err := tx.Delete(&models.SysShift{}, shiftID).Error; err != nil {
		return fmt.Errorf("删除班次(id=%d)失败: %w", shiftID, err)
	}
	return nil
}

// deleteScheduleCascade 手动删除时间安排组及其全部班次/休息段，并先解除设备绑定。
// CN: 旧库没有真实外键时，直接删 sys_shift_schedules 会留下残留启用班次，后续逻辑日计算会误读这些脏数据。
// EN: On legacy DBs without actual foreign keys, deleting only the parent schedule leaves active orphan shifts behind.
// JP: 実外部キーがない旧 DB では親スケジュールだけ削除すると有効な孤立シフトが残り、後続計算が誤読する。
func deleteScheduleCascade(tx *gorm.DB, scheduleID int) error {
	var shiftIDs []int
	if err := tx.Model(&models.SysShift{}).Where("schedule_id = ?", scheduleID).Pluck("id", &shiftIDs).Error; err != nil {
		return fmt.Errorf("查询时间安排下属班次(id=%d)失败: %w", scheduleID, err)
	}
	for _, shiftID := range shiftIDs {
		if err := deleteShiftCascade(tx, shiftID); err != nil {
			return err
		}
	}
	if err := tx.Model(&models.SysDevice{}).Where("schedule_id = ?", scheduleID).Update("schedule_id", nil).Error; err != nil {
		return fmt.Errorf("解除设备与时间安排(id=%d)绑定失败: %w", scheduleID, err)
	}
	if err := tx.Delete(&models.SysShiftSchedule{}, scheduleID).Error; err != nil {
		return fmt.Errorf("删除时间安排(id=%d)失败: %w", scheduleID, err)
	}
	return nil
}

// cleanupOrphanShiftData 清理已经失去父时间安排/父班次的历史脏数据。
// CN: 这是对历史 bug 的止血，保证旧残留班次不会继续参与 GetShifts / GetCurrentShift / GetShiftsForLogicalDay。
// EN: This is a defensive cleanup for historical orphan rows so they no longer affect runtime shift resolution.
// JP: 過去バグで残った孤立行を防御的に清掃し、実行時のシフト判定へ流入させないための処置。
func cleanupOrphanShiftData(tx *gorm.DB) error {
	var orphanShiftIDs []int
	if err := tx.Model(&models.SysShift{}).
		Where("NOT EXISTS (SELECT 1 FROM sys_shift_schedules WHERE sys_shift_schedules.id = sys_shifts.schedule_id)").
		Pluck("id", &orphanShiftIDs).Error; err != nil {
		return fmt.Errorf("查询孤立班次失败: %w", err)
	}
	for _, shiftID := range orphanShiftIDs {
		if err := deleteShiftCascade(tx, shiftID); err != nil {
			return err
		}
	}
	if err := tx.Where("NOT EXISTS (SELECT 1 FROM sys_shifts WHERE sys_shifts.id = sys_shift_breaks.shift_id)").Delete(&models.SysShiftBreak{}).Error; err != nil {
		return fmt.Errorf("清理孤立休息段失败: %w", err)
	}
	return nil
}

// ─── Wails 前端可见的数据结构 ───────────────────────────────

// ShiftBreak 班次内的单个休息时间段（Wails 传输用）
// One break period within a shift (for Wails frontend transfer)
// 1シフト内の1つの休憩時間帯（Wailsフロント転送用）
type ShiftBreak struct {
	ID        int    `json:"id"`
	ShiftID   int    `json:"shift_id"`
	Name      string `json:"name"`
	StartHour int    `json:"start_hour"`
	StartMin  int    `json:"start_min"`
	EndHour   int    `json:"end_hour"`
	EndMin    int    `json:"end_min"`
}

// ShiftConfig 完整班次配置（含所有休息段，Wails 传输用）
// Full shift configuration including all break periods (for Wails frontend transfer)
// 休憩時間帯を含む完全なシフト設定（Wailsフロント転送用）
type ShiftConfig struct {
	ID         int          `json:"id"`
	ScheduleID int          `json:"schedule_id"` // 所属时间安排组 ID
	Name       string       `json:"name"`
	StartHour  int          `json:"start_hour"`
	StartMin   int          `json:"start_min"`
	EndHour    int          `json:"end_hour"`
	EndMin     int          `json:"end_min"`
	IsActive   bool         `json:"is_active"`
	SortOrder  int          `json:"sort_order"`
	Breaks     []ShiftBreak `json:"breaks"`
}

// ShiftScheduleConfig 时间安排组完整配置（含所有班次，Wails 传输用）
// Full schedule-group configuration including all shifts (for Wails frontend transfer)
// シフトグループの完全設定（全シフト含む、Wailsフロント転送用）
type ShiftScheduleConfig struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	SortOrder int           `json:"sort_order"`
	IsActive  bool          `json:"is_active"`
	Shifts    []ShiftConfig `json:"shifts"`
}

// ─── 辅助函数 ────────────────────────────────────────────

// modelToShiftConfig 将 DB 班次模型转换为前端结构
func modelToShiftConfig(s models.SysShift) ShiftConfig {
	breaks := make([]ShiftBreak, len(s.Breaks))
	for i, b := range s.Breaks {
		breaks[i] = ShiftBreak{
			ID:        b.ID,
			ShiftID:   b.ShiftID,
			Name:      b.Name,
			StartHour: int(b.StartHour),
			StartMin:  int(b.StartMin),
			EndHour:   int(b.EndHour),
			EndMin:    int(b.EndMin),
		}
	}
	return ShiftConfig{
		ID:         s.ID,
		ScheduleID: s.ScheduleID,
		Name:       s.Name,
		StartHour:  int(s.StartHour),
		StartMin:   int(s.StartMin),
		EndHour:    int(s.EndHour),
		EndMin:     int(s.EndMin),
		IsActive:   s.IsActive,
		SortOrder:  s.SortOrder,
		Breaks:     breaks,
	}
}

// modelToScheduleConfig 将 DB 时间安排组模型转换为前端结构
func modelToScheduleConfig(sc models.SysShiftSchedule) ShiftScheduleConfig {
	shifts := make([]ShiftConfig, len(sc.Shifts))
	for i, s := range sc.Shifts {
		shifts[i] = modelToShiftConfig(s)
	}
	return ShiftScheduleConfig{
		ID:        sc.ID,
		Name:      sc.Name,
		SortOrder: sc.SortOrder,
		IsActive:  sc.IsActive,
		Shifts:    shifts,
	}
}

// validateShiftConfig 校验班次配置合理性（供保存前使用）
func validateShiftConfig(sc ShiftConfig) error {
	if sc.Name == "" {
		return fmt.Errorf("班次名称不能为空")
	}
	if sc.StartHour < 0 || sc.StartHour > 23 || sc.EndHour < 0 || sc.EndHour > 23 {
		return fmt.Errorf("班次%q：小时必须在0-23之间", sc.Name)
	}
	if sc.StartMin < 0 || sc.StartMin > 59 || sc.EndMin < 0 || sc.EndMin > 59 {
		return fmt.Errorf("班次%q：分钟必须在0-59之间", sc.Name)
	}
	// 跨零点班次合法（如 22:00→06:00）；只拒绝开始与结束完全相同的零时长情况
	if sc.StartHour == sc.EndHour && sc.StartMin == sc.EndMin {
		return fmt.Errorf("班次%q：开始与结束时间不能相同", sc.Name)
	}
	for _, b := range sc.Breaks {
		if b.Name == "" {
			return fmt.Errorf("班次%q中有休息时间段名称为空", sc.Name)
		}
		if b.StartHour < 0 || b.StartHour > 23 || b.EndHour < 0 || b.EndHour > 23 {
			return fmt.Errorf("班次%q - 休息%q：小时必须在0-23之间", sc.Name, b.Name)
		}
		if b.StartMin < 0 || b.StartMin > 59 || b.EndMin < 0 || b.EndMin > 59 {
			return fmt.Errorf("班次%q - 休息%q：分钟必须在0-59之间", sc.Name, b.Name)
		}
		// CN: 允许跨零点的休息段（如 23:00→01:00），只禁止起止时间完全相同（零时长）。
		//     跨零点时实际持续时长 = (EndHour+24)*60+EndMin - (StartHour*60+StartMin)
		// EN: Cross-midnight breaks (e.g., 23:00→01:00) are allowed; only zero-duration is rejected.
		// JP: 深夜跨ぎ休憩（23:00→01:00 等）を許容し、開始＝終了の零時長のみ拒否する。
		if b.StartHour == b.EndHour && b.StartMin == b.EndMin {
			return fmt.Errorf("班次%q - 休息%q：开始与结束时间不能相同", sc.Name, b.Name)
		}
	}
	return nil
}

// ─── Wails 绑定方法 ──────────────────────────────────────

// GetShiftSchedules 获取所有时间安排组（含各组的班次和休息段）
// Returns all schedule groups with their shifts and break periods.
// 全スケジュールグループをシフト・休憩時間帯リストごと返す。
func (a *App) GetShiftSchedules() ([]ShiftScheduleConfig, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	var schedules []models.SysShiftSchedule
	if err := database.DB.
		Preload("Shifts", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Shifts.Breaks").
		Order("sort_order ASC, id ASC").
		Find(&schedules).Error; err != nil {
		return nil, fmt.Errorf("查询时间安排组失败: %w", err)
	}
	result := make([]ShiftScheduleConfig, len(schedules))
	for i, sc := range schedules {
		result[i] = modelToScheduleConfig(sc)
	}
	return result, nil
}

// GetShifts 获取所有活动班次（平铺，跨所有时间安排组；供内部逻辑日计算使用）
// Returns all active shifts flattened across all schedules (for logical-day computation).
// 全スケジュールの全アクティブシフトをフラット化して返す（論理日計算用）。
func (a *App) GetShifts() ([]ShiftConfig, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	var shifts []models.SysShift
	if err := database.DB.
		Joins("INNER JOIN sys_shift_schedules ON sys_shift_schedules.id = sys_shifts.schedule_id").
		Where("sys_shift_schedules.is_active = ?", true).
		Preload("Breaks").
		Order("sys_shift_schedules.sort_order ASC, sys_shifts.sort_order ASC, sys_shifts.id ASC").
		Find(&shifts).Error; err != nil {
		return nil, fmt.Errorf("查询班次配置失败: %w", err)
	}
	result := make([]ShiftConfig, len(shifts))
	for i, s := range shifts {
		result[i] = modelToShiftConfig(s)
	}
	return result, nil
}

// SaveShiftSchedules 保存全部时间安排组（upsert 语义）
//
// CN: 对每个 Schedule：ID > 0 → UPDATE；ID = 0 → INSERT；DB 有但前端未传 → DELETE。
//
//	DELETE 时先将关联设备的 schedule_id 置 NULL，再删除 schedule（CASCADE 清理 shifts/breaks）。
//	Shift 级别：每个 schedule 内的 shifts 同样做 upsert，break 做全量替换。
//
// EN: Upsert schedules & their shifts; deleted schedules first NULL-out device.schedule_id, then CASCADE.
// JP: スケジュールと配下シフトを upsert。削除時は設備の schedule_id を NULL にしてから CASCADE 削除。
func (a *App) SaveShiftSchedules(schedules []ShiftScheduleConfig) error {
	if database.DB == nil {
		return fmt.Errorf("数据库未连接")
	}
	// 前置校验
	for si, sched := range schedules {
		if sched.Name == "" {
			return fmt.Errorf("第%d个时间安排名称不能为空", si+1)
		}
		for i, sc := range sched.Shifts {
			if err := validateShiftConfig(sc); err != nil {
				return fmt.Errorf("时间安排%q - 第%d条班次校验失败: %w", sched.Name, i+1, err)
			}
		}
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := cleanupOrphanShiftData(tx); err != nil {
			return err
		}

		// ── 1. Schedule 层 upsert ────────────────────────────
		var existingSchedIDs []int
		if err := tx.Model(&models.SysShiftSchedule{}).Pluck("id", &existingSchedIDs).Error; err != nil {
			return fmt.Errorf("查询现有时间安排失败: %w", err)
		}
		keepSchedSet := map[int]bool{}
		for _, s := range schedules {
			if s.ID > 0 {
				keepSchedSet[s.ID] = true
			}
		}
		for _, eid := range existingSchedIDs {
			if !keepSchedSet[eid] {
				if err := deleteScheduleCascade(tx, eid); err != nil {
					return err
				}
			}
		}

		for si, sched := range schedules {
			var schedID int
			if sched.ID > 0 {
				schedID = sched.ID
				if err := tx.Model(&models.SysShiftSchedule{}).Where("id = ?", sched.ID).Updates(map[string]interface{}{
					"name":       sched.Name,
					"sort_order": si,
					"is_active":  sched.IsActive,
				}).Error; err != nil {
					return fmt.Errorf("更新时间安排%q失败: %w", sched.Name, err)
				}
			} else {
				row := models.SysShiftSchedule{Name: sched.Name, SortOrder: si, IsActive: sched.IsActive}
				if err := tx.Create(&row).Error; err != nil {
					return fmt.Errorf("新建时间安排%q失败: %w", sched.Name, err)
				}
				schedID = row.ID
			}

			// ── 2. Shift 层 upsert（同一 schedule 内）────────
			var existingShiftIDs []int
			if err := tx.Model(&models.SysShift{}).Where("schedule_id = ?", schedID).Pluck("id", &existingShiftIDs).Error; err != nil {
				return fmt.Errorf("查询时间安排%q的现有班次失败: %w", sched.Name, err)
			}
			keepShiftSet := map[int]bool{}
			for _, sc := range sched.Shifts {
				if sc.ID > 0 {
					keepShiftSet[sc.ID] = true
				}
			}
			for _, eid := range existingShiftIDs {
				if !keepShiftSet[eid] {
					if err := deleteShiftCascade(tx, eid); err != nil {
						return err
					}
				}
			}

			for idx, sc := range sched.Shifts {
				// CN: 班次启用状态完全跟随父级时间安排：父启用 → 子全部启用；父停用 → 子全部停用
				// EN: Shift active state fully mirrors the parent schedule (parent on → all children on; off → all off)
				// JP: シフトの有効状態は親スケジュールに完全追従（親 ON → 子全 ON、親 OFF → 子全 OFF）
				effectiveActive := sched.IsActive
				_ = sc.IsActive // 显式保留字段引用，未来若恢复独立控制可去掉

				var shiftID int
				if sc.ID > 0 {
					shiftID = sc.ID
					if err := tx.Model(&models.SysShift{}).Where("id = ?", sc.ID).Updates(map[string]interface{}{
						"schedule_id": schedID,
						"name":        sc.Name,
						"start_hour":  int8(sc.StartHour),
						"start_min":   int8(sc.StartMin),
						"end_hour":    int8(sc.EndHour),
						"end_min":     int8(sc.EndMin),
						"is_active":   effectiveActive,
						"sort_order":  idx,
					}).Error; err != nil {
						return fmt.Errorf("更新班次%q失败: %w", sc.Name, err)
					}
					if err := tx.Where("shift_id = ?", sc.ID).Delete(&models.SysShiftBreak{}).Error; err != nil {
						return fmt.Errorf("清理班次%q旧休息段失败: %w", sc.Name, err)
					}
				} else {
					row := models.SysShift{
						ScheduleID: schedID,
						Name:       sc.Name,
						StartHour:  int8(sc.StartHour),
						StartMin:   int8(sc.StartMin),
						EndHour:    int8(sc.EndHour),
						EndMin:     int8(sc.EndMin),
						IsActive:   effectiveActive,
						SortOrder:  idx,
					}
					if err := tx.Create(&row).Error; err != nil {
						return fmt.Errorf("新建班次%q失败: %w", sc.Name, err)
					}
					shiftID = row.ID
				}

				// ── 3. Break 层全量替换 ──────────────────────
				for _, bc := range sc.Breaks {
					b := models.SysShiftBreak{
						ShiftID:   shiftID,
						Name:      bc.Name,
						StartHour: int8(bc.StartHour),
						StartMin:  int8(bc.StartMin),
						EndHour:   int8(bc.EndHour),
						EndMin:    int8(bc.EndMin),
					}
					if err := tx.Create(&b).Error; err != nil {
						return fmt.Errorf("班次%q - 写入休息段%q失败: %w", sc.Name, bc.Name, err)
					}
				}
			}
		}
		return nil
	})
}

// ─── 设备-时间安排组关联接口 ─────────────────────────────

// GetScheduleDeviceIDs 获取指定时间安排组关联的所有设备 ID 列表
// Returns device IDs assigned to the given schedule group.
// 指定スケジュールグループに紐づく設備 ID 一覧を返す。
func (a *App) GetScheduleDeviceIDs(scheduleID int) ([]int, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("数据库未连接")
	}
	var ids []int
	err := database.DB.Model(&models.SysDevice{}).
		Where("schedule_id = ?", scheduleID).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, fmt.Errorf("查询时间安排组设备列表失败: %w", err)
	}
	return ids, nil
}

// SetDeviceSchedule 设置设备关联的时间安排组（scheduleID <= 0 表示解除绑定）
//
// CN: 一台设备至多绑定一个时间安排组；scheduleID <= 0 时置 NULL（解除绑定）。
// EN: At most one schedule per device; scheduleID <= 0 clears binding (schedule_id = NULL).
// JP: 1設備に最大1スケジュール。scheduleID <= 0 でバインド解除（schedule_id = NULL）。
func (a *App) SetDeviceSchedule(deviceID int, scheduleID int) error {
	if database.DB == nil {
		return fmt.Errorf("数据库未连接")
	}
	var val interface{}
	if scheduleID > 0 {
		val = scheduleID
	}
	return database.DB.Model(&models.SysDevice{}).
		Where("id = ?", deviceID).
		Update("schedule_id", val).Error
}

// GetDeviceCycleTime 获取指定设备的理论节拍（CT，秒/件）。
// 若设备未单独配置则返回 0（调用方应 fallback 到全局 ProductionCoefficient）。
// Returns the device-level cycle time; returns 0 if not set (caller should fall back to global CT).
// 指定設備のサイクルタイム（秒/個）を返す。未設定は 0 を返す（呼び出し元でグローバル CT へフォールバック）。
func (a *App) GetDeviceCycleTime(deviceID int) (float64, error) {
	if database.DB == nil {
		return 0, fmt.Errorf("数据库未连接")
	}
	var dev models.SysDevice
	if err := database.DB.Select("id, cycle_time").First(&dev, deviceID).Error; err != nil {
		return 0, fmt.Errorf("查询设备(id=%d)失败: %w", deviceID, err)
	}
	if dev.CycleTime != nil && *dev.CycleTime > 0 {
		return *dev.CycleTime, nil
	}
	return 0, nil
}

// SetDeviceCycleTime 设置设备的理论节拍（CT，秒/件）。ct <= 0 时置 NULL（使用全局默认值）。
// Set device-level cycle time (seconds per piece). ct <= 0 clears to NULL (use global default).
// デバイスのサイクルタイム（秒/個）を設定。ct <= 0 で NULL（グローバルデフォルト使用）。
func (a *App) SetDeviceCycleTime(deviceID int, ct float64) error {
	if database.DB == nil {
		return fmt.Errorf("数据库未连接")
	}
	var val interface{}
	if ct > 0 {
		val = ct
	}
	return database.DB.Model(&models.SysDevice{}).
		Where("id = ?", deviceID).
		Update("cycle_time", val).Error
}

// GetCurrentShift 根据当前时刻返回匹配的活动班次（含休息时间段）
// 匹配规则：now 落在 [shift.start, shift.end) 范围内，且 is_active=true。
// 若有多个匹配，返回 sort_order 最小的那个。
// 若无任何匹配，返回 is_active=true 中 sort_order 最小的那个（兜底）。
// 若没有任何活动班次，返回 nil, nil。
//
// Returns the currently active shift matching the current wall-clock time.
// 現在時刻に一致するアクティブなシフトを返す（休憩時間帯付き）。
func (a *App) GetCurrentShift() (*ShiftConfig, error) {
	shifts, err := a.GetShifts()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	nowTotalMin := now.Hour()*60 + now.Minute()

	// 第一轮：找精确时间匹配的活动班次
	for _, s := range shifts {
		if !s.IsActive {
			continue
		}
		startMin := s.StartHour*60 + s.StartMin
		endMin := s.EndHour*60 + s.EndMin
		if nowTotalMin >= startMin && nowTotalMin < endMin {
			result := s
			return &result, nil
		}
	}

	// 兜底：返回第一个活动班次（sort_order 升序，GetShifts 已排好序）
	for _, s := range shifts {
		if s.IsActive {
			result := s
			return &result, nil
		}
	}

	return nil, nil
}

// ─── 逻辑日班次接口 ──────────────────────────────────────

// LogicalDayShift 逻辑日内的单个班次信息（含已到达标记）
// One shift in the logical day with arrival state.
// 論理日内の1シフト情報（到達フラグ付き）。
type LogicalDayShift struct {
	ShiftConfig
	HasArrived        bool   `json:"has_arrived"`         // 当前时刻已到达该班次开始时间
	IsCurrent         bool   `json:"is_current"`          // 当前时刻正处于该班次时间窗口内
	LogicalDate       string `json:"logical_date"`        // "YYYY-MM-DD"，该班次所属的逻辑日期
	CalendarDayOffset int    `json:"calendar_day_offset"` // 0=与逻辑日同自然日；1=逻辑日+1自然日（如三班 0:00-7:40）
	// EN: 0=same calendar day as logical date; 1=logical date +1 day (e.g., 三班 running past midnight).
	// JP: 0=論理日と同自然日；1=論理日+1日（例：三班 0:00-7:40 は翌自然日に運行）。
}

// GetShiftsForLogicalDay 返回当前逻辑日的所有活动班次列表
//
// 逻辑日边界 = 5:00 之后的第一个活动班次的开始时刻：
//   - 当前时刻 >= 第一班开始时刻 → 逻辑日为今天
//   - 当前时刻 < 第一班开始时刻 → 逻辑日为昨天（显示昨天同一时段的数据）
//
// Returns the active shifts for the current logical day with arrival markers.
// 現在の論理日のアクティブなシフト一覧を到達フラグ付きで返す。
func (a *App) GetShiftsForLogicalDay() ([]LogicalDayShift, error) {
	shifts, err := a.GetShifts()
	if err != nil {
		return nil, err
	}

	// 只取 is_active=true 的班次（已按 sort_order 升序排列）
	active := make([]ShiftConfig, 0, len(shifts))
	for _, s := range shifts {
		if s.IsActive {
			active = append(active, s)
		}
	}
	if len(active) == 0 {
		return []LogicalDayShift{}, nil
	}

	// 逻辑日锚点 = sort_order 最小的活动班次（已按升序排列，取首元素）
	// CN: 不依赖硬编码时刻；以最早班次的开始时间作为逻辑日分界。
	// EN: Use the lowest-sort_order active shift start as the logical-day boundary; no hardcoded hour.
	// JP: sort_order 最小シフトの開始時刻を論理日境界とし、ハードコード時刻に依存しない。
	anchorShift := &active[0]
	logicalBoundaryMin := anchorShift.StartHour*60 + anchorShift.StartMin

	now := time.Now()
	nowMin := now.Hour()*60 + now.Minute()
	anchorStartMin := logicalBoundaryMin

	// 判断逻辑日属于今天还是昨天
	var logicalBase time.Time
	if nowMin >= anchorStartMin {
		logicalBase = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	} else {
		logicalBase = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	}
	logicalDateStr := logicalBase.Format("2006-01-02")

	result := make([]LogicalDayShift, len(active))
	for i, s := range active {
		startMin := s.StartHour*60 + s.StartMin
		endMin := s.EndHour*60 + s.EndMin

		// calendarDayOffset: 班次开始时间早于逻辑日锚点 → 该班次运行在逻辑日+1自然日（如三班 0:00-7:40）
		// EN: If shift starts before the logical boundary, it runs on the next calendar day.
		// JP: シフト開始が論理日境界より早い場合、そのシフトは翌自然日に運行する。
		calendarDayOffset := 0
		if startMin < logicalBoundaryMin {
			calendarDayOffset = 1
		}
		calendarBase := logicalBase.AddDate(0, 0, calendarDayOffset)

		shiftStart := calendarBase.Add(time.Duration(startMin) * time.Minute)
		var shiftEnd time.Time
		if endMin > startMin {
			shiftEnd = calendarBase.Add(time.Duration(endMin) * time.Minute)
		} else {
			// 跨零点班次（如 22:00→06:00）
			shiftEnd = calendarBase.Add(time.Duration(endMin+24*60) * time.Minute)
		}

		hasArrived := !now.Before(shiftStart)
		isCurrent := hasArrived && now.Before(shiftEnd)

		result[i] = LogicalDayShift{
			ShiftConfig:       s,
			HasArrived:        hasArrived,
			IsCurrent:         isCurrent,
			LogicalDate:       logicalDateStr,
			CalendarDayOffset: calendarDayOffset,
		}
	}
	return result, nil
}

// GetDefaultShiftWindow 获取兜底的工作时间窗口字符串（用于 OEE SQL）
// 优先取当前活动班次的时间窗口；没有配置任何班次时降级为历史默认值 07:40-16:20。
// Returns the work window (start/end as "HH:MM") for OEE calculation.
// OEE計算用の作業時間ウィンドウ（"HH:MM"形式）を返す。
func (a *App) GetDefaultShiftWindow() (workStart, workEnd string) {
	shift, err := a.GetCurrentShift()
	if err != nil || shift == nil {
		// 没有班次配置时使用历史兜底值
		return "07:40", "16:20"
	}
	return fmt.Sprintf("%02d:%02d", shift.StartHour, shift.StartMin),
		fmt.Sprintf("%02d:%02d", shift.EndHour, shift.EndMin)
}
