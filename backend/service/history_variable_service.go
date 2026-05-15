package service

import (
	"fmt"
	"time"

	"gin-mqtt-pgsql/database"
)

type HistoryRecord struct {
	Timestamp string      `json:"timestamp"`
	Value     interface{} `json:"value"`
}

type HistoryDataResponse struct {
	Records []HistoryRecord `json:"records"`
	Total   int64           `json:"total"`
}

func GetHistoryData(varID int64, startTime, endTime string, page, pageSize int) (HistoryDataResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	baseQuery := database.DB.Table("sys_data_history").Where("var_id = ?", varID)
	if startTime != "" {
		baseQuery = baseQuery.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		baseQuery = baseQuery.Where("created_at <= ?", endTime)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return HistoryDataResponse{}, fmt.Errorf("查询历史数据总数失败: %v", err)
	}

	var results []struct {
		Val       *float64  `gorm:"column:val"`
		StrVal    *string   `gorm:"column:str_val"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	offset := (page - 1) * pageSize
	query := baseQuery.Order("created_at DESC").Limit(pageSize).Offset(offset)
	if err := query.Find(&results).Error; err != nil {
		return HistoryDataResponse{}, fmt.Errorf("查询历史数据失败: %v", err)
	}

	data := make([]HistoryRecord, len(results))
	for i, row := range results {
		record := HistoryRecord{
			Timestamp: row.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if row.StrVal != nil {
			record.Value = *row.StrVal
		} else if row.Val != nil {
			record.Value = *row.Val
		}
		data[i] = record
	}

	return HistoryDataResponse{Records: data, Total: total}, nil
}

func GetAllVariables() ([]database.VariableRow, error) {
	var variables []database.VariableRow
	result := database.DB.Table("sys_variables").Order("id").Find(&variables)
	if result.Error != nil {
		return nil, fmt.Errorf("查询变量配置失败: %v", result.Error)
	}
	return variables, nil
}

// UpdateVariable writes point configuration and bumps the config version for hot reload.
// CN: 点位配置更新后必须更新配置版本，触发采集配置热重载；这里统一放在 service，避免多个入口漏掉。
// EN: Variable updates must bump the config version to trigger hot reload; service centralizes that rule.
// JP: 点位設定更新後は設定バージョンを更新してホットリロードを発火する。この規則を service に集約する。
func UpdateVariable(variable database.VariableRow) error {
	fields := variableFields(variable)
	result := database.DB.Table("sys_variables").Where("id = ?", variable.ID).Updates(fields)
	if result.Error != nil {
		return fmt.Errorf("更新变量配置失败: %v", result.Error)
	}
	return touchConfigVersion()
}

func BatchUpdateVariables(variables []database.VariableRow) error {
	if len(variables) == 0 {
		return fmt.Errorf("没有要更新的变量")
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, variable := range variables {
		result := tx.Table("sys_variables").Where("id = ?", variable.ID).Updates(variableFields(variable))
		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("批量更新失败: %v", result.Error)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	return touchConfigVersion()
}

func CreateVariable(variable database.VariableRow) error {
	if variable.VarName == "" {
		return fmt.Errorf("变量名不能为空")
	}
	if variable.JSONPath == "" {
		return fmt.Errorf("JSON路径不能为空")
	}
	if variable.DataType == nil || *variable.DataType == "" {
		defaultDataType := "FLOAT"
		variable.DataType = &defaultDataType
	}
	if variable.RWMode == nil || *variable.RWMode == "" {
		defaultRWMode := "R"
		variable.RWMode = &defaultRWMode
	}
	if variable.ScaleFactor == 0 {
		variable.ScaleFactor = 1.0
	}

	result := database.DB.Table("sys_variables").Create(variableFields(variable))
	if result.Error != nil {
		return fmt.Errorf("创建变量配置失败: %v", result.Error)
	}
	return touchConfigVersion()
}

func DeleteVariable(id int64) error {
	if id <= 0 {
		return fmt.Errorf("无效的变量ID")
	}

	var count int64
	database.DB.Table("sys_variables").Where("id = ?", id).Count(&count)
	if count == 0 {
		return fmt.Errorf("变量不存在")
	}

	result := database.DB.Table("sys_variables").Where("id = ?", id).Delete(nil)
	if result.Error != nil {
		return fmt.Errorf("删除变量配置失败: %v", result.Error)
	}
	return touchConfigVersion()
}

func BatchDeleteVariables(ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("没有要删除的变量")
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Table("sys_variables").Where("id IN ?", ids).Delete(nil)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("批量删除失败: %v", result.Error)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}
	return touchConfigVersion()
}

func variableFields(variable database.VariableRow) map[string]interface{} {
	return database.AppendVariableAcquisitionConfig(map[string]interface{}{
		"device_id":      variable.DeviceID,
		"var_name":       variable.VarName,
		"display_name":   variable.DisplayName,
		"json_path":      variable.JSONPath,
		"data_type":      variable.DataType,
		"rw_mode":        variable.RWMode,
		"unit":           variable.Unit,
		"scale_factor":   variable.ScaleFactor,
		"offset_val":     variable.OffsetVal,
		"store_mode":     variable.StoreMode,
		"store_cycle":    variable.StoreCycle,
		"store_deadband": variable.StoreDeadband,
		"alarm_enable":   variable.AlarmEnable,
		"limit_hh":       variable.LimitHH,
		"limit_h":        variable.LimitH,
		"limit_l":        variable.LimitL,
		"limit_ll":       variable.LimitLL,
		"deadband":       variable.Deadband,
		"alarm_msg":      variable.AlarmMsg,
	}, variable)
}

func touchConfigVersion() error {
	newVersion := fmt.Sprintf("v%d", time.Now().Unix())
	if err := database.UpdateConfigVersion(newVersion); err != nil {
		return fmt.Errorf("触发热重载失败: %v", err)
	}
	return nil
}
