package service

import (
	"fmt"
	"time"

	"gin-mqtt-pgsql/database"
)

// SyncErrorCode 同步错误码到数据库。
// CN: 错误码属于设备/报警配置数据，Wails 只负责触发，持久化规则放在 service 便于后续 Web API 复用。
// EN: Error codes are device/alarm configuration data; persistence belongs in service for future Web API reuse.
// JP: エラーコードは設備/アラーム設定データであり、永続化規則は将来の Web API 再利用のため service に置く。
func SyncErrorCode(errorCode int, errorMsg string) error {
	if database.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	var existingCode database.ErrorCode
	result := database.DB.Where("error_code = ?", errorCode).First(&existingCode)

	now := time.Now()
	if result.Error == nil {
		existingCode.ErrorMsg = errorMsg
		existingCode.UpdatedAt = now
		if err := database.DB.Save(&existingCode).Error; err != nil {
			return fmt.Errorf("更新错误码失败: %v", err)
		}
		fmt.Printf("✅ [错误码同步] 更新错误码 %d\n", errorCode)
		return nil
	}

	newCode := database.ErrorCode{
		ErrorCode: errorCode,
		ErrorMsg:  errorMsg,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := database.DB.Create(&newCode).Error; err != nil {
		return fmt.Errorf("创建错误码失败: %v", err)
	}
	fmt.Printf("✅ [错误码同步] 创建错误码 %d\n", errorCode)
	return nil
}

// GetAllErrorCodes 获取所有错误码。
func GetAllErrorCodes() ([]*database.ErrorCode, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	var errorCodes []*database.ErrorCode
	result := database.DB.Order("error_code ASC").Find(&errorCodes)
	if result.Error != nil {
		return nil, fmt.Errorf("查询错误码失败: %v", result.Error)
	}
	return errorCodes, nil
}
