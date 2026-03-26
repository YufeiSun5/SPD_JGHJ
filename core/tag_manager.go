// ============================================================================
// 测点管理器 (TagManager) - 内存Map管理 ⭐⭐⭐
// ============================================================================
// 职责: 维护内存中的所有测点
// 线程安全: sync.RWMutex 保护并发读写
//
// 核心数据: map[int64]*models.Tag
//
//	Key: VarID (测点ID)
//	Value: *Tag (测点指针)
//
// 关键方法:
//
//	LoadTags(): 从数据库加载测点配置到内存
//	GetTag(): 根据VarID获取测点 (读锁)
//	GetAllTags(): 获取所有测点 (用于扫描)
//
// 为什么需要内存Map?
//
//	性能: 内存读取 < 1μs, 数据库查询 > 1ms (快1000倍)
//	实时性: 所有Worker直接读内存, 无需等待数据库
//
// 何时修改此文件:
//   - 需要添加测点索引 (如按设备ID索引)
//   - 需要实现测点热重载
//   - 需要添加测点统计功能
//
// ============================================================================
package core

import (
	"gin-mqtt-pgsql/models"
	"sync"
)

// TagManager 全局测点管理器 - 内存为王
type TagManager struct {
	mu   sync.RWMutex
	tags map[int64]*models.Tag // key: var_id
}

var globalTagManager *TagManager

// InitTagManager 初始化全局测点管理器
func InitTagManager() {
	globalTagManager = &TagManager{
		tags: make(map[int64]*models.Tag),
	}
}

// GetTagManager 获取全局测点管理器实例
func GetTagManager() *TagManager {
	return globalTagManager
}

// AddTag 添加测点
func (tm *TagManager) AddTag(tag *models.Tag) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tags[tag.VarID] = tag
}

// GetTag 获取测点
func (tm *TagManager) GetTag(varID int64) (*models.Tag, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	tag, exists := tm.tags[varID]
	return tag, exists
}

// GetTagByName 按变量名查找（遍历）
func (tm *TagManager) GetTagByName(varName string) (*models.Tag, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	for _, tag := range tm.tags {
		if tag.VarName == varName {
			return tag, true
		}
	}
	return nil, false
}

// GetAllTags 获取所有测点 (用于Scanner)
func (tm *TagManager) GetAllTags() []*models.Tag {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]*models.Tag, 0, len(tm.tags))
	for _, tag := range tm.tags {
		result = append(result, tag)
	}
	return result
}

// RemoveTag 移除测点
func (tm *TagManager) RemoveTag(varID int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.tags, varID)
}

// LoadTags 从数据库加载所有测点配置 (热重载安全)
func (tm *TagManager) LoadTags(tags []*models.Tag) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 构建新配置的映射表
	newTags := make(map[int64]*models.Tag)
	for _, tag := range tags {
		newTags[tag.VarID] = tag
	}

	// 🔧 热重载优化: 保留运行时状态
	// 对于仍然存在的测点，继承其运行时状态
	for varID, newTag := range newTags {
		if oldTag, exists := tm.tags[varID]; exists {
			// 测点仍存在，保留运行时状态
			newTag.CurrentValue = oldTag.CurrentValue
			newTag.LastValue = oldTag.LastValue
			newTag.CurrentStrValue = oldTag.CurrentStrValue
			newTag.LastStrValue = oldTag.LastStrValue
			newTag.LastUpdateTime = oldTag.LastUpdateTime
			newTag.LastStoreTime = oldTag.LastStoreTime

			// 🔧 保留报警状态 (重要: 防止报警丢失)
			newTag.AlarmState = oldTag.AlarmState
			newTag.AlarmStartTime = oldTag.AlarmStartTime
			newTag.AlarmRecordID = oldTag.AlarmRecordID
		}
	}

	// 原子替换: 一次性切换到新配置
	tm.tags = newTags
}

// Count 获取测点数量
func (tm *TagManager) Count() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.tags)
}
