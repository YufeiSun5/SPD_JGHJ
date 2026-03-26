// ============================================================================
// 逻辑处理核心 (Logic Worker) - 系统最核心的组件 ⭐⭐⭐
// ============================================================================
// 职责: MQTT消息解析 → 内存更新 → 业务分发
// 协程数: 20个并发
// 输入: LogicChan (MQTT原始消息, 缓冲2000)
// 输出: ChangeChan, CycleChan, AlarmChan, SSEChan, 触发任务
//
// 核心设计: 双循环一致性保证 (模仿PLC扫描周期)
//
//	Pass 1: 解析JSON → 更新内存Map → 记录变化 (只赋值,不执行业务)
//	Pass 2: 基于快照执行业务逻辑 (所有变量已是最新值)
//
// 为什么需要双循环?
//
//	避免时空错乱: 确保脚本/规则引擎读取的所有变量都是同一时刻的值
//	例如: "电流>=200A 且 电压>=25V" 时判定合格
//	     如果逐变量更新,可能电流更新时电压还是旧值,导致误判
//
// 何时修改此文件:
//   - 需要支持新的JSON格式 (如新的SCADA协议)
//   - 需要支持新的数据类型
//   - 需要添加新的业务逻辑分发
//
// ============================================================================
package workers

import (
	"fmt"
	"log"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/models"

	"github.com/tidwall/gjson"
)

// ChangedTagInfo 记录变化的测点信息（用于双循环一致性）
type ChangedTagInfo struct {
	Tag       *models.Tag
	OldValue  float64
	NewValue  float64
	IsString  bool
	StrValue  string
	Timestamp time.Time
	InitOnly  bool
}

func enqueueChangeStore(workerID int, info *ChangedTagInfo) {
	storeTask := &models.StoreTask{
		VarID:     info.Tag.VarID,
		VarName:   info.Tag.VarName,
		Value:     info.NewValue,
		StrValue:  info.StrValue,
		IsString:  info.IsString,
		Timestamp: info.Timestamp,
	}

	select {
	case core.ChangeChan <- storeTask:
	default:
		log.Printf("[LogicWorker-%d] ChangeChan已满，丢弃任务: %s", workerID, info.Tag.VarName)
		return
	}

	info.Tag.UpdateStoreTime(info.Timestamp)
}

// StartLogicWorkers 启动逻辑处理器池 (x20) - 计算密集
func StartLogicWorkers(count int) {
	log.Printf("[LogicWorker] 启动 %d 个逻辑处理协程...", count)

	for i := 0; i < count; i++ {
		go logicWorker(i)
	}

	log.Printf("[LogicWorker] ✅ 所有逻辑处理器已启动")
}

// logicWorker 单个逻辑处理协程
func logicWorker(id int) {
	log.Printf("[LogicWorker-%d] 启动成功，等待消息...", id)

	for msg := range core.LogicChan {
		// 高速处理: 解析 -> 更新内存 -> 判断 -> 分发
		processMessage(id, msg)
	}

	log.Printf("[LogicWorker-%d] 通道关闭，协程退出", id)
}

// processMessage 处理单条MQTT消息
func processMessage(workerID int, msg *models.MQTTMessage) {
	gatewayID := msg.GatewayID // 记录消息来源的网关ID
	// 1. 转换为 JSON 字符串供 gjson 解析
	jsonStr := string(msg.Payload)

	// 🔧 调试: 打印收到的原始 MQTT 消息
	// log.Printf("[LogicWorker-%d] 📥 收到MQTT消息 - Topic:%s, Payload:%s", workerID, msg.Topic, jsonStr)

	// 验证 JSON 格式
	if !gjson.Valid(jsonStr) {
		log.Printf("[LogicWorker-%d] ❌ 无效的JSON格式: %s", workerID, jsonStr)
		return
	}

	// 🔧 SCADA格式支持: 提取共享的公共值 (PVs)
	publicValues := gjson.Get(jsonStr, "PVs")
	hasPVs := publicValues.Exists()
	// if hasPVs {
	// 	log.Printf("[LogicWorker-%d] 📋 检测到SCADA格式，PVs: %v", workerID, publicValues.Value())
	// }

	// ====================================================================
	// 双循环一致性原理实现 (PLC扫描周期模式)
	// Pass 1: 输入采样 - 建立当前时刻的全量快照
	// Pass 2: 程序执行 - 基于快照执行所有业务逻辑
	// ====================================================================

	tagManager := core.GetTagManager()
	allTags := tagManager.GetAllTags()

	// 记录变化的测点（用于第二轮循环）
	var changedTags []*ChangedTagInfo

	// ====================================================================
	// Pass 1: 仅更新内存快照 (Input Scan)
	// 目标: 以极快的速度建立"当前时刻的全量快照"
	// 原则: 只赋值，不判断，不触发任何业务逻辑
	// ====================================================================
	processedCount := 0
	for _, tag := range allTags {
		// 🔧 默认质量码: 1=Good (如果没有SCADA格式的质量码，默认为Good)
		quality := 1

		// 使用 gjson 提取值 (支持复杂路径如: Objs.#(N=="可写变量1").1)
		result := gjson.Get(jsonStr, tag.JSONPath)

		// 🔧 SCADA格式: 如果对象内没有值，尝试从 PVs 中获取
		if !result.Exists() && hasPVs {
			// 首先检查该变量是否在 Objs 数组中（通过名称匹配）
			// 提取 json_path 中的变量名 (例如: Objs.#(N=="可写变量1").1 -> "可写变量1")
			objExists := false
			if len(tag.JSONPath) > 0 {
				// 检查 Objs 数组中是否存在匹配的对象
				objCheckPath := "Objs.#(N==\"" + tag.VarName + "\")"
				objCheck := gjson.Get(jsonStr, objCheckPath)
				objExists = objCheck.Exists()
			}

			// 只有当变量在 Objs 中存在时，才从 PVs 获取共享值
			if objExists {
				// 提取 json_path 最后的属性名 (例如: "Objs.#(N=="xxx").1" -> "1")
				// 尝试从 PVs.1 中获取共享值
				pathParts := tag.JSONPath
				lastDotIdx := len(pathParts) - 1
				for i := len(pathParts) - 1; i >= 0; i-- {
					if pathParts[i] == '.' {
						lastDotIdx = i
						break
					}
				}
				propertyKey := pathParts[lastDotIdx+1:]

				// 从 PVs 中获取共享值
				result = gjson.Get(jsonStr, "PVs."+propertyKey)
				// if result.Exists() {
				// 	log.Printf("[LogicWorker-%d] 🔄 使用PVs共享值 - %s: %v (键:%s)",
				// 		workerID, tag.VarName, result.Value(), propertyKey)
				// }
			}
		}

		if !result.Exists() {
			// 该测点在本条消息中不存在，跳过
			continue
		}

		// 🔧 检查值是否为 null - 避免将 null 解析为 0
		if result.Type == gjson.Null {
			log.Printf("[LogicWorker-%d] ⚠️ 跳过 null 值 - %s (路径:%s)", workerID, tag.VarName, tag.JSONPath)
			continue
		}

		// 🔧 SCADA格式: 检查质量码 Q (属性3), 192=Good(1), 其他=Bad(0)
		// 注意: Bad数据也会更新，只是标记质量为0
		if hasPVs {
			// 提取当前对象 (例如: Objs.#(N=="可写变量1"))
			objPath := tag.JSONPath
			if lastDot := len(objPath) - 1; lastDot >= 0 {
				for i := len(objPath) - 1; i >= 0; i-- {
					if objPath[i] == '.' {
						objPath = objPath[:i]
						break
					}
				}
			}

			// 检查质量码: 先检查对象内的 "3"，如果没有则检查 PVs.3
			qualityCode := gjson.Get(jsonStr, objPath+".3")
			if !qualityCode.Exists() {
				qualityCode = gjson.Get(jsonStr, "PVs.3")
			}

			if qualityCode.Exists() {
				if qualityCode.Int() == 192 {
					quality = 1 // Good
				} else {
					quality = 0 // Bad
					log.Printf("[LogicWorker-%d] ⚠️ 质量码异常 - %s: Q=%d (Bad), 仍然更新数据",
						workerID, tag.VarName, qualityCode.Int())
				}
			}
		}

		// 🔧 调试: 记录成功提取的值
		// log.Printf("[LogicWorker-%d] ✅ 提取成功 - %s: %v (类型:%s, 路径:%s)",
		// 	workerID, tag.VarName, result.Value(), result.Type, tag.JSONPath)
		processedCount++

		// 根据数据类型处理
		var floatValue float64
		var strValue string
		var isString bool = false

		switch tag.DataType {
		case "INT16", "INT32", "INT64", "INT", "INTEGER", "LONG":
			floatValue = float64(result.Int())
		case "FLOAT", "DOUBLE", "REAL":
			floatValue = result.Float()
		case "STRING", "TEXT":
			// 字符串类型
			isString = true
			strValue = result.String()
		case "BOOL", "BOOLEAN":
			if result.Bool() {
				floatValue = 1.0
			} else {
				floatValue = 0.0
			}
		default:
			// 默认尝试解析为浮点数
			floatValue = result.Float()
		}

		// Pass 1 核心: 只更新内存，记录变化，不执行任何业务逻辑
		var processedValue float64
		var changed bool
		var oldValue float64

		isFirstUpdate := tag.GetLastUpdateTime().IsZero()
		if isString {
			// 字符串类型: 更新字符串值（带质量戳）
			changed = tag.UpdateStringValue(strValue, msg.Timestamp, quality)
		} else {
			// 数值类型: 应用缩放和偏移后更新（带质量戳）
			oldValue = tag.GetValue()
			processedValue = floatValue*tag.ScaleFactor + tag.OffsetVal
			changed = tag.UpdateValue(processedValue, msg.Timestamp, quality)
		}

		// 记录变化的测点，留待 Pass 2 处理
		if changed {
			// 🔍 特别关注可写变量1的变化
			if tag.VarName == "可写变量1" {
				log.Printf("[LogicWorker-%d] 🔥🔥🔥 可写变量1 变化检测: %.2f -> %.2f",
					workerID, oldValue, processedValue)
			}
			changedTags = append(changedTags, &ChangedTagInfo{
				Tag:       tag,
				OldValue:  oldValue,
				NewValue:  processedValue,
				IsString:  isString,
				StrValue:  strValue,
				Timestamp: msg.Timestamp,
			})
		} else if isFirstUpdate && (tag.StoreMode == 1 || tag.StoreMode == 3) {
			// 冷启动首包不触发任务/报警，但变动存储变量需要补存一次当前有效值。
			if !isString || strValue != "" {
				changedTags = append(changedTags, &ChangedTagInfo{
					Tag:       tag,
					OldValue:  oldValue,
					NewValue:  processedValue,
					IsString:  isString,
					StrValue:  strValue,
					Timestamp: msg.Timestamp,
					InitOnly:  true,
				})
			}
		}
	}

	// 🔧 调试: Pass 1 完成汇总
	// if processedCount > 0 {
	// 	log.Printf("[LogicWorker-%d] ✅ Pass 1 完成: 成功提取 %d 个测点，%d 个值变化",
	// 		workerID, processedCount, len(changedTags))
	// } else {
	// 	log.Printf("[LogicWorker-%d] ⚠️ Pass 1 未提取到任何测点数据，可能 json_path 配置不匹配", workerID)
	// }

	// ====================================================================
	// Pass 2: 执行业务逻辑 (Logic Execution)
	// 目标: 基于 Pass 1 建立的快照，执行所有复杂业务
	// 保证: 此时所有相关变量都已更新，脚本/报警读取其他变量时不会出现时空错乱
	// ====================================================================
	for _, info := range changedTags {
		tag := info.Tag

		if info.InitOnly {
			enqueueChangeStore(workerID, info)
			continue
		}

		// 1. 触发数据改变任务（脚本/规则引擎）
		// 关键: 此时其他变量(如 temperature, pressure)都已在 Pass 1 更新完毕
		// 脚本中读取 GlobalVarMap 的任何变量，都是最新值！
		scheduler := GetTaskScheduler()
		if scheduler != nil {
			// 🔥 获取质量码/连接状态 (用于质量码监控)
			oldQuality := tag.GetLastQuality()
			newQuality := tag.GetQuality()

			if info.IsString {
				// 字符串类型：传递字符串值作为触发数据
				if tag.VarName == "可写变量1" || tag.VarName == "可写变量2" {
					log.Printf("[LogicWorker-%d] 🎯 变量变化: %s (字符串) -> %s", workerID, tag.VarName, info.StrValue)
				}
				scheduler.TriggerStringChangeTask(tag.VarID, info.StrValue, info.Timestamp, gatewayID, oldQuality, newQuality)
			} else {
				// 数值类型：传递数值进行变化检测
				if tag.VarName == "可写变量1" || tag.VarName == "可写变量2" {
					log.Printf("[LogicWorker-%d] 🎯 变量变化: %s (数值) %.2f -> %.2f", workerID, tag.VarName, info.OldValue, info.NewValue)
				}
				scheduler.TriggerDataChangeTask(tag.VarID, info.OldValue, info.NewValue, info.Timestamp, gatewayID, oldQuality, newQuality)
			}
		}

		// 2. 判断是否需要存储 (变动存储)
		if tag.ShouldStore() {
			enqueueChangeStore(workerID, info)
		}

		// 3. 检查报警 (仅数值类型)
		if !info.IsString {
			currentAlarmType, limitValue := tag.CheckAlarm()
			oldAlarmState, oldRecordID := tag.GetAlarmState()

			if currentAlarmType != oldAlarmState {
				// 报警状态变化
				if currentAlarmType != "" {
					// 新报警触发
					alarmTask := &models.AlarmTask{
						VarID:      tag.VarID,
						VarName:    tag.VarName,
						Value:      info.NewValue,
						AlarmType:  currentAlarmType,
						LimitValue: limitValue, // 新增: 记录被超过的阈值
						AlarmMsg:   tag.AlarmMsg,
						StartTime:  info.Timestamp,
						IsRecover:  false,
					}

					// 非阻塞发送
					select {
					case core.AlarmChan <- alarmTask:
					default:
						log.Printf("[LogicWorker-%d] AlarmChan已满，丢弃报警: %s", workerID, tag.VarName)
					}
				} else {
					// 报警恢复
					if oldRecordID > 0 {
						recoverTask := &models.AlarmTask{
							VarID:         tag.VarID,
							VarName:       tag.VarName,
							Value:         info.NewValue,
							AlarmType:     oldAlarmState,
							LimitValue:    0, // 恢复时不需要阈值
							StartTime:     info.Timestamp,
							IsRecover:     true,
							AlarmRecordID: oldRecordID,
						}

						select {
						case core.AlarmChan <- recoverTask:
						default:
							log.Printf("[LogicWorker-%d] AlarmChan已满，丢弃恢复: %s", workerID, tag.VarName)
						}
					}

					// 清除报警状态
					tag.SetAlarmState("", 0)
				}
			}
		}

		// 4. 推送到SSE (实时数据展示)
		sseData := map[string]interface{}{
			"var_name": tag.VarName,
			"quality":  tag.GetQuality(), // 添加质量戳: 1=Good, 0=Bad
		}

		if info.IsString {
			sseData["value"] = info.StrValue
			sseData["type"] = "string"
		} else {
			sseData["value"] = info.NewValue
			sseData["type"] = "numeric"
			sseData["alarm"] = tag.AlarmState
		}

		sseMsg := &models.SSEMessage{
			EventType: "data_update",
			Data:      sseData,
			Timestamp: info.Timestamp,
		}

		select {
		case core.SSEChan <- sseMsg:
		default:
			// SSE满了不记录，避免日志过多
		}
	}

	// ====================================================================
	// Pass 2.5: 触发条件事件任务
	// 在所有变量更新完成后，统一检查条件任务
	// ====================================================================
	if len(changedTags) > 0 {
		// 构建当前所有变量的快照（用于条件表达式求值）
		scheduler := GetTaskScheduler()
		if scheduler != nil && scheduler.HasConditionTasks() {
			conditionData := buildConditionData(allTags)
			scheduler.TriggerConditionTask(conditionData, msg.Timestamp)
		}
	}

	// ====================================================================
	// Pass 2 完成
	// ====================================================================
	// log.Printf("[LogicWorker-%d] ✅ Pass 2 完成: 处理了 %d 个变化的测点 (存储/报警/任务/SSE)",
	// 	workerID, len(changedTags))
}

// buildConditionData 构建条件表达式求值所需的数据快照
// 将所有变量的当前值转换为 map[string]interface{} 格式
func buildConditionData(tags []*models.Tag) map[string]interface{} {
	data := make(map[string]interface{})

	for _, tag := range tags {
		// 使用变量名作为键（条件表达式中使用变量名）
		// 同时支持 var_id 作为键（可选）
		if tag.DataType == "STRING" || tag.DataType == "TEXT" {
			data[tag.VarName] = tag.GetStringValue()
		} else {
			data[tag.VarName] = tag.GetValue()
		}

		// 同时添加 var_{id} 格式的键，方便使用 ID 引用
		varKey := fmt.Sprintf("var_%d", tag.VarID)
		if tag.DataType == "STRING" || tag.DataType == "TEXT" {
			data[varKey] = tag.GetStringValue()
		} else {
			data[varKey] = tag.GetValue()
		}
	}

	return data
}
