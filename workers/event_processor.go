// ============================================================================
// 事件处理执行器 (Event Processor) - 任务动作执行 ⭐⭐
// ============================================================================
// 职责: 执行任务的5种动作
// 协程数: 3个并发
// 输入: EventChan (任务事件, 缓冲300)
//
// 动作类型:
//  1. HTTP请求 (1): GET/POST/PUT, 自定义Headers, 超时控制
//  2. MQTT发布 (2): 网关感知, QoS/Retain, SCADA格式支持
//  3. 数据库操作 (3): 原始SQL 或 预定义操作 (调用数据访问层)
//  4. 脚本执行 (4): bash/python/powershell, 超时保护
//  5. 日志写入 (5): INFO/WARN/ERROR, 控制台/文件
//
// 模板变量系统:
//
//	支持 {{var_id}}, {{new_value}}, {{old_value}}, {{trigger_time}} 等
//	在执行前自动替换为实际值
//
// 预定义数据库操作:
//
//	update_device_status: 更新设备状态
//	end_device_status: 结束设备状态
//	increment_production_qty: 增加产量 (工单+班次)
//	log_system_alarm: 记录系统报警
//
// 何时修改此文件:
//   - 需要支持新的动作类型 (如发邮件、发短信)
//   - 需要添加新的预定义数据库操作
//   - 需要修改模板变量替换逻辑
//
// ============================================================================
package workers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/gateway"
	"gin-mqtt-pgsql/models"
)

// 仅用于 use_change_delta=true 的累计计数器纠偏。
// 当采集瞬时出现 65,0,66 这类抖动时，用上次成功处理的非零值修正增量。
var productionCounterLastValue sync.Map

func getProductionCounterKey(deviceID int, triggerData map[string]interface{}) string {
	if varID, ok := getTriggerIntValue(triggerData, "var_id"); ok {
		return fmt.Sprintf("%d_%d", deviceID, varID)
	}
	return fmt.Sprintf("%d", deviceID)
}

func getTriggerIntValue(triggerData map[string]interface{}, key string) (int, bool) {
	val, exists := triggerData[key]
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case float64:
		return int(v), true
	case float32:
		return int(v), true
	case int:
		return v, true
	case int64:
		return int(v), true
	case int32:
		return int(v), true
	default:
		return 0, false
	}
}

// StartEventProcessors 启动事件处理器池 (x3-5) - IO密集
func StartEventProcessors(count int) {
	log.Printf("[EventProcessor] 启动 %d 个事件处理协程...", count)

	for i := 0; i < count; i++ {
		go eventProcessor(i)
	}

	log.Printf("[EventProcessor] ✅ 所有事件处理器已启动")
}

// eventProcessor 单个事件处理协程
func eventProcessor(id int) {
	log.Printf("[EventProcessor-%d] 启动成功，等待任务...", id)

	for event := range core.EventChan {
		// 执行任务动作
		executeTaskEvent(id, event)
	}

	log.Printf("[EventProcessor-%d] 通道关闭，协程退出", id)
}

// executeTaskEvent 执行任务事件
func executeTaskEvent(workerID int, event *models.TaskEvent) {
	startTime := time.Now()

	log.Printf("[EventProcessor-%d] 📤 执行任务: %s (类型:%d, 动作:%d)",
		workerID, event.TaskName, event.TaskType, event.ActionType)

	var err error
	var result string

	// 根据动作类型执行
	switch event.ActionType {
	case models.ActionHTTPRequest:
		result, err = executeHTTPAction(event)
	case models.ActionMQTTPublish:
		result, err = executeMQTTAction(event)
	case models.ActionDatabase:
		result, err = executeDatabaseAction(event)
	case models.ActionScript:
		result, err = executeScriptAction(event)
	case models.ActionLog:
		result, err = executeLogAction(event)
	default:
		err = fmt.Errorf("未知的动作类型: %d", event.ActionType)
	}

	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		log.Printf("[EventProcessor-%d] ❌ 任务失败: %s, 耗时=%dms, 错误=%v",
			workerID, event.TaskName, duration, err)
	} else {
		log.Printf("[EventProcessor-%d] ✅ 任务成功: %s, 耗时=%dms, 结果=%s",
			workerID, event.TaskName, duration, result)
	}

	// ✅ 写入任务执行日志到数据库
	saveExecutionLog(event, err == nil, err, duration, result)
}

// saveExecutionLog 保存任务执行日志到数据库
func saveExecutionLog(event *models.TaskEvent, success bool, execErr error, duration int64, result string) {
	errorMsg := ""
	if execErr != nil {
		errorMsg = execErr.Error()
	}

	logEntry := &models.TaskExecutionLog{
		TaskID:      event.TaskID,
		TaskName:    event.TaskName,
		ExecuteTime: event.TriggerTime,
		Success:     success,
		ErrorMsg:    errorMsg,
		Duration:    int(duration),
		Result:      result,
	}

	// 异步保存，不阻塞任务执行
	go func() {
		if err := database.SaveTaskExecutionLog(logEntry); err != nil {
			log.Printf("[EventProcessor] ⚠️ 保存任务日志失败: %v", err)
		}
	}()
}

// executeHTTPAction 执行 HTTP 请求动作
func executeHTTPAction(event *models.TaskEvent) (string, error) {
	var config models.HTTPActionConfig
	if err := json.Unmarshal([]byte(event.ActionConfig), &config); err != nil {
		return "", fmt.Errorf("解析HTTP配置失败: %w", err)
	}

	// 替换模板变量
	body := replaceTemplateVars(config.Body, event.TriggerData)
	url := replaceTemplateVars(config.URL, event.TriggerData)

	// 创建请求
	req, err := http.NewRequest(config.Method, url, bytes.NewBufferString(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 设置超时
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	result := fmt.Sprintf("状态码:%d, 响应:%s", resp.StatusCode, string(respBody))

	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("HTTP错误: %d", resp.StatusCode)
	}

	return result, nil
}

// executeMQTTAction 执行 MQTT 发布动作
func executeMQTTAction(event *models.TaskEvent) (string, error) {
	var config models.MQTTActionConfig
	if err := json.Unmarshal([]byte(event.ActionConfig), &config); err != nil {
		return "", fmt.Errorf("解析MQTT配置失败: %w", err)
	}

	// 替换模板变量
	payload := replaceTemplateVars(config.Payload, event.TriggerData)
	topic := replaceTemplateVars(config.Topic, event.TriggerData)

	// ✅ 使用网关管理器发布消息（使用触发该事件的网关）
	gwManager := gateway.GetGatewayManager()
	if gwManager == nil {
		return "", fmt.Errorf("网关管理器未初始化")
	}

	// 优先使用触发事件的网关ID
	var gw *gateway.Gateway
	var err error

	if event.GatewayID > 0 {
		// 尝试获取指定网关
		gw, err = gwManager.GetGateway(event.GatewayID)
		if err != nil {
			log.Printf("[EventProcessor] ⚠️ 网关 %d 不可用，尝试使用其他网关: %v", event.GatewayID, err)
			gw, err = gwManager.GetFirstActiveGateway()
		}
	} else {
		// 没有指定网关，使用第一个活跃的
		gw, err = gwManager.GetFirstActiveGateway()
	}

	if err != nil {
		return "", fmt.Errorf("没有可用的网关: %w", err)
	}

	// 通过网关发布消息
	if err := gw.Publish(topic, payload, config.QoS, config.Retain); err != nil {
		return "", fmt.Errorf("MQTT发布失败: %w", err)
	}

	result := fmt.Sprintf("发布到主题:%s, 负载:%s", topic, payload)
	log.Printf("[EventProcessor] ✅ MQTT发布成功: %s", result)

	return result, nil
}

// executeDatabaseAction 执行数据库操作动作
func executeDatabaseAction(event *models.TaskEvent) (string, error) {
	var config models.DatabaseActionConfig
	if err := json.Unmarshal([]byte(event.ActionConfig), &config); err != nil {
		return "", fmt.Errorf("解析数据库配置失败: %w", err)
	}

	// ✅ 模式B: 优先检查预定义操作
	if config.Operation != "" {
		return executePreDefinedDBOperation(config, event.TriggerData, event.TriggerTime)
	}

	// ⚠️ 模式A: 原始SQL（向后兼容，用于简单场景）
	sql := replaceTemplateVars(config.SQL, event.TriggerData)

	// 执行 SQL
	db := database.DB
	if db == nil {
		return "", fmt.Errorf("数据库未初始化")
	}

	result := db.Exec(sql)
	if result.Error != nil {
		return "", fmt.Errorf("SQL执行失败: %w", result.Error)
	}

	resultMsg := fmt.Sprintf("执行SQL成功, 影响行数:%d, SQL:%s", result.RowsAffected, sql)
	log.Printf("[EventProcessor] ✅ 数据库操作: %s", resultMsg)

	return resultMsg, nil
}

// executePreDefinedDBOperation 执行预定义的数据库操作
// 这个函数是核心：它像一个路由器，根据操作类型调用对应的数据访问层函数
func executePreDefinedDBOperation(config models.DatabaseActionConfig, triggerData map[string]interface{}, triggerTime time.Time) (string, error) {
	log.Printf("[EventProcessor] 📋 执行预定义操作: %s, 参数: %v", config.Operation, config.OpParams)

	// 根据操作类型路由到对应的执行函数
	switch config.Operation {

	// ========================================
	// 设备状态管理操作
	// ========================================
	case models.DBOpUpdateDeviceStatus:
		return execUpdateDeviceStatus(config.OpParams, triggerData)

	case models.DBOpUpdateDeviceStatusConditional:
		return execUpdateDeviceStatusConditional(config.OpParams, triggerData)

	case models.DBOpEndDeviceStatus:
		return execEndDeviceStatus(config.OpParams, triggerData)

	// ========================================
	// 生产记录操作
	// ========================================
	case models.DBOpIncrementProductionQty:
		return execIncrementProductionQty(config.OpParams, triggerData, triggerTime)

	// ========================================
	// 未来扩展：工单管理操作
	// ========================================
	case models.DBOpStartOrder:
		return "", fmt.Errorf("工单操作暂未实现")

	case models.DBOpCompleteOrder:
		return "", fmt.Errorf("工单操作暂未实现")

	// ========================================
	// 系统报警操作
	// ========================================
	case models.DBOpLogSystemAlarm:
		return execLogSystemAlarm(config.OpParams, triggerData)

	// ========================================
	// 未知操作
	// ========================================
	default:
		return "", fmt.Errorf("未知的预定义操作: %s (支持的操作: %s, %s)",
			config.Operation,
			models.DBOpUpdateDeviceStatus,
			models.DBOpEndDeviceStatus)
	}
}

// execUpdateDeviceStatus 执行"更新设备状态"操作
func execUpdateDeviceStatus(params map[string]interface{}, triggerData map[string]interface{}) (string, error) {
	// 1. 提取参数（支持从配置或触发数据中获取）
	deviceID, err := getIntParam(params, triggerData, "device_id")
	if err != nil {
		return "", fmt.Errorf("缺少参数 device_id: %w", err)
	}

	status, err := getInt8Param(params, triggerData, "status")
	if err != nil {
		return "", fmt.Errorf("缺少参数 status: %w", err)
	}

	// remark 是可选参数
	remarkStr := getStringParam(params, triggerData, "remark", "")
	var remark *string
	if remarkStr != "" {
		remark = &remarkStr
	}

	// 2. ⭐ 调用数据访问层函数（这就是关键！）
	// 注意: 如果当前状态与目标状态相同，函数会返回当前状态而不插入新记录
	result, err := database.UpdateDeviceStatus(deviceID, status, remark)
	if err != nil {
		return "", fmt.Errorf("更新设备状态失败: %w", err)
	}

	// 3. 返回执行结果
	// 🔥 检查是否是"状态未变化"的情况（开始时间早于1秒前，说明是旧记录）
	isUnchanged := time.Since(result.StartTime) > 1*time.Second

	var resultMsg string
	if isUnchanged {
		resultMsg = fmt.Sprintf("设备状态未变化: 设备ID=%d, 当前状态=%d(%s) (跳过重复插入)",
			result.DeviceID, result.Status, getStatusName(result.Status))
		log.Printf("[EventProcessor] ℹ️ %s", resultMsg)
	} else {
		resultMsg = fmt.Sprintf("设备状态已更新: 设备ID=%d, 新状态=%d(%s), 开始时间=%s",
			result.DeviceID, result.Status, getStatusName(result.Status), result.StartTime.Format("2006-01-02 15:04:05"))
		log.Printf("[EventProcessor] ✅ %s", resultMsg)
	}

	return resultMsg, nil
}

// execUpdateDeviceStatusConditional 执行"条件更新设备状态"操作
// 🔥 在更新前检查多个变量的值和质量码是否满足条件
func execUpdateDeviceStatusConditional(params map[string]interface{}, triggerData map[string]interface{}) (string, error) {
	// 1. 提取基本参数
	deviceID, err := getIntParam(params, triggerData, "device_id")
	if err != nil {
		return "", fmt.Errorf("缺少参数 device_id: %w", err)
	}

	status, err := getInt8Param(params, triggerData, "status")
	if err != nil {
		return "", fmt.Errorf("缺少参数 status: %w", err)
	}

	remarkStr := getStringParam(params, triggerData, "remark", "")
	var remark *string
	if remarkStr != "" {
		remark = &remarkStr
	}

	// 2. 🔥 提取并检查条件
	conditions, ok := params["conditions"].(map[string]interface{})
	if !ok || conditions == nil {
		return "", fmt.Errorf("缺少参数 conditions")
	}

	// 检查所有条件是否满足
	conditionsMet, conditionDetails := checkVariableConditions(conditions)
	if !conditionsMet {
		resultMsg := fmt.Sprintf("条件不满足，跳过更新: 设备ID=%d, 条件=%s", deviceID, conditionDetails)
		log.Printf("[EventProcessor] ⚠️ %s", resultMsg)
		return resultMsg, nil // 不是错误，只是条件不满足
	}

	// 3. 条件满足，执行更新
	result, err := database.UpdateDeviceStatus(deviceID, status, remark)
	if err != nil {
		return "", fmt.Errorf("更新设备状态失败: %w", err)
	}

	// 4. 返回执行结果
	resultMsg := fmt.Sprintf("设备状态已更新: 设备ID=%d, 新状态=%d(%s), 开始时间=%s, 条件=%s",
		result.DeviceID, result.Status, getStatusName(result.Status), result.StartTime.Format("2006-01-02 15:04:05"), conditionDetails)

	log.Printf("[EventProcessor] ✅ %s", resultMsg)
	return resultMsg, nil
}

// checkVariableConditions 检查变量条件是否满足
// 返回: (是否满足, 条件详情字符串)
func checkVariableConditions(conditions map[string]interface{}) (bool, string) {
	var details []string
	tagManager := core.GetTagManager()
	if tagManager == nil {
		return false, "变量管理器未初始化"
	}

	// 遍历所有条件
	for key, expectedValue := range conditions {
		// 解析条件键: "开机状态_var_id", "开机状态_value", "开机状态_quality"
		if strings.HasSuffix(key, "_var_id") {
			// 这是变量ID，跳过（由对应的value/quality条件处理）
			continue
		}

		// 提取变量名前缀
		var varPrefix string
		var checkType string // "value" 或 "quality"
		if strings.HasSuffix(key, "_value") {
			varPrefix = strings.TrimSuffix(key, "_value")
			checkType = "value"
		} else if strings.HasSuffix(key, "_quality") {
			varPrefix = strings.TrimSuffix(key, "_quality")
			checkType = "quality"
		} else {
			continue // 未知的条件类型
		}

		// 获取对应的 var_id
		varIDKey := varPrefix + "_var_id"
		varIDFloat, ok := conditions[varIDKey].(float64)
		if !ok {
			details = append(details, fmt.Sprintf("%s=缺少var_id", varPrefix))
			return false, strings.Join(details, ", ")
		}
		varID := int64(varIDFloat)

		// 从变量管理器获取变量
		tag, exists := tagManager.GetTag(varID)
		if !exists || tag == nil {
			details = append(details, fmt.Sprintf("%s(ID=%d)=变量不存在", varPrefix, varID))
			return false, strings.Join(details, ", ")
		}

		// 检查条件
		if checkType == "value" {
			expectedFloat, ok := expectedValue.(float64)
			if !ok {
				details = append(details, fmt.Sprintf("%s=期望值格式错误", varPrefix))
				return false, strings.Join(details, ", ")
			}
			actualValue := tag.GetValue()
			if actualValue != expectedFloat {
				details = append(details, fmt.Sprintf("%s值不匹配(期望=%.0f,实际=%.0f)", varPrefix, expectedFloat, actualValue))
				return false, strings.Join(details, ", ")
			}
			details = append(details, fmt.Sprintf("%s值=%.0f✓", varPrefix, actualValue))
		} else if checkType == "quality" {
			expectedInt, ok := expectedValue.(float64)
			if !ok {
				details = append(details, fmt.Sprintf("%s=期望质量码格式错误", varPrefix))
				return false, strings.Join(details, ", ")
			}
			actualQuality := tag.GetQuality()
			if actualQuality != int(expectedInt) {
				details = append(details, fmt.Sprintf("%s质量码不匹配(期望=%d,实际=%d)", varPrefix, int(expectedInt), actualQuality))
				return false, strings.Join(details, ", ")
			}
			details = append(details, fmt.Sprintf("%s质量=%d✓", varPrefix, actualQuality))
		}
	}

	return true, strings.Join(details, ", ")
}

// execEndDeviceStatus 执行"结束设备状态"操作
func execEndDeviceStatus(params map[string]interface{}, triggerData map[string]interface{}) (string, error) {
	// 1. 提取参数
	deviceID, err := getIntParam(params, triggerData, "device_id")
	if err != nil {
		return "", fmt.Errorf("缺少参数 device_id: %w", err)
	}

	remarkStr := getStringParam(params, triggerData, "remark", "")
	var remark *string
	if remarkStr != "" {
		remark = &remarkStr
	}

	// 2. ⭐ 调用数据访问层函数
	// 注意: 如果当前已经是停机状态，函数会直接返回nil不执行操作
	err = database.EndDeviceStatus(deviceID, remark)
	if err != nil {
		return "", fmt.Errorf("结束设备状态失败: %w", err)
	}

	// 3. 返回执行结果
	resultMsg := fmt.Sprintf("设备%d状态已更新为停机", deviceID)
	log.Printf("[EventProcessor] ✅ %s", resultMsg)
	return resultMsg, nil
}

// execIncrementProductionQty 执行"增加产量"操作（同时更新工单和班次）
// 支持两种模式:
//  1. 固定增量模式: 配置 ok_qty_delta/ng_qty_delta (如脉冲计数)
//  2. 自动增量模式: 配置 use_change_delta=true (如累加计数器)
func execIncrementProductionQty(params map[string]interface{}, triggerData map[string]interface{}, triggerTime time.Time) (string, error) {
	// 1. 提取参数
	deviceID, err := getIntParam(params, triggerData, "device_id")
	if err != nil {
		return "", fmt.Errorf("缺少参数 device_id: %w", err)
	}

	// 2. 判断计数模式
	useChangeDelta := false
	if val, ok := params["use_change_delta"]; ok {
		// 支持多种类型: bool, string, float64
		switch v := val.(type) {
		case bool:
			useChangeDelta = v
		case string:
			useChangeDelta = (v == "true" || v == "1")
		case float64:
			useChangeDelta = (v != 0)
		case int:
			useChangeDelta = (v != 0)
		}
		log.Printf("[execIncrementProductionQty] use_change_delta=%v (原始类型: %T)", useChangeDelta, val)
	}

	var okDelta, ngDelta int
	var newCounterValue int
	var counterValueValid bool

	if useChangeDelta {
		// ========================================
		// 模式A: 自动增量模式 (累加计数器)
		// ========================================
		log.Printf("[execIncrementProductionQty] 🔄 使用自动增量模式")

		// 从 triggerData 中获取变化量 (new_value - old_value)
		changeVal, exists := triggerData["change"]
		if !exists {
			return "", fmt.Errorf("use_change_delta=true 但 triggerData 中没有 change 字段")
		}

		// 转换为整数
		switch v := changeVal.(type) {
		case float64:
			okDelta = int(v)
		case int:
			okDelta = v
		case int64:
			okDelta = int(v)
		default:
			return "", fmt.Errorf("change 字段类型错误: %T", changeVal)
		}

		log.Printf("[execIncrementProductionQty] 📊 从 triggerData 获取增量: change=%d (原始值: %v, 类型: %T)", okDelta, changeVal, changeVal)

		oldCounterValue, oldExists := getTriggerIntValue(triggerData, "old_value")
		newCounterValue, newExists := getTriggerIntValue(triggerData, "new_value")
		counterValueValid = newExists

		if oldExists && newExists && oldCounterValue == 0 && newCounterValue > 0 {
			if varID, ok := getTriggerIntValue(triggerData, "var_id"); ok {
				lastNonzero, found, err := database.GetLastNonzeroHistoryValue(int64(varID), triggerTime)
				if err != nil {
					return "", fmt.Errorf("查询产量计数器历史非零值失败: %w", err)
				}
				if found {
					correctedDelta := correctProductionCounterDelta(okDelta, oldCounterValue, newCounterValue, &lastNonzero)
					if correctedDelta != okDelta {
						log.Printf("[execIncrementProductionQty] 🔧 按历史非零值修正计数器恢复: old=%d, new=%d, last_nonzero=%d, change=%d->%d",
							oldCounterValue, newCounterValue, lastNonzero, okDelta, correctedDelta)
						okDelta = correctedDelta
					}
				}
			}

			counterKey := getProductionCounterKey(deviceID, triggerData)
			if lastValRaw, ok := productionCounterLastValue.Load(counterKey); ok {
				lastCounterValue := lastValRaw.(int)
				if lastCounterValue > 0 && newCounterValue >= lastCounterValue {
					correctedDelta := correctProductionCounterDelta(okDelta, oldCounterValue, newCounterValue, &lastCounterValue)
					if correctedDelta != okDelta {
						log.Printf("[execIncrementProductionQty] 🔧 检测到瞬时置零恢复，修正增量: old=%d, new=%d, last=%d, change=%d->%d",
							oldCounterValue, newCounterValue, lastCounterValue, okDelta, correctedDelta)
						okDelta = correctedDelta
					}
				}
			}
		}

		// 确保增量为正数 (防止设备重启导致计数器归零)
		if okDelta < 0 {
			log.Printf("[EventProcessor] ⚠️ 检测到计数器归零 (change=%d)，忽略此次更新", okDelta)
			return fmt.Sprintf("设备%d计数器归零，已忽略 (change=%d)", deviceID, okDelta), nil
		}

		ngDelta = 0 // 默认不良品为0
	} else {
		// ========================================
		// 模式B: 固定增量模式 (脉冲计数)
		// ========================================
		log.Printf("[execIncrementProductionQty] 🔢 使用固定增量模式")

		// 良品增量（默认+1）
		okDelta = 1
		if _, ok := params["ok_qty_delta"]; ok {
			if intVal, err := getIntParam(params, triggerData, "ok_qty_delta"); err == nil {
				okDelta = intVal
			}
		}

		// 不良品增量（默认+0）
		ngDelta = 0
		if _, ok := params["ng_qty_delta"]; ok {
			if intVal, err := getIntParam(params, triggerData, "ng_qty_delta"); err == nil {
				ngDelta = intVal
			}
		}

		log.Printf("[execIncrementProductionQty] 📊 使用固定增量: ok=%d, ng=%d", okDelta, ngDelta)
	}

	// 3. ⭐ 调用数据访问层函数（会同时更新班次和工单）
	err = database.IncrementProductionQtyByDevice(deviceID, okDelta, ngDelta)
	if err != nil {
		return "", fmt.Errorf("增加产量失败: %w", err)
	}

	historyResult := ""
	if shouldWriteManualHistoryPulse(params, triggerData, ngDelta) {
		historyResult = writeManualHistoryPulse(triggerData, triggerTime)
	}

	if useChangeDelta && counterValueValid && newCounterValue > 0 {
		productionCounterLastValue.Store(getProductionCounterKey(deviceID, triggerData), newCounterValue)
	}

	// 4. 返回执行结果
	mode := "固定增量"
	if useChangeDelta {
		mode = "自动增量"
	}
	resultMsg := fmt.Sprintf("设备%d产量已更新 [%s]: +%d良品, +%d不良品（已同步到工单和班次）", deviceID, mode, okDelta, ngDelta)
	if historyResult != "" {
		resultMsg += "；" + historyResult
	}
	log.Printf("[EventProcessor] ✅ %s", resultMsg)
	return resultMsg, nil
}

func shouldWriteManualHistoryPulse(params map[string]interface{}, triggerData map[string]interface{}, ngDelta int) bool {
	if triggerData == nil {
		return false
	}
	isManualButton := triggerData["trigger_source"] == "manual" && triggerData["trigger_type"] == "frontend_button"
	if !isManualButton {
		return false
	}
	if enabled := getBoolParam(params, triggerData, "write_manual_history", false); enabled {
		return true
	}

	// CN: 兼容现有现场库配置，NG 手动按钮即使还没补 write_manual_history，也应按 trigger_var_id 记录历史脉冲。
	// EN: Keep existing site DB configs working: manual NG tasks should backfill a history pulse by trigger_var_id even before write_manual_history is added.
	// JP: 既存現場 DB 設定との互換性として、write_manual_history 未追加でも手動 NG タスクは trigger_var_id で履歴パルスを補完する。
	return ngDelta != 0
}

func writeManualHistoryPulse(triggerData map[string]interface{}, triggerTime time.Time) string {
	// CN: 前端 NG 按钮直接手动触发任务，绕过 MQTT 入库；工单更新成功后补一条 val=1 历史脉冲，保持报表口径一致。
	// EN: Frontend NG buttons manually trigger tasks and bypass MQTT storage; after the order update succeeds, backfill one val=1 history pulse for reporting consistency.
	// JP: フロントの NG ボタンは手動でタスクを起動し MQTT 保存を通らないため、工単更新成功後に val=1 の履歴パルスを補完し帳票口径を一致させる。
	varID, ok := getTriggerIntValue(triggerData, "manual_trigger_var_id")
	if !ok || varID <= 0 {
		resultMsg := "手动历史脉冲未写入: 缺少有效 manual_trigger_var_id"
		log.Printf("[EventProcessor] ⚠️ %s", resultMsg)
		return resultMsg
	}

	if err := database.InsertHistoryData(int64(varID), 1, triggerTime); err != nil {
		resultMsg := fmt.Sprintf("手动历史脉冲写入失败: var_id=%d, err=%v", varID, err)
		log.Printf("[EventProcessor] ⚠️ %s", resultMsg)
		return resultMsg
	}

	resultMsg := fmt.Sprintf("已补写手动历史脉冲: var_id=%d, val=1", varID)
	log.Printf("[EventProcessor] ✅ %s", resultMsg)
	return resultMsg
}

// correctProductionCounterDelta 修正累计计数器从 0 恢复到非零值时的假增量。
// CN: 设备重连/采集恢复常见 old=0,new=历史累计值，不能把恢复值当成本次产量。
// EN: Reconnect recovery often reports old=0,new=historical total; do not count that total as new output.
// JP: 再接続復旧では old=0,new=累積値 になりやすいため、その累積値を新規産量として数えない。
func correctProductionCounterDelta(rawDelta, oldValue, newValue int, lastNonzero *int) int {
	if rawDelta < 0 || oldValue != 0 || newValue <= 0 || lastNonzero == nil || *lastNonzero <= 0 {
		return rawDelta
	}
	if newValue >= *lastNonzero {
		return newValue - *lastNonzero
	}
	return newValue
}

// executeScriptAction 执行脚本动作
func executeScriptAction(event *models.TaskEvent) (string, error) {
	var config models.ScriptActionConfig
	if err := json.Unmarshal([]byte(event.ActionConfig), &config); err != nil {
		return "", fmt.Errorf("解析脚本配置失败: %w", err)
	}

	// 替换模板变量
	args := make([]string, len(config.Args))
	for i, arg := range config.Args {
		args[i] = replaceTemplateVars(arg, event.TriggerData)
	}

	// 创建命令
	var cmd *exec.Cmd
	switch config.ScriptType {
	case "bash", "sh":
		cmd = exec.Command("bash", append([]string{config.ScriptPath}, args...)...)
	case "python":
		cmd = exec.Command("python", append([]string{config.ScriptPath}, args...)...)
	case "powershell":
		cmd = exec.Command("powershell", append([]string{"-File", config.ScriptPath}, args...)...)
	default:
		return "", fmt.Errorf("不支持的脚本类型: %s", config.ScriptType)
	}

	// 设置超时
	timeout := 60 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}

	// 执行命令
	done := make(chan error, 1)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("启动脚本失败: %w", err)
	}

	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		cmd.Process.Kill()
		return "", fmt.Errorf("脚本执行超时: %v", timeout)
	case err := <-done:
		result := output.String()
		if err != nil {
			return result, fmt.Errorf("脚本执行失败: %w", err)
		}
		return result, nil
	}
}

// executeLogAction 执行日志写入动作
func executeLogAction(event *models.TaskEvent) (string, error) {
	var config models.LogActionConfig
	if err := json.Unmarshal([]byte(event.ActionConfig), &config); err != nil {
		return "", fmt.Errorf("解析日志配置失败: %w", err)
	}

	// 替换模板变量
	message := replaceTemplateVars(config.Message, event.TriggerData)

	// 输出日志
	logMsg := fmt.Sprintf("[%s] %s", config.LogLevel, message)

	switch config.LogLevel {
	case "ERROR":
		log.Printf("🔴 %s", logMsg)
	case "WARN":
		log.Printf("🟡 %s", logMsg)
	default:
		log.Printf("🟢 %s", logMsg)
	}

	// TODO: 如果指定了 FilePath，写入文件

	return logMsg, nil
}

// replaceTemplateVars 替换模板变量 {{var_name}}
func replaceTemplateVars(template string, data map[string]interface{}) string {
	result := template
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// execLogSystemAlarm 执行"记录系统报警"操作
func execLogSystemAlarm(params map[string]interface{}, triggerData map[string]interface{}) (string, error) {
	varID, err := getInt64Param(params, triggerData, "var_id")
	if err != nil {
		return "", fmt.Errorf("缺少参数 var_id: %w", err)
	}

	varName := getStringParam(params, triggerData, "var_name", "未知设备")

	errorCode, err := getIntParam(params, triggerData, "error_code")
	if err != nil {
		return "", fmt.Errorf("缺少参数 error_code: %w", err)
	}

	recordID, err := database.LogSystemAlarmWithErrorCode(varID, varName, errorCode)
	if err != nil {
		return "", fmt.Errorf("记录系统报警失败: %w", err)
	}

	resultMsg := fmt.Sprintf("系统报警已记录: RecordID=%d, 设备=%s, 错误码=%d", recordID, varName, errorCode)
	log.Printf("[EventProcessor] 🚨 %s", resultMsg)
	return resultMsg, nil
}

// ========================================
// 参数提取辅助函数
// ========================================

// getIntParam 获取整数参数（优先从params获取，其次从triggerData）
func getIntParam(params map[string]interface{}, triggerData map[string]interface{}, key string) (int, error) {
	log.Printf("[getIntParam] 🔍 查找参数 %s", key)
	log.Printf("[getIntParam] params=%v", params)
	log.Printf("[getIntParam] triggerData=%v", triggerData)

	// 优先从 op_params 获取
	if val, ok := params[key]; ok {
		log.Printf("[getIntParam] 在params中找到 %s = %v (类型: %T)", key, val, val)
		switch v := val.(type) {
		case float64:
			log.Printf("[getIntParam] ✅ 返回 float64: %d", int(v))
			return int(v), nil
		case int:
			log.Printf("[getIntParam] ✅ 返回 int: %d", v)
			return v, nil
		case int64:
			log.Printf("[getIntParam] ✅ 返回 int64: %d", int(v))
			return int(v), nil
		case string:
			log.Printf("[getIntParam] 发现字符串值: %s, 尝试从triggerData解析", v)
			// 尝试从模板变量中获取 - 去掉 {{}} 包裹
			templateKey := v
			if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
				templateKey = strings.TrimSpace(v[2 : len(v)-2])
				log.Printf("[getIntParam] 解析模板变量: %s -> %s", v, templateKey)
			}

			if triggerData != nil {
				if tvVal, ok := triggerData[templateKey]; ok {
					log.Printf("[getIntParam] 在triggerData中找到 %s = %v (类型: %T)", templateKey, tvVal, tvVal)
					if fv, ok := tvVal.(float64); ok {
						log.Printf("[getIntParam] ✅ 从triggerData返回: %d", int(fv))
						return int(fv), nil
					}
					if iv, ok := tvVal.(int); ok {
						log.Printf("[getIntParam] ✅ 从triggerData返回 int: %d", iv)
						return iv, nil
					}
				} else {
					log.Printf("[getIntParam] ⚠️ triggerData中没有找到键: %s", templateKey)
				}
			}
		}
	}

	// 其次从 trigger_data 获取（支持模板变量）
	if triggerData != nil {
		if val, ok := triggerData[key]; ok {
			log.Printf("[getIntParam] 在triggerData中找到 %s = %v (类型: %T)", key, val, val)
			switch v := val.(type) {
			case float64:
				log.Printf("[getIntParam] ✅ 从triggerData返回 float64: %d", int(v))
				return int(v), nil
			case int:
				log.Printf("[getIntParam] ✅ 从triggerData返回 int: %d", v)
				return v, nil
			case int64:
				log.Printf("[getIntParam] ✅ 从triggerData返回 int64: %d", int(v))
				return int(v), nil
			}
		}
	}

	log.Printf("[getIntParam] ❌ 参数 %s 未找到或类型不匹配", key)
	return 0, fmt.Errorf("参数 %s 不存在或类型错误", key)
}

// getInt8Param 获取 int8 参数
func getInt8Param(params map[string]interface{}, triggerData map[string]interface{}, key string) (int8, error) {
	val, err := getIntParam(params, triggerData, key)
	if err != nil {
		return 0, err
	}
	return int8(val), nil
}

// getInt64Param 获取 int64 参数
func getInt64Param(params map[string]interface{}, triggerData map[string]interface{}, key string) (int64, error) {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v), nil
		case int:
			return int64(v), nil
		case int64:
			return v, nil
		case string:
			if triggerData != nil {
				if tvVal, ok := triggerData[v]; ok {
					if fv, ok := tvVal.(float64); ok {
						return int64(fv), nil
					}
				}
			}
		}
	}

	if triggerData != nil {
		if val, ok := triggerData[key]; ok {
			switch v := val.(type) {
			case float64:
				return int64(v), nil
			case int:
				return int64(v), nil
			case int64:
				return v, nil
			}
		}
	}

	return 0, fmt.Errorf("参数 %s 不存在或类型错误", key)
}

// getStringParam 获取字符串参数（带默认值）
func getStringParam(params map[string]interface{}, triggerData map[string]interface{}, key string, defaultValue string) string {
	// 优先从 op_params 获取
	if val, ok := params[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}

	// 其次从 trigger_data 获取
	if triggerData != nil {
		if val, ok := triggerData[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}

	return defaultValue
}

func getBoolParam(params map[string]interface{}, triggerData map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "1"
		case float64:
			return v != 0
		case int:
			return v != 0
		case int64:
			return v != 0
		}
	}

	if triggerData != nil {
		if val, ok := triggerData[key]; ok {
			switch v := val.(type) {
			case bool:
				return v
			case string:
				return v == "true" || v == "1"
			case float64:
				return v != 0
			case int:
				return v != 0
			case int64:
				return v != 0
			}
		}
	}

	return defaultValue
}

// getStatusName 获取状态名称
func getStatusName(status int8) string {
	names := map[int8]string{0: "空闲", 1: "运行", 2: "故障"}
	if name, ok := names[status]; ok {
		return name
	}
	return "未知"
}
