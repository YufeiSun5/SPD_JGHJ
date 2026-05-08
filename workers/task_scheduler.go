// ============================================================================
// 任务调度器 (Task Scheduler) - 任务系统核心 ⭐⭐
// ============================================================================
// 职责: 管理3种任务的触发
// 协程数: 1个单例
// 输出: EventChan (任务事件, 缓冲300)
//
// 任务类型:
//  1. 定时任务 (TaskType=1): 每秒检查 interval_sec 或 cron_expr
//  2. 数据改变任务 (TaskType=2): 由 LogicWorker 主动触发
//  3. 条件事件任务 (TaskType=3): 条件表达式求值 (TODO)
//
// 触发流程:
//
//	LogicWorker Pass 2 → TriggerDataChangeTask(varID, old, new)
//	                   ↓
//	         检查 change_type (ANY/FALSE_TO_TRUE/THRESHOLD等)
//	                   ↓
//	         非阻塞发送到 EventChan
//
// 何时修改此文件:
//   - 需要支持新的变化类型 (如 EQUAL, NOT_EQUAL)
//   - 需要实现 Cron 表达式解析
//   - 需要实现条件表达式引擎
//
// ============================================================================
package workers

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/models"

	"github.com/expr-lang/expr"
)

// TaskScheduler 定时任务调度器 - 单例协程
type TaskScheduler struct {
	mu              sync.RWMutex
	scheduledTasks  map[int64]*models.Task // 定时任务列表
	dataChangeTasks map[int64]*models.Task // 数据改变任务列表
	conditionTasks  map[int64]*models.Task // 条件事件任务列表
	stopChan        chan struct{}

	// 🔥 去重机制: 防止条件任务在短时间内被多个协程重复触发
	// key: "taskID_timestamp_ms", value: bool
	recentTriggers sync.Map

	// 🔥 条件任务边沿检测: 记录上次求值结果，只在 false→true 时触发
	// CN: 防止条件持续为真时每个周期都重复触发（如设备持续离线导致日志暴增）
	// EN: Prevents repeated triggers while condition stays true (e.g. device stays offline).
	// JP: 条件が真のまま維持されても繰り返し発火しないよう、前回結果を記録する。
	conditionLastResult map[int64]bool
}

var globalScheduler *TaskScheduler

// InitTaskScheduler 初始化任务调度器
func InitTaskScheduler() *TaskScheduler {
	globalScheduler = &TaskScheduler{
		scheduledTasks:      make(map[int64]*models.Task),
		dataChangeTasks:     make(map[int64]*models.Task),
		conditionTasks:      make(map[int64]*models.Task),
		stopChan:            make(chan struct{}),
		conditionLastResult: make(map[int64]bool),
	}
	return globalScheduler
}

// GetTaskScheduler 获取全局调度器实例
func GetTaskScheduler() *TaskScheduler {
	return globalScheduler
}

// StartTaskScheduler 启动任务调度器
// 🔥 多协程架构:
//   - 1个协程: 定时任务扫描 (每秒检查)
//   - 5个协程: 触发事件处理池 (并发处理数据改变/条件任务)
//   - 1个协程: 定期清理去重记录 (每10秒清理一次)
func StartTaskScheduler() {
	log.Println("[TaskScheduler] 启动任务调度协程...")

	// 启动定时任务扫描协程
	go globalScheduler.runScheduledTaskScanner()

	// 启动触发事件处理池 (5个协程)
	for i := 0; i < 5; i++ {
		go globalScheduler.runTriggerProcessor(i)
	}

	// 启动去重记录清理协程
	go globalScheduler.runDedupeCleanup()

	log.Println("[TaskScheduler] ✅ 任务调度器已启动 (1个定时扫描 + 5个触发处理 + 1个去重清理)")
}

// LoadTasks 加载任务配置到内存
func (ts *TaskScheduler) LoadTasks(tasks []*models.Task) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// 清空现有任务
	ts.scheduledTasks = make(map[int64]*models.Task)
	ts.dataChangeTasks = make(map[int64]*models.Task)
	ts.conditionTasks = make(map[int64]*models.Task)

	// 按类型分类加载
	for _, task := range tasks {
		if !task.IsEnabled {
			continue // 跳过未启用的任务
		}

		switch task.TaskType {
		case models.TaskTypeScheduled:
			ts.scheduledTasks[task.TaskID] = task
		case models.TaskTypeDataChange:
			ts.dataChangeTasks[task.TaskID] = task
		case models.TaskTypeCondition:
			// 🔥 如果条件任务配置了定时扫描（interval_sec > 0），也加入定时任务列表
			if task.IntervalSec > 0 {
				ts.scheduledTasks[task.TaskID] = task
				log.Printf("[TaskScheduler] 条件任务 '%s' (ID=%d) 配置了定时扫描: 每%d秒", task.TaskName, task.TaskID, task.IntervalSec)
			} else {
				ts.conditionTasks[task.TaskID] = task
			}
		}
	}

	log.Printf("[TaskScheduler] 加载任务: 定时=%d, 数据改变=%d, 条件=%d",
		len(ts.scheduledTasks), len(ts.dataChangeTasks), len(ts.conditionTasks))
}

// runScheduledTaskScanner 定时任务扫描协程 (单例)
func (ts *TaskScheduler) runScheduledTaskScanner() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	log.Println("[TaskScheduler-Scanner] 开始扫描定时任务...")

	for {
		select {
		case now := <-ticker.C:
			// 定时任务检查
			ts.checkScheduledTasks(now)

		case <-ts.stopChan:
			log.Println("[TaskScheduler-Scanner] 定时扫描已停止")
			return
		}
	}
}

// runTriggerProcessor 触发事件处理协程 (多个并发)
func (ts *TaskScheduler) runTriggerProcessor(id int) {
	log.Printf("[TaskScheduler-Processor-%d] 启动触发事件处理协程...", id)

	for {
		select {
		case triggerEvent := <-core.TriggerChan:
			// 处理异步触发事件
			ts.handleTriggerEvent(triggerEvent)

		case <-ts.stopChan:
			log.Printf("[TaskScheduler-Processor-%d] 触发处理已停止", id)
			return
		}
	}
}

// runDedupeCleanup 定期清理过期的去重记录
// 🔥 清理策略: 每10秒清理一次，删除超过5秒的记录
func (ts *TaskScheduler) runDedupeCleanup() {
	log.Println("[TaskScheduler-Cleanup] 启动去重记录清理协程...")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			cutoff := now.Add(-5 * time.Second)  // 5秒前的记录
			cutoffMs := cutoff.UnixMilli() / 100 // 转换为100ms粒度的时间戳

			cleanCount := 0
			ts.recentTriggers.Range(func(key, value interface{}) bool {
				// key 格式: "taskID_timestamp"
				// 提取 timestamp 部分
				keyStr := key.(string)
				var taskID, timestamp int64
				fmt.Sscanf(keyStr, "%d_%d", &taskID, &timestamp)

				// 如果记录超过5秒，删除
				if timestamp < cutoffMs {
					ts.recentTriggers.Delete(key)
					cleanCount++
				}
				return true // 继续遍历
			})

			if cleanCount > 0 {
				log.Printf("[TaskScheduler-Cleanup] 清理了 %d 条过期去重记录", cleanCount)
			}

		case <-ts.stopChan:
			log.Println("[TaskScheduler-Cleanup] 清理协程已停止")
			return
		}
	}
}

// checkScheduledTasks 检查定时任务是否到期
func (ts *TaskScheduler) checkScheduledTasks(now time.Time) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	for _, task := range ts.scheduledTasks {
		if ts.shouldRunScheduledTask(task, now) {
			// 🔥 如果是条件事件任务（task_type=3），需要评估条件表达式
			if task.TaskType == models.TaskTypeCondition && task.ConditionExpr != "" {
				result, err := evaluateCondition(task.ConditionExpr, nil)
				if err != nil {
					log.Printf("[TaskScheduler-Scanner] ⚠️ 条件表达式求值失败 - %s: %v", task.TaskName, err)
				} else {
					// 边沿检测：只在 false→true 跳变时触发
					// CN: 条件持续为真时不重复触发，防止日志爆涨（如设备持续离线）
					// EN: Only fire on false→true rising edge; skip while condition stays true.
					// JP: false→true の立ち上がりエッジのみ発火し、条件が真のまま繰り返さない。
					prevResult := ts.conditionLastResult[task.TaskID]
					if result && !prevResult {
						log.Printf("[TaskScheduler-Scanner] ✅ 条件上升沿，触发任务 - %s", task.TaskName)
						ts.triggerTask(task, now, nil, 0)
					}
					ts.conditionLastResult[task.TaskID] = result
				}
			} else {
				// 普通定时任务，直接触发
				ts.triggerTask(task, now, nil, 0)
			}

			// 更新最后执行时间 (在实际项目中应该写回数据库)
			task.LastRunTime = now
		}
	}
}

// shouldRunScheduledTask 判断定时任务是否应该执行
func (ts *TaskScheduler) shouldRunScheduledTask(task *models.Task, now time.Time) bool {
	// 如果有 IntervalSec 配置，使用简单间隔模式
	if task.IntervalSec > 0 {
		if task.LastRunTime.IsZero() {
			return true // 首次执行
		}
		elapsed := now.Sub(task.LastRunTime).Seconds()
		return elapsed >= float64(task.IntervalSec)
	}

	// TODO: 如果有 CronExpr，使用 Cron 表达式解析 (需要引入 cron 库)
	// 这里简化处理，可以集成 github.com/robfig/cron/v3

	return false
}

// handleTriggerEvent 处理触发事件 (在 TaskScheduler 协程中执行)
func (ts *TaskScheduler) handleTriggerEvent(event *models.TaskTriggerEvent) {
	switch event.EventType {
	case "data_change":
		ts.processDataChangeTask(event.VarID, event.OldValue, event.NewValue, event.Timestamp, event.GatewayID, event.OldQuality, event.NewQuality)
	case "data_change_string":
		ts.processStringChangeTask(event.VarID, event.StrValue, event.Timestamp, event.GatewayID)
	case "condition":
		ts.processConditionTask(event.ConditionData, event.Timestamp)
	default:
		log.Printf("[TaskScheduler] ⚠️ 未知的触发事件类型: %s", event.EventType)
	}
}

// TriggerDataChangeTask 触发数据改变任务 (从 LogicWorker 调用)
// 【已改为异步非阻塞】发送到 TriggerChan，由 TaskScheduler 协程处理
// 🔥 不丢弃任务: 使用超大缓冲队列 (10000)，队列满时会阻塞等待
func (ts *TaskScheduler) TriggerDataChangeTask(varID int64, oldValue, newValue float64, timestamp time.Time, gatewayID int, oldQuality, newQuality int) {
	event := &models.TaskTriggerEvent{
		EventType:  "data_change",
		VarID:      varID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Timestamp:  timestamp,
		GatewayID:  gatewayID,
		OldQuality: oldQuality, // 🔥 传递旧质量码/连接状态
		NewQuality: newQuality, // 🔥 传递新质量码/连接状态
	}

	// 队列长度监控 (每1000个触发一次告警)
	queueLen := len(core.TriggerChan)
	if queueLen > 5000 && queueLen%1000 == 0 {
		log.Printf("[TaskScheduler] ⚠️ TriggerChan 队列堆积: %d/10000", queueLen)
	}

	// 直接发送，不丢弃 (队列满时会阻塞，但缓冲10000足够大)
	core.TriggerChan <- event
}

// processDataChangeTask 处理数据改变任务 (原 TriggerDataChangeTask 的核心逻辑)
// 支持数值和布尔类型的变化检测，以及质量码/连接状态监控
func (ts *TaskScheduler) processDataChangeTask(varID int64, oldValue, newValue float64, timestamp time.Time, gatewayID int, oldQuality, newQuality int) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// 只输出可写变量1和可写变量2的日志
	isTargetVar := (varID == 1 || varID == 2)
	if isTargetVar {
		log.Printf("[TaskScheduler] 🔍 检查数据改变任务: varID=%d, 值: %.2f->%.2f, 任务总数=%d",
			varID, oldValue, newValue, len(ts.dataChangeTasks))
	}

	matchCount := 0
	for _, task := range ts.dataChangeTasks {
		if task.TriggerVarID == varID {
			matchCount++
			if isTargetVar {
				log.Printf("[TaskScheduler] ✓ 找到匹配任务: %s (ID=%d), 变化类型=%s",
					task.TaskName, task.TaskID, task.ChangeType)
			}

			// 检查变化类型
			shouldTrigger := false
			switch task.ChangeType {
			case "ANY":
				shouldTrigger = true
			case "INCREASE":
				shouldTrigger = newValue > oldValue
			case "DECREASE":
				shouldTrigger = newValue < oldValue
			case "THRESHOLD":
				diff := newValue - oldValue
				if diff < 0 {
					diff = -diff
				}
				shouldTrigger = diff >= task.ChangeThreshold
			case "FALSE_TO_TRUE":
				// 布尔值从 false(0) 变为 true(1)
				shouldTrigger = (oldValue == 0 && newValue == 1)
				if isTargetVar {
					log.Printf("[TaskScheduler]   检查 FALSE_TO_TRUE: old=%.2f, new=%.2f, 结果=%v",
						oldValue, newValue, shouldTrigger)
				}
			case "TRUE_TO_FALSE":
				// 布尔值从 true(1) 变为 false(0)
				shouldTrigger = (oldValue == 1 && newValue == 0)
			case "QUALITY_GOOD":
				// 🔥 质量码/连接状态从 Bad(0) → Good(1) = 设备从离线变为在线
				shouldTrigger = (oldQuality == 0 && newQuality == 1)
				if shouldTrigger {
					log.Printf("[TaskScheduler] 🔌 设备上线: varID=%d, 质量码 %d→%d", varID, oldQuality, newQuality)
				}
			case "QUALITY_BAD":
				// 🔥 质量码/连接状态从 Good(1) → Bad(0) = 设备从在线变为离线
				shouldTrigger = (oldQuality == 1 && newQuality == 0)
				if shouldTrigger {
					log.Printf("[TaskScheduler] 🔌 设备离线: varID=%d, 质量码 %d→%d", varID, oldQuality, newQuality)
				}
			}

			if shouldTrigger {
				if isTargetVar {
					log.Printf("[TaskScheduler] ✅ 任务满足条件，准备触发: %s", task.TaskName)
				}
				triggerData := map[string]interface{}{
					"var_id":    varID,
					"old_value": oldValue,
					"new_value": newValue,
					"change":    newValue - oldValue,
				}
				ts.triggerTask(task, timestamp, triggerData, gatewayID)
			} else {
				if isTargetVar {
					log.Printf("[TaskScheduler] ❌ 任务不满足条件: %s (变化类型=%s)",
						task.TaskName, task.ChangeType)
				}
			}
		}
	}

	if matchCount == 0 && isTargetVar {
		log.Printf("[TaskScheduler] ⚠️ 没有找到匹配 varID=%d 的任务", varID)
	}
}

// TriggerStringChangeTask 触发字符串类型的数据改变任务
// 【已改为异步非阻塞】发送到 TriggerChan，由 TaskScheduler 协程处理
// 🔥 不丢弃任务: 使用超大缓冲队列 (10000)
func (ts *TaskScheduler) TriggerStringChangeTask(varID int64, newValue string, timestamp time.Time, gatewayID int, oldQuality, newQuality int) {
	event := &models.TaskTriggerEvent{
		EventType:  "data_change_string",
		VarID:      varID,
		StrValue:   newValue,
		Timestamp:  timestamp,
		GatewayID:  gatewayID,
		OldQuality: oldQuality, // 🔥 传递旧质量码/连接状态
		NewQuality: newQuality, // 🔥 传递新质量码/连接状态
	}

	// 直接发送，不丢弃
	core.TriggerChan <- event
}

// processStringChangeTask 处理字符串类型的数据改变任务 (原 TriggerStringChangeTask 的核心逻辑)
func (ts *TaskScheduler) processStringChangeTask(varID int64, newValue string, timestamp time.Time, gatewayID int) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	for _, task := range ts.dataChangeTasks {
		if task.TriggerVarID == varID && task.ChangeType == "ANY" {
			triggerData := map[string]interface{}{
				"var_id":    varID,
				"new_value": newValue,
				"value":     newValue,
			}
			ts.triggerTask(task, timestamp, triggerData, gatewayID)
		}
	}
}

// TriggerConditionTask 触发条件事件任务 (从外部调用，传入所有相关变量的快照)
// 【已改为异步非阻塞】发送到 TriggerChan，由 TaskScheduler 协程处理
// 🔥 不丢弃任务: 使用超大缓冲队列 (10000)
func (ts *TaskScheduler) TriggerConditionTask(conditionData map[string]interface{}, timestamp time.Time) {
	event := &models.TaskTriggerEvent{
		EventType:     "condition",
		ConditionData: conditionData,
		Timestamp:     timestamp,
		GatewayID:     0, // 条件任务暂时使用网关0
	}

	// 直接发送，不丢弃
	core.TriggerChan <- event
}

// processConditionTask 处理条件事件任务 (原 TriggerConditionTask 的核心逻辑)
func (ts *TaskScheduler) processConditionTask(conditionData map[string]interface{}, timestamp time.Time) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// 遍历所有条件任务，逐个求值
	for _, task := range ts.conditionTasks {
		result, err := evaluateCondition(task.ConditionExpr, conditionData)
		if err != nil {
			log.Printf("[TaskScheduler] ⚠️ 条件表达式求值失败 - %s: %v", task.TaskName, err)
			continue
		}

		if result {
			log.Printf("[TaskScheduler] ✅ 条件满足，触发任务 - %s", task.TaskName)
			ts.triggerTask(task, timestamp, conditionData, 0) // 条件任务暂时使用网关0
		}
	}
}

// evaluateCondition 求值条件表达式
// 支持格式示例:
//   - "var1 == 1"
//   - "var1 == 1 && var2 == 0"
//   - "temp > 50 || pressure > 100"
//   - "var1 == 1 and var2 == 0"  (支持 and/or/not)
//   - "GetVarValue(76) == 1"  (🔥 获取变量值)
//   - "GetVarQuality(76) == 1"  (🔥 获取质量码)
//
// 变量从 conditionData 中读取，也可以使用函数从内存读取
func evaluateCondition(exprStr string, data map[string]interface{}) (bool, error) {
	log.Printf("[TaskScheduler] 🔍 条件表达式求值: %s", exprStr)

	// 🔥 扩展环境：添加自定义函数
	env := map[string]interface{}{
		"GetVarValue":   getVarValueFunc,   // 获取变量值
		"GetVarQuality": getVarQualityFunc, // 获取质量码
	}

	// 合并 conditionData 到环境
	for k, v := range data {
		env[k] = v
	}

	// 编译表达式
	program, err := expr.Compile(exprStr, expr.Env(env), expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("表达式编译失败: %w", err)
	}

	// 执行表达式
	output, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("表达式执行失败: %w", err)
	}

	// 转换为布尔值
	result, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("表达式结果不是布尔值: %v", output)
	}

	log.Printf("[TaskScheduler] ✅ 条件求值结果: %v", result)
	return result, nil
}

// getVarValueFunc 获取变量值的函数（用于条件表达式）
func getVarValueFunc(varID int64) float64 {
	tagManager := core.GetTagManager()
	if tagManager == nil {
		log.Printf("[TaskScheduler] ⚠️ TagManager 未初始化")
		return 0
	}

	tag, exists := tagManager.GetTag(varID)
	if !exists || tag == nil {
		log.Printf("[TaskScheduler] ⚠️ 变量不存在: VarID=%d", varID)
		return 0
	}

	value := tag.GetValue()
	log.Printf("[TaskScheduler] 📊 GetVarValue(%d) = %.2f", varID, value)
	return value
}

// getVarQualityFunc 获取质量码的函数（用于条件表达式）
func getVarQualityFunc(varID int64) int {
	tagManager := core.GetTagManager()
	if tagManager == nil {
		log.Printf("[TaskScheduler] ⚠️ TagManager 未初始化")
		return 0
	}

	tag, exists := tagManager.GetTag(varID)
	if !exists || tag == nil {
		log.Printf("[TaskScheduler] ⚠️ 变量不存在: VarID=%d", varID)
		return 0
	}

	quality := tag.GetQuality()
	// 🔥 增强日志：显示质量码含义
	statusText := "离线"
	if quality == 1 {
		statusText = "在线"
	}
	log.Printf("[TaskScheduler] 🔌 GetVarQuality(%d) = %d (%s)", varID, quality, statusText)
	return quality
}

// triggerTask 触发任务执行 - 发送到 EventChan
// 🔥 去重机制: 条件任务在 100ms 内只触发一次，防止多协程重复触发
func (ts *TaskScheduler) triggerTask(task *models.Task, triggerTime time.Time, triggerData map[string]interface{}, gatewayID int) {
	// 🔥 去重逻辑: 仅对条件任务生效
	if task.TaskType == models.TaskTypeCondition {
		// 去重键: taskID_timestamp(100ms粒度)
		// 例如: "123_16800000000" 表示任务123在某个100ms时间窗口
		dedupeKey := fmt.Sprintf("%d_%d", task.TaskID, triggerTime.UnixMilli()/100)

		// LoadOrStore: 如果 key 不存在则存储并返回 false，如果已存在则返回 true
		if _, exists := ts.recentTriggers.LoadOrStore(dedupeKey, triggerTime); exists {
			log.Printf("[TaskScheduler] ⚠️ 条件任务重复触发已拦截: %s (ID=%d)", task.TaskName, task.TaskID)
			return
		}
		// 注意: 去重记录会由 runDedupeCleanup 协程定期清理
	}

	event := &models.TaskEvent{
		TaskID:       task.TaskID,
		TaskName:     task.TaskName,
		TaskType:     task.TaskType,
		ActionType:   task.ActionType,
		ActionConfig: task.ActionConfig,
		TriggerTime:  triggerTime,
		TriggerData:  triggerData,
		GatewayID:    gatewayID, // 记录触发事件的网关ID
	}

	// 非阻塞发送
	select {
	case core.EventChan <- event:
		log.Printf("[TaskScheduler] ✅ 触发任务: %s (ID=%d, 类型=%d)", task.TaskName, task.TaskID, task.TaskType)
	default:
		log.Printf("[TaskScheduler] ⚠️ EventChan已满，丢弃任务: %s", task.TaskName)
	}
}

// Stop 停止调度器
func (ts *TaskScheduler) Stop() {
	close(ts.stopChan)
}

// GetTaskCount 获取任务统计
func (ts *TaskScheduler) GetTaskCount() map[string]int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return map[string]int{
		"scheduled":   len(ts.scheduledTasks),
		"data_change": len(ts.dataChangeTasks),
		"condition":   len(ts.conditionTasks),
	}
}

// HasConditionTasks 检查是否有条件任务
func (ts *TaskScheduler) HasConditionTasks() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.conditionTasks) > 0
}

// ManualTriggerTask 手动触发任务 (从前端按钮调用)
// 用于替代MQTT数据变动触发，直接执行任务
func (ts *TaskScheduler) ManualTriggerTask(taskID int64, gatewayID int, triggerData map[string]interface{}) error {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// 从所有任务列表中查找
	var task *models.Task
	if t, ok := ts.scheduledTasks[taskID]; ok {
		task = t
	} else if t, ok := ts.dataChangeTasks[taskID]; ok {
		task = t
	} else if t, ok := ts.conditionTasks[taskID]; ok {
		task = t
	}

	if task == nil {
		return fmt.Errorf("任务不存在或未启用: taskID=%d", taskID)
	}

	log.Printf("[TaskScheduler] 🖱️ 手动触发任务: %s (ID=%d), 来源=前端按钮", task.TaskName, taskID)

	// 如果没有提供triggerData，创建一个默认的
	if triggerData == nil {
		triggerData = map[string]interface{}{
			"trigger_source": "manual",
			"trigger_type":   "frontend_button",
		}
	} else {
		// 添加触发来源标记
		triggerData["trigger_source"] = "manual"
		triggerData["trigger_type"] = "frontend_button"
	}
	// CN: 手动触发绕过 MQTT 采集层，这里把任务原始触发点位带给执行器，供需要补写历史脉冲的任务使用。
	// EN: Manual triggers bypass MQTT acquisition, so pass the task's source variable to the executor for optional history pulse backfill.
	// JP: 手動トリガーは MQTT 収集層を経由しないため、履歴パルス補完用にタスク元の変数を実行器へ渡す。
	triggerData["manual_trigger_var_id"] = task.TriggerVarID
	triggerData["manual_trigger_var_name"] = task.TriggerVarName

	// 触发任务
	ts.triggerTask(task, time.Now(), triggerData, gatewayID)

	return nil
}
