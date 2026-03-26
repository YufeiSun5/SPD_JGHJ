package models

import "time"

// MQTTMessage MQTT原始消息 - 放入 logicChan
type MQTTMessage struct {
	GatewayID int       `json:"gateway_id"`
	Topic     string    `json:"topic"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

// StoreTask 存储任务 - 放入 changeChan 或 cycleChan
type StoreTask struct {
	VarID     int64     `json:"var_id"`
	VarName   string    `json:"var_name"`
	Value     float64   `json:"value"`
	StrValue  string    `json:"str_value"` // 字符串类型的值
	IsString  bool      `json:"is_string"` // 是否为字符串类型
	Timestamp time.Time `json:"timestamp"`
}

// AlarmTask 报警任务 - 放入 alarmChan
type AlarmTask struct {
	VarID         int64     `json:"var_id"`
	VarName       string    `json:"var_name"`
	Value         float64   `json:"value"`       // 数值报警:触发值; 系统报警:错误码
	AlarmType     string    `json:"alarm_type"`  // "HH", "H", "L", "LL", "SYS"(系统/设备故障)
	LimitValue    float64   `json:"limit_value"` // 被超过的阈值 (新增)
	AlarmMsg      string    `json:"alarm_msg"`
	StartTime     time.Time `json:"start_time"`
	IsRecover     bool      `json:"is_recover"`      // 是否恢复报警
	AlarmRecordID int64     `json:"alarm_record_id"` // 报警记录ID
}

// SSEMessage SSE推送消息 - 放入 sseChan
type SSEMessage struct {
	EventType string                 `json:"event_type"` // "data_update", "alarm"
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}
