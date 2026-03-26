// ============================================================================
// 全局通道定义 (Channels) - Worker之间的通信桥梁
// ============================================================================
// 职责: 定义所有Worker之间的通信通道
//
// 通道列表:
//
//	LogicChan   chan *MQTTMessage  // 缓冲2000 - MQTT消息 → LogicWorker
//	ChangeChan  chan *StoreTask    // 缓冲500  - 变动存储 → ChangeWorker
//	CycleChan   chan *StoreTask    // 缓冲500  - 定时存储 → CycleWorker
//	AlarmChan   chan *AlarmTask    // 缓冲200  - 报警任务 → AlarmWorker
//	EventChan   chan *TaskEvent    // 缓冲300  - 任务事件 → EventProcessor
//	SSEChan     chan *SSEMessage   // 缓冲100  - SSE推送 → SSE Broadcaster
//
// 为什么使用Channel?
//
//	解耦: 生产者和消费者独立运行
//	缓冲: 应对突发流量, 避免阻塞
//	并发: 多个Worker并发消费, 提高吞吐量
//
// 缓冲大小设计原则:
//
//	LogicChan: 最大 (2000) - 入口通道, 承受MQTT突发流量
//	ChangeChan/CycleChan: 中等 (500) - 批量写入, 缓冲足够
//	AlarmChan: 较小 (200) - 报警少, 但重要
//	EventChan: 中等 (300) - 任务触发频率中等
//	SSEChan: 较小 (100) - 前端推送, 实时性要求高
//
// 何时修改此文件:
//   - 需要调整缓冲大小 (根据实际压测结果)
//   - 需要添加新的通道 (如 EmailChan)
//
// ============================================================================
package core

import (
	"gin-mqtt-pgsql/models"
	"log"
	"time"
)

// 全局 Channel 定义 - 解耦计算与IO

var (
	// 高速通道 - 计算密集
	LogicChan chan *models.MQTTMessage // MQTT接收 -> LogicWorker, 缓冲2000

	// 中速通道 - IO密集
	ChangeChan chan *models.StoreTask // 变动存储, 缓冲500
	CycleChan  chan *models.StoreTask // 定时存储, 缓冲500

	// 低速通道 - 极慢IO
	AlarmChan chan *models.AlarmTask // 报警处理, 缓冲200

	// 前端推送通道
	SSEChan chan *models.SSEMessage // SSE广播, 缓冲100

	// 任务事件通道 - 中速IO
	EventChan chan *models.TaskEvent // 定时任务/条件事件, 缓冲300

	// 任务触发通道 - 解耦 LogicWorker 和 TaskScheduler
	// 🔥 超大缓冲: 10000 (支持高并发场景，不丢弃任务)
	TriggerChan chan *models.TaskTriggerEvent // 任务触发事件, 缓冲10000
)

// InitChannels 初始化所有通道
func InitChannels() {
	LogicChan = make(chan *models.MQTTMessage, 2000)
	ChangeChan = make(chan *models.StoreTask, 500)
	CycleChan = make(chan *models.StoreTask, 500)
	AlarmChan = make(chan *models.AlarmTask, 200)
	SSEChan = make(chan *models.SSEMessage, 100)
	EventChan = make(chan *models.TaskEvent, 300)
	TriggerChan = make(chan *models.TaskTriggerEvent, 10000) // 超大缓冲，不丢弃任务
}

// GetChannelStats 获取通道状态 - 用于监控
func GetChannelStats() map[string]int {
	return map[string]int{
		"logic_chan":   len(LogicChan),
		"change_chan":  len(ChangeChan),
		"cycle_chan":   len(CycleChan),
		"alarm_chan":   len(AlarmChan),
		"sse_chan":     len(SSEChan),
		"event_chan":   len(EventChan),
		"trigger_chan": len(TriggerChan),
	}
}

// CheckChannelHealth 检查通道健康状态，返回告警信息
func CheckChannelHealth() []string {
	stats := GetChannelStats()
	alerts := []string{}

	// TriggerChan 告警阈值
	if stats["trigger_chan"] > 5000 {
		alerts = append(alerts, "⚠️ TriggerChan 队列堆积严重 (>5000)")
	} else if stats["trigger_chan"] > 3000 {
		alerts = append(alerts, "⚠️ TriggerChan 队列堆积较多 (>3000)")
	}

	// 其他通道告警（超过80%）
	thresholds := map[string]int{
		"logic_chan":  1600, // 80% of 2000
		"change_chan": 400,  // 80% of 500
		"cycle_chan":  400,  // 80% of 500
		"alarm_chan":  160,  // 80% of 200
		"event_chan":  240,  // 80% of 300
	}

	for name, threshold := range thresholds {
		if stats[name] > threshold {
			alerts = append(alerts, "⚠️ "+name+" 队列使用率超过80%")
		}
	}

	return alerts
}

// TriggerSystemAlarm 触发系统报警 (设备故障/系统错误)
// varID: 关联的测点ID (如果有), 0表示系统级报警
// varName: 设备/系统名称
// errorCode: 错误码
// alarmMsg: 报警描述信息
func TriggerSystemAlarm(varID int64, varName string, errorCode float64, alarmMsg string) {
	alarmTask := &models.AlarmTask{
		VarID:      varID,
		VarName:    varName,
		Value:      errorCode, // 系统报警时，Value字段存储错误码
		AlarmType:  "SYS",
		LimitValue: 0, // 系统报警不需要阈值
		AlarmMsg:   alarmMsg,
		StartTime:  time.Now(),
		IsRecover:  false,
	}

	// 非阻塞发送
	select {
	case AlarmChan <- alarmTask:
		log.Printf("[SystemAlarm] 🚨 系统报警已触发: %s (错误码: %.0f)", varName, errorCode)
	default:
		log.Printf("[SystemAlarm] ⚠️ AlarmChan已满，系统报警丢弃: %s", varName)
	}
}

// RecoverSystemAlarm 恢复系统报警
// alarmRecordID: 要恢复的报警记录ID
// varID: 关联的测点ID
// varName: 设备/系统名称
func RecoverSystemAlarm(alarmRecordID int64, varID int64, varName string) {
	recoverTask := &models.AlarmTask{
		VarID:         varID,
		VarName:       varName,
		Value:         0,
		AlarmType:     "SYS",
		LimitValue:    0,
		StartTime:     time.Now(),
		IsRecover:     true,
		AlarmRecordID: alarmRecordID,
	}

	// 非阻塞发送
	select {
	case AlarmChan <- recoverTask:
		log.Printf("[SystemAlarm] ✅ 系统报警已恢复: %s (RecordID: %d)", varName, alarmRecordID)
	default:
		log.Printf("[SystemAlarm] ⚠️ AlarmChan已满，系统报警恢复丢弃: %s", varName)
	}
}
