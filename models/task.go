// ============================================================================
// 任务配置模型 (Task) - 任务系统数据结构
// ============================================================================
// 职责: 定义任务的类型、触发条件、动作配置
//
// 任务类型 (TaskType):
//
//	1 - 定时任务: 按 interval_sec 或 cron_expr 定时执行
//	2 - 数据改变任务: 变量值变化时触发
//	3 - 条件事件任务: 条件表达式满足时触发
//
// 变化类型 (ChangeType) - 用于数据改变任务:
//
//	ANY: 任意变化
//	INCREASE: 增加
//	DECREASE: 减少
//	THRESHOLD: 变化量超过阈值
//	FALSE_TO_TRUE: 布尔值从0变为1
//	TRUE_TO_FALSE: 布尔值从1变为0
//	QUALITY_GOOD: 质量码从Bad(0)变为Good(1) 🔥 新增
//	QUALITY_BAD: 质量码从Good(1)变为Bad(0) 🔥 新增
//
// 动作类型 (ActionType):
//
//	1 - HTTP请求: GET/POST/PUT
//	2 - MQTT发布: 发布消息到MQTT主题
//	3 - 数据库操作: 执行SQL或预定义操作
//	4 - 脚本执行: bash/python/powershell
//	5 - 日志写入: INFO/WARN/ERROR
//
// 何时修改此文件:
//   - 需要添加新的任务类型
//   - 需要添加新的变化类型
//   - 需要添加新的动作类型
//
// ============================================================================
package models

import (
	"time"
)

// TaskType 任务类型
type TaskType int

const (
	TaskTypeScheduled  TaskType = 1 // 定时任务
	TaskTypeDataChange TaskType = 2 // 数据改变任务
	TaskTypeCondition  TaskType = 3 // 条件事件任务
)

// TaskActionType 任务动作类型
type TaskActionType int

const (
	ActionHTTPRequest TaskActionType = 1 // HTTP 请求
	ActionMQTTPublish TaskActionType = 2 // MQTT 发布
	ActionDatabase    TaskActionType = 3 // 数据库操作
	ActionScript      TaskActionType = 4 // 执行脚本
	ActionLog         TaskActionType = 5 // 写日志
)

// Task 任务配置 (存储在数据库)
type Task struct {
	TaskID      int64     `json:"task_id" gorm:"primaryKey"`
	TaskName    string    `json:"task_name"`
	TaskType    TaskType  `json:"task_type"`   // 1=定时, 2=数据改变, 3=条件
	IsEnabled   bool      `json:"is_enabled"`  // 是否启用
	Description string    `json:"description"` // 任务描述
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 定时任务配置 (TaskType=1)
	CronExpr    string    `json:"cron_expr"`     // Cron 表达式 (如: "*/5 * * * *")
	IntervalSec int       `json:"interval_sec"`  // 简单间隔(秒), 优先使用 cron_expr
	LastRunTime time.Time `json:"last_run_time"` // 上次执行时间 (内存维护)

	// 数据改变任务配置 (TaskType=2)
	TriggerVarID    int64   `json:"trigger_var_id"`   // 触发的测点ID
	TriggerVarName  string  `json:"trigger_var_name"` // 触发的测点名称
	ChangeType      string  `json:"change_type"`      // ANY(任意变化), INCREASE(增加), DECREASE(减少), THRESHOLD(阈值), FALSE_TO_TRUE(false→true), TRUE_TO_FALSE(true→false), QUALITY_GOOD(质量变好), QUALITY_BAD(质量变坏)
	ChangeThreshold float64 `json:"change_threshold"` // 变化阈值（用于THRESHOLD类型）

	// 条件事件任务配置 (TaskType=3)
	ConditionExpr string `json:"condition_expr"` // 条件表达式 (如: "temp>50 AND pressure>100")

	// 任务动作配置
	ActionType   TaskActionType `json:"action_type"`   // 动作类型
	ActionConfig string         `json:"action_config"` // 动作配置 JSON 字符串
}

// TableName 指定表名为 sys_tasks
func (Task) TableName() string {
	return "sys_tasks"
}

// TaskTriggerEvent 任务触发事件 (LogicWorker -> TaskScheduler)
// 用于解耦 LogicWorker 和 TaskScheduler，避免阻塞
type TaskTriggerEvent struct {
	EventType string    `json:"event_type"` // "data_change" / "data_change_string" / "condition"
	Timestamp time.Time `json:"timestamp"`
	GatewayID int       `json:"gateway_id"`

	// 数据改变任务字段
	VarID    int64   `json:"var_id,omitempty"`
	OldValue float64 `json:"old_value,omitempty"`
	NewValue float64 `json:"new_value,omitempty"`
	StrValue string  `json:"str_value,omitempty"` // 字符串类型的新值

	// 🔥 质量码/连接状态字段 (用于质量码监控)
	OldQuality int `json:"old_quality,omitempty"` // 旧质量码: 1=在线, 0=离线
	NewQuality int `json:"new_quality,omitempty"` // 新质量码: 1=在线, 0=离线

	// 条件事件任务字段
	ConditionData map[string]interface{} `json:"condition_data,omitempty"`
}

// TaskEvent 任务执行事件 (通过 Channel 传递)
type TaskEvent struct {
	TaskID       int64                  `json:"task_id"`
	TaskName     string                 `json:"task_name"`
	TaskType     TaskType               `json:"task_type"`
	ActionType   TaskActionType         `json:"action_type"`
	ActionConfig string                 `json:"action_config"`
	TriggerTime  time.Time              `json:"trigger_time"`
	TriggerData  map[string]interface{} `json:"trigger_data"` // 触发时的上下文数据
	GatewayID    int                    `json:"gateway_id"`   // 触发事件的网关ID
}

// TaskExecutionLog 任务执行日志
type TaskExecutionLog struct {
	LogID       int64     `json:"log_id" gorm:"primaryKey"`
	TaskID      int64     `json:"task_id"`
	TaskName    string    `json:"task_name"`
	ExecuteTime time.Time `json:"execute_time"`
	Success     bool      `json:"success"`
	ErrorMsg    string    `json:"error_msg"`
	Duration    int       `json:"duration"` // 执行耗时(毫秒)
	Result      string    `json:"result"`   // 执行结果
}

// TableName 指定表名为 sys_task_execution_logs
func (TaskExecutionLog) TableName() string {
	return "sys_task_execution_logs"
}

// HTTPActionConfig HTTP 请求动作配置
type HTTPActionConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"` // GET, POST, PUT
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`    // 支持模板变量 {{var_name}}
	Timeout int               `json:"timeout"` // 超时时间(秒)
}

// MQTTActionConfig MQTT 发布动作配置
type MQTTActionConfig struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"` // 支持模板变量 {{var_name}}
	QoS     byte   `json:"qos"`
	Retain  bool   `json:"retain"`
}

// DatabaseOperationType 数据库预定义操作类型
type DatabaseOperationType string

const (
	// 设备状态管理
	DBOpUpdateDeviceStatus            DatabaseOperationType = "update_device_status"             // 更新设备状态
	DBOpUpdateDeviceStatusConditional DatabaseOperationType = "update_device_status_conditional" // 🔥 条件更新设备状态（检查多个变量条件）
	DBOpEndDeviceStatus               DatabaseOperationType = "end_device_status"                // 结束设备状态

	// 工单管理
	DBOpStartOrder    DatabaseOperationType = "start_order"    // 开始工单
	DBOpCompleteOrder DatabaseOperationType = "complete_order" // 完成工单

	// 生产记录
	DBOpIncrementProductionQty DatabaseOperationType = "increment_production_qty" // 增加产量（同时更新工单和班次）

	// 系统报警
	DBOpLogSystemAlarm DatabaseOperationType = "log_system_alarm" // 记录系统报警
)

// DatabaseActionConfig 数据库操作动作配置
type DatabaseActionConfig struct {
	// 模式A: 原始SQL（向后兼容，简单场景）
	SQL    string                 `json:"sql,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`

	// 模式B: 预定义操作（新增，复杂业务逻辑）
	Operation DatabaseOperationType  `json:"operation,omitempty"` // 操作类型常量
	OpParams  map[string]interface{} `json:"op_params,omitempty"` // 操作参数
}

// ScriptActionConfig 脚本执行动作配置
type ScriptActionConfig struct {
	ScriptType string   `json:"script_type"` // bash, python, powershell
	ScriptPath string   `json:"script_path"` // 脚本路径
	Args       []string `json:"args"`        // 命令行参数
	Timeout    int      `json:"timeout"`     // 超时时间(秒)
}

// LogActionConfig 日志写入动作配置
type LogActionConfig struct {
	LogLevel string `json:"log_level"` // INFO, WARN, ERROR
	Message  string `json:"message"`   // 日志内容, 支持模板变量 {{var_name}}
	FilePath string `json:"file_path"` // 日志文件路径(可选)
}
