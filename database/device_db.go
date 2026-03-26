package database

import (
	"gin-mqtt-pgsql/models"
)

// ========================================================
// 设备管理
// ========================================================

// GetAllDevices 获取所有设备
func GetAllDevices() ([]*models.SysDevice, error) {
	var devices []*models.SysDevice
	err := DB.Order("id").Find(&devices).Error
	return devices, err
}

// GetDeviceByID 根据ID获取设备
func GetDeviceByID(id int) (*models.SysDevice, error) {
	var device models.SysDevice
	err := DB.First(&device, id).Error
	return &device, err
}

// CreateDevice 创建设备
func CreateDevice(device *models.SysDevice) error {
	return DB.Create(device).Error
}

// UpdateDevice 更新设备
func UpdateDevice(id int, updates map[string]interface{}) error {
	return DB.Model(&models.SysDevice{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteDevice 删除设备
func DeleteDevice(id int) error {
	return DB.Delete(&models.SysDevice{}, id).Error
}





