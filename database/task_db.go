package database

import (
	"fmt"
	"log"

	"gin-mqtt-pgsql/models"
)

// LoadTasks 从数据库加载所有启用的任务
func LoadTasks() ([]*models.Task, error) {
	var tasks []*models.Task

	// 查询所有启用的任务
	result := DB.Where("is_enabled = ?", true).Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("查询任务失败: %w", result.Error)
	}

	log.Printf("[Database] 加载任务配置: %d 条", len(tasks))
	return tasks, nil
}

// SaveTaskExecutionLog 保存任务执行日志
func SaveTaskExecutionLog(logEntry *models.TaskExecutionLog) error {
	result := DB.Create(logEntry)
	if result.Error != nil {
		return fmt.Errorf("保存任务执行日志失败: %w", result.Error)
	}
	return nil
}

// GetTaskByID 根据ID获取任务
func GetTaskByID(taskID int64) (*models.Task, error) {
	var task models.Task
	result := DB.First(&task, taskID)
	if result.Error != nil {
		return nil, fmt.Errorf("查询任务失败: %w", result.Error)
	}
	return &task, nil
}

// UpdateTaskLastRunTime 更新任务最后执行时间
func UpdateTaskLastRunTime(taskID int64, lastRunTime interface{}) error {
	result := DB.Model(&models.Task{}).
		Where("task_id = ?", taskID).
		Update("last_run_time", lastRunTime)

	if result.Error != nil {
		return fmt.Errorf("更新任务执行时间失败: %w", result.Error)
	}
	return nil
}

// CreateTasksTable 创建任务相关表 (迁移脚本)
func CreateTasksTable() error {
	// 创建 tasks 表
	if err := DB.AutoMigrate(&models.Task{}); err != nil {
		return fmt.Errorf("创建 tasks 表失败: %w", err)
	}

	// 创建 task_execution_logs 表
	if err := DB.AutoMigrate(&models.TaskExecutionLog{}); err != nil {
		return fmt.Errorf("创建 task_execution_logs 表失败: %w", err)
	}

	log.Println("[Database] ✅ 任务表结构已创建")
	return nil
}

// InsertSampleTasks 插入示例任务 (仅用于测试)
func InsertSampleTasks() error {
	// 示例1: 定时任务 - 每5分钟执行一次
	task1 := &models.Task{
		TaskName:    "定时统计任务",
		TaskType:    models.TaskTypeScheduled,
		IsEnabled:   true,
		Description: "每5分钟统计一次数据",
		IntervalSec: 300, // 5分钟
		ActionType:  models.ActionLog,
		ActionConfig: `{
			"log_level": "INFO",
			"message": "定时统计任务执行: 时间={{trigger_time}}"
		}`,
	}

	// 示例2: 数据改变任务 - 温度变化超过5度时触发
	task2 := &models.Task{
		TaskName:        "温度变化告警",
		TaskType:        models.TaskTypeDataChange,
		IsEnabled:       true,
		Description:     "温度变化超过5度时发送HTTP告警",
		TriggerVarName:  "温度传感器1",
		ChangeType:      "THRESHOLD",
		ChangeThreshold: 5.0,
		ActionType:      models.ActionHTTPRequest,
		ActionConfig: `{
			"url": "http://localhost:8080/api/alerts",
			"method": "POST",
			"headers": {
				"Content-Type": "application/json"
			},
			"body": "{\"type\":\"temperature_change\",\"old\":{{old_value}},\"new\":{{new_value}}}"
		}`,
	}

	// 示例3: 条件事件任务 - 温度>50且压力>100时触发
	task3 := &models.Task{
		TaskName:      "高温高压告警",
		TaskType:      models.TaskTypeCondition,
		IsEnabled:     false, // 默认禁用，需要手动启用
		Description:   "温度>50且压力>100时发送MQTT消息",
		ConditionExpr: "temp>50 AND pressure>100",
		ActionType:    models.ActionMQTTPublish,
		ActionConfig: `{
			"topic": "alerts/critical",
			"payload": "{\"alert\":\"high_temp_pressure\",\"temp\":{{temp}},\"pressure\":{{pressure}}}",
			"qos": 1,
			"retain": false
		}`,
	}

	// 批量插入
	tasks := []*models.Task{task1, task2, task3}
	result := DB.Create(&tasks)
	if result.Error != nil {
		return fmt.Errorf("插入示例任务失败: %w", result.Error)
	}

	log.Printf("[Database] ✅ 插入 %d 条示例任务", len(tasks))
	return nil
}





