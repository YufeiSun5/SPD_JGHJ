// ============================================================================
// 测点内存模型 (Tag) - 系统最核心的数据结构 ⭐⭐⭐
// ============================================================================
// 职责: 维护测点的配置和运行时状态
// 线程安全: sync.RWMutex 保护并发读写
//
// 核心字段:
//
//	基础配置: VarID, VarName, DataType, JSONPath (只读,初始化后不变)
//	运行时状态: CurrentValue, LastValue, LastUpdateTime (需加锁)
//	报警配置: AlarmEnable, LimitHH/H/L/LL, Deadband (只读)
//	存储配置: StoreMode, StoreCycle, StoreDeadband (只读)
//
// 关键方法:
//
//	UpdateValue(): 线程安全更新值, 返回是否变化
//	ShouldStore(): 判断变动存储 (检查死区)
//	ShouldCycleStore(): 判断定时存储 (检查周期)
//	CheckAlarm(): 检查报警状态 (返回报警类型和阈值)
//
// 存储模式 (StoreMode):
//
//	0 - 不存储: 临时计算变量
//	1 - 变动存储: 值变化超过 store_deadband 时存储
//	2 - 定时存储: 每隔 store_cycle 秒存储一次
//	3 - 混合存储: 变动存储 + 定时存储 (推荐)
//
// 何时修改此文件:
//   - 需要支持新的数据类型
//   - 需要修改存储判断逻辑
//   - 需要添加新的报警类型
//
// ============================================================================
package models

import (
	"math"
	"sync"
	"time"
)

// Tag 测点内存模型 - 核心数据结构
type Tag struct {
	// 基础配置 (只读,初始化后不变)
	VarID       int64   `json:"var_id"`
	VarName     string  `json:"var_name"`
	DisplayName string  `json:"display_name"`
	DataType    string  `json:"data_type"` // BOOL, INT16, INT32, INT64, FLOAT, DOUBLE, STRING
	RWMode      string  `json:"rw_mode"`   // R, W, RW
	Unit        string  `json:"unit"`
	JSONPath    string  `json:"json_path"`
	ScaleFactor float64 `json:"scale_factor"`
	OffsetVal   float64 `json:"offset_val"`

	// 报警配置 (只读)
	AlarmEnable bool    `json:"alarm_enable"`
	LimitHH     float64 `json:"limit_hh"`
	LimitH      float64 `json:"limit_h"`
	LimitL      float64 `json:"limit_l"`
	LimitLL     float64 `json:"limit_ll"`
	Deadband    float64 `json:"deadband"`
	AlarmMsg    string  `json:"alarm_msg"`

	// 存储配置 (只读)
	StoreMode     int     `json:"store_mode"`     // 0-不存, 1-变化, 2-定时, 3-混合
	StoreCycle    int     `json:"store_cycle"`    // 定时周期(秒)
	StoreDeadband float64 `json:"store_deadband"` // 存储死区

	// 采集防抖/启动快照配置 (只读)
	SuspiciousValue       *float64 `json:"suspicious_value"`        // NULL=关闭采集防抖
	DebounceThreshold     float64  `json:"debounce_threshold"`      // LastValidValue 大于该值才拦截可疑值
	StartupSnapshotEnable *int     `json:"startup_snapshot_enable"` // NULL=兼容旧行为, 1=开启, 0=关闭

	// 运行时状态 (需加锁修改)
	mu                     sync.RWMutex
	CurrentValue           float64   `json:"current_value"`
	LastValue              float64   `json:"last_value"`
	CurrentStrValue        string    `json:"current_str_value"` // 字符串类型的值
	LastStrValue           string    `json:"last_str_value"`    // 上次字符串值
	LastUpdateTime         time.Time `json:"last_update_time"`
	LastStoreTime          time.Time `json:"last_store_time"`
	IsFirstUpdate          bool      `json:"is_first_update"`           // 是否首次更新 (用于过滤冷启动)
	Quality                int       `json:"quality"`                   // 数据质量/连接状态: 1=在线(192), 0=离线(非192)
	LastQuality            int       `json:"last_quality"`              // 上次质量码/连接状态 (用于检测在线/离线变化)
	DebouncePending        bool      `json:"debounce_pending"`          // 是否存在等待判定的可疑值
	DebounceValue          float64   `json:"debounce_value"`            // 等待判定的可疑值
	DebounceTime           time.Time `json:"debounce_time"`             // 可疑值首次进入观察窗的时间
	DebounceQuality        int       `json:"debounce_quality"`          // 可疑值对应的数据质量
	DebounceLastValidValue *float64  `json:"debounce_last_valid_value"` // 启动前历史最近非零值，用于首帧可疑值判定

	// 报警状态
	AlarmState     string    `json:"alarm_state"`      // "", "HH", "H", "L", "LL"
	AlarmStartTime time.Time `json:"alarm_start_time"` // 报警开始时间
	AlarmRecordID  int64     `json:"alarm_record_id"`  // 当前报警记录ID
}

// ValueChange 描述一次被采集层确认后的数值变化。
type ValueChange struct {
	OldValue  float64
	NewValue  float64
	Timestamp time.Time
}

// GetValue 线程安全获取当前值
func (t *Tag) GetValue() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.CurrentValue
}

// GetLastUpdateTime 线程安全获取最后更新时间
func (t *Tag) GetLastUpdateTime() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.LastUpdateTime
}

// UpdateValue 线程安全更新值 - 返回是否变动 (数值类型)
// quality: 数据质量/连接状态 (1=在线, 0=离线)
func (t *Tag) UpdateValue(newValue float64, updateTime time.Time, quality int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 首次更新标记：系统启动后第一次收到数据
	if t.IsFirstUpdate {
		// 首次更新：直接赋值，不算作"变化"
		t.CurrentValue = newValue
		t.LastValue = newValue // 避免下次计算出错误的change
		t.LastUpdateTime = updateTime
		t.Quality = quality     // 更新质量戳/连接状态
		t.LastQuality = quality // 初始化上次质量码
		t.IsFirstUpdate = false
		return false // 不触发任务
	}

	changed := false
	if t.CurrentValue != newValue {
		t.LastValue = t.CurrentValue
		t.CurrentValue = newValue
		changed = true
	}

	// 🔥 更新质量码/连接状态
	t.LastQuality = t.Quality
	t.Quality = quality

	t.LastUpdateTime = updateTime
	return changed
}

// ApplyNumericSample 在线程安全边界内处理数值采样，并返回需要进入业务分发的真实变化。
// CN: 可疑值观察窗必须在更新 CurrentValue 前执行，假 0 不能污染内存、任务触发和历史入库。
// EN: The suspicious-value window runs before CurrentValue changes, so a false zero cannot pollute memory, tasks, or history.
// JP: 疑わしい値の観察窓は CurrentValue 更新前に実行し、偽の 0 がメモリ・タスク・履歴を汚さないようにする。
func (t *Tag) ApplyNumericSample(newValue float64, updateTime time.Time, quality int) []ValueChange {
	t.mu.Lock()
	defer t.mu.Unlock()

	suspiciousConfigured := t.SuspiciousValue != nil
	isSuspicious := suspiciousConfigured && floatsEqual(newValue, *t.SuspiciousValue)

	if t.IsFirstUpdate {
		if suspiciousConfigured && t.DebouncePending && !isSuspicious {
			referenceValue, hasReference := t.startupDebounceReferenceLocked()
			if hasReference {
				return t.resolveStartupDebounceLocked(newValue, updateTime, quality, referenceValue)
			}
			t.clearDebounceLocked()
		}

		if isSuspicious {
			referenceValue, hasReference := t.startupDebounceReferenceLocked()
			if hasReference && referenceValue > t.DebounceThreshold {
				// CN: 启动首帧可疑值必须先观察，不能被 InitOnly 当快照写入历史表。
				// EN: A suspicious first frame must be observed first; InitOnly must not store it as a startup snapshot.
				// JP: 起動初回の疑わしい値は先に観察し、InitOnly で履歴に保存しない。
				t.DebouncePending = true
				t.DebounceValue = newValue
				t.DebounceTime = updateTime
				t.DebounceQuality = quality
				t.LastQuality = t.Quality
				t.Quality = quality
				return nil
			}
		}

		t.CurrentValue = newValue
		t.LastValue = newValue
		t.LastUpdateTime = updateTime
		t.Quality = quality
		t.LastQuality = quality
		t.IsFirstUpdate = false
		t.clearDebounceLocked()
		return nil
	}

	if t.SuspiciousValue == nil {
		return t.applyNumericValueLocked(newValue, updateTime, quality)
	}

	lastValid := t.CurrentValue

	if isSuspicious {
		if lastValid > t.DebounceThreshold {
			// CN: 大计数值后的可疑值先观察，不落库；连续可疑值只保留一笔，避免日志放大。
			// EN: After a large counter value, hold the suspicious value in memory only; repeated suspicious samples stay collapsed.
			// JP: 大きなカウンタ値後の疑わしい値はメモリ観察のみとし、連続値は1件に畳み込む。
			if !t.DebouncePending {
				t.DebounceTime = updateTime
			}
			t.DebouncePending = true
			t.DebounceValue = newValue
			t.DebounceQuality = quality
			t.LastQuality = t.Quality
			t.Quality = quality
			t.LastUpdateTime = updateTime
			return nil
		}

		t.clearDebounceLocked()
		return t.applyNumericValueLocked(newValue, updateTime, quality)
	}

	if t.DebouncePending {
		pendingValue := t.DebounceValue
		pendingTime := t.DebounceTime
		pendingQuality := t.DebounceQuality
		t.clearDebounceLocked()

		if newValue >= lastValid {
			// CN: 恢复值不低于 LastValidValue，说明中间可疑值是假回落，直接丢弃。
			// EN: If the recovered value is not below LastValidValue, the suspicious value was a false drop and is discarded.
			// JP: 復帰値が LastValidValue 以上なら、中間の疑わしい値は偽低下として破棄する。
			return t.applyNumericValueLocked(newValue, updateTime, quality)
		}

		// CN: 恢复值低于 LastValidValue，认定为真实复位；先补发基准值，再处理新值。
		// EN: If the recovered value is lower, treat it as a real reset; emit the baseline value first, then the new value.
		// JP: 復帰値が低い場合は実リセットとみなし、基準値を先に出してから新値を処理する。
		changes := t.applyNumericValueLocked(pendingValue, pendingTime, pendingQuality)
		changes = append(changes, t.applyNumericValueLocked(newValue, updateTime, quality)...)
		return changes
	}

	return t.applyNumericValueLocked(newValue, updateTime, quality)
}

func (t *Tag) startupDebounceReferenceLocked() (float64, bool) {
	if t.DebounceLastValidValue == nil {
		return 0, false
	}
	return *t.DebounceLastValidValue, true
}

func (t *Tag) resolveStartupDebounceLocked(newValue float64, updateTime time.Time, quality int, referenceValue float64) []ValueChange {
	pendingValue := t.DebounceValue
	pendingTime := t.DebounceTime
	pendingQuality := t.DebounceQuality
	t.clearDebounceLocked()

	if newValue >= referenceValue {
		t.CurrentValue = newValue
		t.LastValue = newValue
		t.LastUpdateTime = updateTime
		t.Quality = quality
		t.LastQuality = quality
		t.IsFirstUpdate = false
		return nil
	}

	t.CurrentValue = referenceValue
	t.LastValue = referenceValue
	t.Quality = pendingQuality
	t.LastQuality = pendingQuality
	t.IsFirstUpdate = false

	changes := t.applyNumericValueLocked(pendingValue, pendingTime, pendingQuality)
	changes = append(changes, t.applyNumericValueLocked(newValue, updateTime, quality)...)
	return changes
}

func (t *Tag) applyNumericValueLocked(newValue float64, updateTime time.Time, quality int) []ValueChange {
	changed := false
	oldValue := t.CurrentValue
	if t.CurrentValue != newValue {
		t.LastValue = t.CurrentValue
		t.CurrentValue = newValue
		changed = true
	}

	t.LastQuality = t.Quality
	t.Quality = quality
	t.LastUpdateTime = updateTime

	if !changed {
		return nil
	}

	return []ValueChange{{
		OldValue:  oldValue,
		NewValue:  newValue,
		Timestamp: updateTime,
	}}
}

func (t *Tag) clearDebounceLocked() {
	t.DebouncePending = false
	t.DebounceValue = 0
	t.DebounceTime = time.Time{}
	t.DebounceQuality = 0
}

func floatsEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.000000001
}

// HasPendingDebounce 判断是否有采集层可疑值仍在观察窗内。
func (t *Tag) HasPendingDebounce() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.DebouncePending
}

// ShouldStartupSnapshot 判断冷启动首帧是否只入库不触发业务。
// CN: NULL 保持旧表行为；显式 0 可关闭首帧穿透，避免启动时写入不需要的快照。
// EN: NULL preserves legacy-table behavior; explicit 0 disables first-frame storage when a point should not punch through.
// JP: NULL は旧テーブル動作を維持し、明示的な 0 は不要な初回フレーム保存を止める。
func (t *Tag) ShouldStartupSnapshot() bool {
	if t.StoreMode != 1 && t.StoreMode != 3 {
		return false
	}
	if t.StartupSnapshotEnable == nil {
		return true
	}
	return *t.StartupSnapshotEnable == 1
}

// UpdateStringValue 线程安全更新字符串值 - 返回是否变动
// quality: 数据质量/连接状态 (1=在线, 0=离线)
func (t *Tag) UpdateStringValue(newValue string, updateTime time.Time, quality int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 首次更新标记：系统启动后第一次收到数据
	if t.IsFirstUpdate {
		// 首次更新：直接赋值，不算作"变化"
		t.CurrentStrValue = newValue
		t.LastStrValue = newValue
		t.LastUpdateTime = updateTime
		t.Quality = quality     // 更新质量戳/连接状态
		t.LastQuality = quality // 初始化上次质量码
		t.IsFirstUpdate = false
		return false // 不触发任务
	}

	changed := false
	if t.CurrentStrValue != newValue {
		t.LastStrValue = t.CurrentStrValue
		t.CurrentStrValue = newValue
		changed = true
	}

	// 🔥 更新质量码/连接状态
	t.LastQuality = t.Quality
	t.Quality = quality

	t.LastUpdateTime = updateTime
	return changed
}

// GetStringValue 线程安全获取字符串值
func (t *Tag) GetStringValue() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.CurrentStrValue
}

// GetQuality 线程安全获取数据质量/连接状态
func (t *Tag) GetQuality() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Quality
}

// GetLastQuality 线程安全获取上次质量码/连接状态
func (t *Tag) GetLastQuality() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.LastQuality
}

// ShouldStore 判断是否需要存储 (变动存储模式)
func (t *Tag) ShouldStore() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.StoreMode == 0 { // 不存储
		return false
	}

	// 变动存储模式 (1 或 3)
	if t.StoreMode == 1 || t.StoreMode == 3 {
		// 🔧 字符串类型: 检查字符串是否变化
		if t.DataType == "STRING" || t.DataType == "TEXT" {
			return t.CurrentStrValue != t.LastStrValue
		}

		// 数值类型: 检查变化量是否超过死区
		diff := t.CurrentValue - t.LastValue
		if diff < 0 {
			diff = -diff
		}
		if diff > t.StoreDeadband {
			return true
		}
	}

	return false
}

// ShouldCycleStore 判断是否到达定时存储周期
func (t *Tag) ShouldCycleStore(now time.Time) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.StoreMode != 2 && t.StoreMode != 3 { // 只有定时或混合模式
		return false
	}

	// 🔧 修复: store_cycle=0 表示不定时存储
	if t.StoreCycle <= 0 {
		return false
	}

	// ✅ 修复: 首次存储时，允许存储0值
	// 原因: 测试环境下很多变量初始值就是0，也需要存储
	// 如果需要区分"未初始化"和"值为0"，可以检查 LastUpdateTime
	if t.LastStoreTime.IsZero() {
		// 首次存储: 检查是否更新过（收到过数据）
		if t.LastUpdateTime.IsZero() {
			return false // 从未收到数据，不存储
		}
		return true // 收到过数据（即使值为0），也存储
	}

	elapsed := now.Sub(t.LastStoreTime).Seconds()
	return elapsed >= float64(t.StoreCycle)
}

// UpdateStoreTime 更新最后存储时间
func (t *Tag) UpdateStoreTime(storeTime time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.LastStoreTime = storeTime
}

// CheckAlarm 检查报警状态 - 返回报警类型和对应的阈值
// 返回: (报警类型, 阈值), 例如 ("HH", 85.0) 或 ("", 0)
func (t *Tag) CheckAlarm() (string, float64) {
	if !t.AlarmEnable {
		return "", 0
	}

	t.mu.RLock()
	val := t.CurrentValue
	t.mu.RUnlock()

	// 按优先级判断: HH > H > L > LL
	if t.LimitHH != 0 && val >= t.LimitHH+t.Deadband {
		return "HH", t.LimitHH
	}
	if t.LimitH != 0 && val >= t.LimitH+t.Deadband {
		return "H", t.LimitH
	}
	if t.LimitL != 0 && val <= t.LimitL-t.Deadband {
		return "L", t.LimitL
	}
	if t.LimitLL != 0 && val <= t.LimitLL-t.Deadband {
		return "LL", t.LimitLL
	}

	return "", 0 // 正常
}

// SetAlarmState 设置报警状态
func (t *Tag) SetAlarmState(state string, recordID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.AlarmState = state
	if state != "" {
		t.AlarmStartTime = time.Now()
		t.AlarmRecordID = recordID
	} else {
		t.AlarmRecordID = 0
	}
}

// GetAlarmState 获取当前报警状态
func (t *Tag) GetAlarmState() (string, int64) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.AlarmState, t.AlarmRecordID
}
