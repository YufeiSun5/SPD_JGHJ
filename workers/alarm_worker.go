// ============================================================================
// 报警处理器 (Alarm Worker) - 报警记录写入
// ============================================================================
// 职责: 写入报警记录 + 更新报警状态
// 协程数: 3个
// 输入: AlarmChan (报警任务, 缓冲200)
//
// 处理类型:
//
//	数值报警: HH/H/L/LL (超限报警)
//	系统报警: SYS (设备故障、错误码)
//
// 报警流程:
//  1. 写入报警记录到 sys_alarm_records
//  2. 更新 Tag 的报警状态 (记录报警ID)
//  3. 可扩展: 发邮件/短信通知 (TODO)
//
// 为什么单独一个Worker?
//
//	报警数量少, 但IO慢 (数据库写入 + 可能的邮件发送)
//	独立协程池避免阻塞主流程
//
// 何时修改此文件:
//   - 需要添加邮件/短信通知
//   - 需要实现报警升级逻辑
//   - 需要修改报警记录格式
//
// ============================================================================
package workers

import (
	"log"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
)

// StartAlarmWorkers 启动报警处理工作池 (x3) - 极慢IO
func StartAlarmWorkers(count int) {
	log.Printf("[AlarmWorker] 启动 %d 个报警处理协程...", count)

	for i := 0; i < count; i++ {
		go alarmWorker(i)
	}

	log.Printf("[AlarmWorker] ✅ 所有报警处理器已启动")
}

// alarmWorker 单个报警处理协程
func alarmWorker(id int) {
	log.Printf("[AlarmWorker-%d] 启动成功，等待报警任务...", id)

	for task := range core.AlarmChan {
		if task.IsRecover {
			// 报警恢复
			handleAlarmRecover(id, task)
		} else {
			// 新报警触发
			handleAlarmTrigger(id, task)
		}
	}

	log.Printf("[AlarmWorker-%d] 通道关闭，协程退出", id)
}

// handleAlarmTrigger 处理报警触发
func handleAlarmTrigger(workerID int, task *models.AlarmTask) {
	// log.Printf("[AlarmWorker-%d] 🚨 报警触发: %s=%s, 值=%.2f",
	// 	workerID, task.VarName, task.AlarmType, task.Value)

	// 1. 写入报警记录
	recordID, err := database.InsertAlarmRecord(task)
	if err != nil {
		log.Printf("[AlarmWorker-%d] 写入报警记录失败: %v", workerID, err)
		return
	}

	// 2. 更新Tag的报警状态 (记录ID) - 仅数值报警(HH/H/L/LL)需要更新Tag状态
	// 系统报警(SYS)不更新Tag状态，因为它们是系统级别的
	if task.AlarmType != "SYS" {
		tagManager := core.GetTagManager()
		if tag, exists := tagManager.GetTag(task.VarID); exists {
			tag.SetAlarmState(task.AlarmType, recordID)
		}
	}

	// log.Printf("[AlarmWorker-%d] ✅ 报警记录已创建: ID=%d", workerID, recordID)

	// 3. 发送邮件/短信 (可选，后续实现)
	// sendAlarmNotification(task)
}

// handleAlarmRecover 处理报警恢复
func handleAlarmRecover(workerID int, task *models.AlarmTask) {
	// log.Printf("[AlarmWorker-%d] ✅ 报警恢复: %s, 值=%.2f",
	// 	workerID, task.VarName, task.Value)

	// 更新报警记录的end_time
	if err := database.UpdateAlarmRecordEndTime(task.AlarmRecordID, time.Now()); err != nil {
		log.Printf("[AlarmWorker-%d] 更新报警恢复时间失败: %v", workerID, err)
		return
	}

	// log.Printf("[AlarmWorker-%d] ✅ 报警恢复已记录: RecordID=%d", workerID, task.AlarmRecordID)
}
