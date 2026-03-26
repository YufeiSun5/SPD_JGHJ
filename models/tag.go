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

	// 运行时状态 (需加锁修改)
	mu              sync.RWMutex
	CurrentValue    float64   `json:"current_value"`
	LastValue       float64   `json:"last_value"`
	CurrentStrValue string    `json:"current_str_value"` // 字符串类型的值
	LastStrValue    string    `json:"last_str_value"`    // 上次字符串值
	LastUpdateTime  time.Time `json:"last_update_time"`
	LastStoreTime   time.Time `json:"last_store_time"`
	IsFirstUpdate   bool      `json:"is_first_update"` // 是否首次更新 (用于过滤冷启动)
	Quality         int       `json:"quality"`         // 数据质量/连接状态: 1=在线(192), 0=离线(非192)
	LastQuality     int       `json:"last_quality"`    // 上次质量码/连接状态 (用于检测在线/离线变化)

	// 报警状态
	AlarmState     string    `json:"alarm_state"`      // "", "HH", "H", "L", "LL"
	AlarmStartTime time.Time `json:"alarm_start_time"` // 报警开始时间
	AlarmRecordID  int64     `json:"alarm_record_id"`  // 当前报警记录ID
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
