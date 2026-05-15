package service

import (
	"fmt"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"

	"gorm.io/gorm"
)

// ShiftBreak 班次内的单个休息时间段。
// CN: 这是 Wails 与后续 Screen API 共享的班次传输结构，来源是 sys_shift_breaks。
// EN: Shared shift-break DTO for Wails and future Screen APIs, backed by sys_shift_breaks.
// JP: Wails と将来の Screen API が共有する休憩 DTO。データ元は sys_shift_breaks。
type ShiftBreak struct {
	ID        int    `json:"id"`
	ShiftID   int    `json:"shift_id"`
	Name      string `json:"name"`
	StartHour int    `json:"start_hour"`
	StartMin  int    `json:"start_min"`
	EndHour   int    `json:"end_hour"`
	EndMin    int    `json:"end_min"`
}

// ShiftConfig 完整班次配置（含所有休息段）。
// CN: 前端保存班次、OEE 逻辑日计算、快照生成都复用同一个 DTO，避免多个口径。
// EN: One DTO is reused by schedule editing, OEE logical-day calculation, and snapshot generation.
// JP: スケジュール編集、OEE 論理日計算、スナップショット生成で同一 DTO を再利用する。
type ShiftConfig struct {
	ID         int          `json:"id"`
	ScheduleID int          `json:"schedule_id"`
	Name       string       `json:"name"`
	StartHour  int          `json:"start_hour"`
	StartMin   int          `json:"start_min"`
	EndHour    int          `json:"end_hour"`
	EndMin     int          `json:"end_min"`
	IsActive   bool         `json:"is_active"`
	SortOrder  int          `json:"sort_order"`
	Breaks     []ShiftBreak `json:"breaks"`
}

// ShiftScheduleConfig 时间安排组完整配置（含所有班次）。
// CN: 设备绑定的是时间安排组，不是单个班次；保存时按组做 upsert。
// EN: Devices bind to schedule groups rather than individual shifts; save uses group-level upsert.
// JP: 設備は単一シフトではなくスケジュールグループへ紐づく。保存はグループ単位の upsert。
type ShiftScheduleConfig struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	SortOrder int           `json:"sort_order"`
	IsActive  bool          `json:"is_active"`
	Shifts    []ShiftConfig `json:"shifts"`
}

// LogicalDayShift 逻辑日内的单个班次信息（含已到达标记）。
// CN: CalendarDayOffset 是 OEE/快照跨零点正确性的关键字段，不能退回硬编码 5 点边界。
// EN: CalendarDayOffset is required for correct cross-midnight OEE/snapshot windows; do not restore a hardcoded 5AM boundary.
// JP: CalendarDayOffset は深夜跨ぎの OEE/スナップショット窓に必須。5時固定境界へ戻さないこと。
type LogicalDayShift struct {
	ShiftConfig
	HasArrived        bool   `json:"has_arrived"`
	IsCurrent         bool   `json:"is_current"`
	LogicalDate       string `json:"logical_date"`
	CalendarDayOffset int    `json:"calendar_day_offset"`
}

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
// CN: 这是对历史 bug 的止血，保证旧残留班次不会继续参与逻辑日和当前班次判断。
// EN: Defensive cleanup for historical orphan rows so they no longer affect runtime shift resolution.
// JP: 過去バグで残った孤立行を防御的に清掃し、実行時のシフト判定へ流入させない。
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

// ModelToShiftConfig 将 DB 班次模型转换为共享 DTO。
func ModelToShiftConfig(s models.SysShift) ShiftConfig {
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

func modelToScheduleConfig(sc models.SysShiftSchedule) ShiftScheduleConfig {
	shifts := make([]ShiftConfig, len(sc.Shifts))
	for i, s := range sc.Shifts {
		shifts[i] = ModelToShiftConfig(s)
	}
	return ShiftScheduleConfig{
		ID:        sc.ID,
		Name:      sc.Name,
		SortOrder: sc.SortOrder,
		IsActive:  sc.IsActive,
		Shifts:    shifts,
	}
}

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
		if b.StartHour == b.EndHour && b.StartMin == b.EndMin {
			return fmt.Errorf("班次%q - 休息%q：开始与结束时间不能相同", sc.Name, b.Name)
		}
	}
	return nil
}

// GetShiftSchedules 获取所有时间安排组（含各组的班次和休息段）。
func GetShiftSchedules() ([]ShiftScheduleConfig, error) {
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

// GetShifts 获取所有活动班次（平铺，跨所有时间安排组；供内部逻辑日计算使用）。
func GetShifts() ([]ShiftConfig, error) {
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
		result[i] = ModelToShiftConfig(s)
	}
	return result, nil
}

// SaveShiftSchedules 保存全部时间安排组（upsert 语义）。
// CN: 前端传来的完整树是正本，DB 有但前端未传的 schedule/shift 会删除；break 每次全量替换。
// EN: The frontend tree is authoritative; missing schedules/shifts are deleted and breaks are fully replaced.
// JP: フロントから来るツリーを正本とし、欠落した schedule/shift は削除、break は毎回全置換する。
func SaveShiftSchedules(schedules []ShiftScheduleConfig) error {
	if database.DB == nil {
		return fmt.Errorf("数据库未连接")
	}
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
				effectiveActive := sched.IsActive
				_ = sc.IsActive

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

// GetScheduleDeviceIDs 获取指定时间安排组关联的所有设备 ID 列表。
func GetScheduleDeviceIDs(scheduleID int) ([]int, error) {
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

// SetDeviceSchedule 设置设备关联的时间安排组（scheduleID <= 0 表示解除绑定）。
func SetDeviceSchedule(deviceID int, scheduleID int) error {
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
func GetDeviceCycleTime(deviceID int) (float64, error) {
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
func SetDeviceCycleTime(deviceID int, ct float64) error {
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

// GetCurrentShift 根据当前时刻返回匹配的活动班次（含休息时间段）。
func GetCurrentShift() (*ShiftConfig, error) {
	shifts, err := GetShifts()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	nowTotalMin := now.Hour()*60 + now.Minute()

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

	for _, s := range shifts {
		if s.IsActive {
			result := s
			return &result, nil
		}
	}

	return nil, nil
}

// GetShiftsForLogicalDay 返回当前逻辑日的所有活动班次列表。
// CN: 逻辑日边界由 sort_order 最小的活动班次决定；早于该边界的班次属于逻辑日+1自然日。
// EN: The logical boundary is the active shift with the lowest sort_order; shifts earlier than it run on logical date +1.
// JP: 論理日境界は sort_order 最小の有効シフト。境界より早いシフトは論理日+1自然日に属する。
func GetShiftsForLogicalDay() ([]LogicalDayShift, error) {
	shifts, err := GetShifts()
	if err != nil {
		return nil, err
	}

	active := make([]ShiftConfig, 0, len(shifts))
	for _, s := range shifts {
		if s.IsActive {
			active = append(active, s)
		}
	}
	if len(active) == 0 {
		return []LogicalDayShift{}, nil
	}

	anchorShift := &active[0]
	logicalBoundaryMin := anchorShift.StartHour*60 + anchorShift.StartMin

	now := time.Now()
	nowMin := now.Hour()*60 + now.Minute()

	var logicalBase time.Time
	if nowMin >= logicalBoundaryMin {
		logicalBase = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	} else {
		logicalBase = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	}
	logicalDateStr := logicalBase.Format("2006-01-02")

	result := make([]LogicalDayShift, len(active))
	for i, s := range active {
		startMin := s.StartHour*60 + s.StartMin
		endMin := s.EndHour*60 + s.EndMin

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

// GetDefaultShiftWindow 获取兜底的工作时间窗口字符串（用于 OEE SQL）。
func GetDefaultShiftWindow() (workStart, workEnd string) {
	shift, err := GetCurrentShift()
	if err != nil || shift == nil {
		return "07:40", "16:20"
	}
	return fmt.Sprintf("%02d:%02d", shift.StartHour, shift.StartMin),
		fmt.Sprintf("%02d:%02d", shift.EndHour, shift.EndMin)
}
