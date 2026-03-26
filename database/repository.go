// ============================================================================
// 数据库访问层 (Database Repository) - 核心数据访问 ⭐⭐⭐
// ============================================================================
// 职责: 封装所有数据库操作，提供统一的数据访问接口
//
// 核心功能分类:
//
// 【1. 测点配置管理】
//
//	LoadVariables: 加载所有测点配置到内存
//
// 【2. 历史数据存储】
//
//	InsertHistoryDataBatch: 批量写入历史数据（性能优化）
//
// 【3. 报警管理】
//
//	InsertAlarmRecord: 写入报警记录
//	UpdateAlarmRecordEndTime: 更新报警恢复时间
//	LogSystemAlarmWithErrorCode: 记录系统报警（查询错误码表）
//
// 【4. 设备管理】
//
//	GetDeviceByID: 根据ID获取设备信息
//	GetAllDevices: 获取所有设备列表
//
// 【5. 生产管理】
//
//	IncrementProductionQty: 增加产量（工单+班次）⭐ 复杂业务逻辑
//	  - 自动查找活动工单
//	  - 自动查找活动班次
//	  - 事务保证数据一致性
//
// 【6. 配置管理】
//
//	GetConfigVersion: 获取配置版本号（用于热重载）
//
// 【7. 错误码管理】
//
//	GetErrorCodeInfo: 根据错误码查询错误信息
//
// 【8. 网关管理】
//
//	LoadGatewayConfigs: 加载MQTT网关配置
//
// 设计原则:
//   - 所有数据库操作封装在此文件
//   - 使用事务保证数据一致性
//   - 批量操作优化性能
//   - 错误处理统一规范
//
// 何时修改此文件:
//   - 需要添加新的数据库表操作
//   - 需要优化SQL查询性能
//   - 需要添加新的预定义业务操作
//
// ============================================================================
package database

import (
	"fmt"
	"log"
	"time"

	"gin-mqtt-pgsql/models"
)

// GatewayConfig 网关配置结构
type GatewayConfig struct {
	ID       int    `gorm:"column:id"`
	Name     string `gorm:"column:gw_name"`
	Status   int    `gorm:"column:status"`
	Broker   string `gorm:"column:mqtt_broker"`
	ClientID string `gorm:"column:mqtt_client_id"`
	Username string `gorm:"column:mqtt_user"`
	Password string `gorm:"column:mqtt_pass"`
	Topic    string `gorm:"column:mqtt_topic"`
}

// TableName 指定表名
func (GatewayConfig) TableName() string {
	return "sys_gateways"
}

// LoadGateways 从数据库加载所有启用的网关配置
func LoadGateways() ([]GatewayConfig, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var gateways []GatewayConfig
	result := DB.Where("status = ?", 1).Order("id").Find(&gateways)
	if result.Error != nil {
		return nil, fmt.Errorf("查询网关配置失败: %w", result.Error)
	}

	return gateways, nil
}

// VariableRow 数据库变量记录
type VariableRow struct {
	ID            int64    `gorm:"column:id"`
	DeviceID      int      `gorm:"column:device_id"`
	VarName       string   `gorm:"column:var_name"`
	DisplayName   *string  `gorm:"column:display_name"`
	DataType      *string  `gorm:"column:data_type"`
	RWMode        *string  `gorm:"column:rw_mode"`
	Unit          *string  `gorm:"column:unit"`
	JSONPath      string   `gorm:"column:json_path"`
	ScaleFactor   float64  `gorm:"column:scale_factor"`
	OffsetVal     float64  `gorm:"column:offset_val"`
	AlarmEnable   bool     `gorm:"column:alarm_enable"`
	LimitHH       *float64 `gorm:"column:limit_hh"`
	LimitH        *float64 `gorm:"column:limit_h"`
	LimitL        *float64 `gorm:"column:limit_l"`
	LimitLL       *float64 `gorm:"column:limit_ll"`
	Deadband      *float64 `gorm:"column:deadband"`
	AlarmMsg      *string  `gorm:"column:alarm_msg"`
	StoreMode     int      `gorm:"column:store_mode"`
	StoreCycle    int      `gorm:"column:store_cycle"`
	StoreDeadband float64  `gorm:"column:store_deadband"`
}

// TableName 指定表名
func (VariableRow) TableName() string {
	return "sys_variables"
}

// LoadVariables 从数据库加载所有变量配置
func LoadVariables() ([]*models.Tag, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var rows []VariableRow
	result := DB.Table("sys_variables v").
		Select("v.*").
		Joins("INNER JOIN sys_devices d ON v.device_id = d.id").
		Joins("INNER JOIN sys_gateways g ON d.gateway_id = g.id").
		Where("g.status = ?", 1).
		Order("v.id").
		Find(&rows)

	if result.Error != nil {
		return nil, fmt.Errorf("查询变量配置失败: %w", result.Error)
	}

	var tags []*models.Tag
	for _, row := range rows {
		tag := &models.Tag{
			VarID:         row.ID,
			VarName:       row.VarName,
			JSONPath:      row.JSONPath,
			ScaleFactor:   row.ScaleFactor,
			OffsetVal:     row.OffsetVal,
			AlarmEnable:   row.AlarmEnable,
			StoreMode:     row.StoreMode,
			StoreCycle:    row.StoreCycle,
			StoreDeadband: row.StoreDeadband,
			DataType:      "FLOAT", // 默认值
			RWMode:        "R",     // 默认只读
			IsFirstUpdate: true,    // 🔥 冷启动保护：首次更新不触发任务
		}

		// 处理可空字段
		if row.DisplayName != nil {
			tag.DisplayName = *row.DisplayName
		}
		if row.DataType != nil {
			tag.DataType = *row.DataType
		}
		if row.RWMode != nil {
			tag.RWMode = *row.RWMode
		}
		if row.Unit != nil {
			tag.Unit = *row.Unit
		}
		if row.AlarmMsg != nil {
			tag.AlarmMsg = *row.AlarmMsg
		}
		if row.LimitHH != nil {
			tag.LimitHH = *row.LimitHH
		}
		if row.LimitH != nil {
			tag.LimitH = *row.LimitH
		}
		if row.LimitL != nil {
			tag.LimitL = *row.LimitL
		}
		if row.LimitLL != nil {
			tag.LimitLL = *row.LimitLL
		}
		if row.Deadband != nil {
			tag.Deadband = *row.Deadband
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// HistoryData 历史数据记录
type HistoryData struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	VarID     int64     `gorm:"column:var_id"`
	Val       *float64  `gorm:"column:val"`     // 数值 (可为NULL,字符串类型时为NULL)
	StrVal    *string   `gorm:"column:str_val"` // 字符串值 (可为NULL,数值类型时为NULL)
	CreatedAt time.Time `gorm:"column:created_at"`
}

// TableName 指定表名
func (HistoryData) TableName() string {
	return "sys_data_history"
}

// InsertHistoryData 插入历史数据
func InsertHistoryData(varID int64, value float64, timestamp time.Time) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	data := HistoryData{
		VarID:     varID,
		Val:       &value, // 指针类型
		CreatedAt: timestamp,
	}

	result := DB.Create(&data)
	if result.Error != nil {
		return fmt.Errorf("插入历史数据失败: %w", result.Error)
	}

	return nil
}

// InsertHistoryDataBatch 批量插入历史数据
func InsertHistoryDataBatch(tasks []*models.StoreTask) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	if len(tasks) == 0 {
		return nil
	}

	// 转换为数据库模型
	dataList := make([]HistoryData, len(tasks))
	for i, task := range tasks {
		data := HistoryData{
			VarID:     task.VarID,
			CreatedAt: task.Timestamp,
		}

		// 根据类型选择存储字段
		if task.IsString {
			// 字符串类型: 只存 str_val, val 为 NULL
			data.StrVal = &task.StrValue
			data.Val = nil
		} else {
			// 数值类型: 只存 val, str_val 为 NULL
			data.Val = &task.Value
			data.StrVal = nil
		}

		dataList[i] = data
	}

	// 批量插入 (GORM会自动使用事务)
	result := DB.Create(&dataList)
	if result.Error != nil {
		log.Printf("[数据库] 批量插入失败: %v", result.Error)
		return fmt.Errorf("批量插入失败: %w", result.Error)
	}

	return nil
}

// AlarmRecord 报警记录
type AlarmRecord struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement"`
	VarID      int64      `gorm:"column:var_id"`
	VarName    *string    `gorm:"column:var_name"`
	Val        *float64   `gorm:"column:val"`         // 数值报警:触发值; 系统报警:错误码
	AlarmType  string     `gorm:"column:alarm_type"`  // HH, H, L, LL, SYS(系统/设备故障)
	LimitValue *float64   `gorm:"column:limit_value"` // 被超过的阈值 (新增)
	Msg        *string    `gorm:"column:msg"`
	StartTime  time.Time  `gorm:"column:start_time"`
	EndTime    *time.Time `gorm:"column:end_time"`
	AckStatus  int        `gorm:"column:ack_status;default:0"`
}

// TableName 指定表名
func (AlarmRecord) TableName() string {
	return "sys_alarm_records"
}

// InsertAlarmRecord 插入报警记录
func InsertAlarmRecord(task *models.AlarmTask) (int64, error) {
	if DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	record := AlarmRecord{
		VarID:      task.VarID,
		VarName:    &task.VarName,
		Val:        &task.Value,
		AlarmType:  task.AlarmType,
		LimitValue: &task.LimitValue, // 新增: 记录阈值
		Msg:        &task.AlarmMsg,
		StartTime:  task.StartTime,
		AckStatus:  0,
	}

	result := DB.Create(&record)
	if result.Error != nil {
		return 0, fmt.Errorf("插入报警记录失败: %w", result.Error)
	}

	return record.ID, nil
}

// UpdateAlarmRecordEndTime 更新报警恢复时间
func UpdateAlarmRecordEndTime(recordID int64, endTime time.Time) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	result := DB.Model(&AlarmRecord{}).
		Where("id = ?", recordID).
		Update("end_time", endTime)

	if result.Error != nil {
		return fmt.Errorf("更新报警恢复时间失败: %w", result.Error)
	}

	return nil
}

// ConfigVersion 配置版本表
type ConfigVersion struct {
	ID          int       `gorm:"column:id;primaryKey"`
	VersionCode string    `gorm:"column:version_code"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// TableName 指定表名
func (ConfigVersion) TableName() string {
	return "sys_config_version"
}

// GetConfigVersion 获取当前配置版本号
func GetConfigVersion() (string, error) {
	if DB == nil {
		return "", fmt.Errorf("database not initialized")
	}

	var config ConfigVersion
	result := DB.First(&config, 1) // ID=1的记录
	if result.Error != nil {
		return "", fmt.Errorf("获取配置版本失败: %w", result.Error)
	}

	return config.VersionCode, nil
}

// UpdateConfigVersion 更新配置版本号 (触发热重载)
func UpdateConfigVersion(newVersion string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	result := DB.Model(&ConfigVersion{}).
		Where("id = ?", 1).
		Update("version_code", newVersion)

	if result.Error != nil {
		return fmt.Errorf("更新配置版本失败: %w", result.Error)
	}

	return nil
}

// ErrorCode 错误代码表
type ErrorCode struct {
	ErrorCode int       `gorm:"column:error_code;primaryKey" json:"ErrorCode"`
	ErrorMsg  string    `gorm:"column:error_msg" json:"ErrorMsg"`
	CreatedAt time.Time `gorm:"column:created_at" json:"CreatedAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"UpdatedAt"`
}

func (ErrorCode) TableName() string {
	return "sys_error_codes"
}

// GetErrorCodeInfo 根据错误码获取错误信息
func GetErrorCodeInfo(errorCode int) (*ErrorCode, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var errCode ErrorCode
	result := DB.Where("error_code = ?", errorCode).First(&errCode)
	if result.Error != nil {
		return nil, fmt.Errorf("错误码 %d 不存在", errorCode)
	}

	return &errCode, nil
}

// LogSystemAlarmWithErrorCode 记录系统报警（通过错误码）
// 🔥 改造：只存储报警码，不查询错误信息（前端查询时再关联错误码表）
func LogSystemAlarmWithErrorCode(varID int64, varName string, errorCode int) (int64, error) {
	if DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	// 错误码为0时忽略（表示正常状态）
	if errorCode == 0 {
		return 0, fmt.Errorf("错误码为0，忽略")
	}

	// 🔥 直接创建报警记录，只存储错误码，不存储错误信息
	// 错误信息将在前端查询时通过联表获取
	errorCodeFloat := float64(errorCode)
	record := AlarmRecord{
		VarID:      varID,
		VarName:    &varName,
		Val:        &errorCodeFloat,
		AlarmType:  "SYS",
		LimitValue: nil,
		Msg:        nil, // 🔥 不存储错误信息，前端联表查询
		StartTime:  time.Now(),
		AckStatus:  0,
	}

	result := DB.Create(&record)
	if result.Error != nil {
		return 0, fmt.Errorf("插入系统报警记录失败: %w", result.Error)
	}

	return record.ID, nil
}
