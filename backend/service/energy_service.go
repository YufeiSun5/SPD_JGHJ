package service

import (
	"fmt"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
)

// DeviceEnergyData 设备能耗数据。
// CN: 实时功率来自内存 TagManager，日电能来自 sys_data_history 差值，集中在 service 保证后续小屏只读复用。
// EN: Real-time power comes from TagManager memory and daily energy from sys_data_history deltas; service keeps it reusable read-only data.
// JP: リアルタイム電力は TagManager メモリ、当日電力量は sys_data_history 差分から取得し、小画面向け読み取り再利用に備える。
type DeviceEnergyData struct {
	DeviceID         int     `json:"device_id"`
	DeviceName       string  `json:"device_name"`
	RealTimePower    float64 `json:"real_time_power"`
	TodayConsumption float64 `json:"today_consumption"`
	PowerUnit        string  `json:"power_unit"`
	EnergyUnit       string  `json:"energy_unit"`
}

func GetTodayEnergyConsumption(varID int64) (float64, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var result struct {
		MaxVal *float64 `gorm:"column:max_val"`
		MinVal *float64 `gorm:"column:min_val"`
	}

	err := database.DB.Table("sys_data_history").
		Select("MAX(val) as max_val, MIN(val) as min_val").
		Where("var_id = ? AND created_at >= ? AND val IS NOT NULL", varID, todayStart).
		Scan(&result).Error
	if err != nil {
		return 0, fmt.Errorf("查询今日电能失败: %v", err)
	}
	if result.MaxVal == nil || result.MinVal == nil {
		return 0, nil
	}
	consumption := *result.MaxVal - *result.MinVal
	if consumption < 0 {
		consumption = 0
	}
	return consumption, nil
}

func GetAllDevicesEnergyData() ([]*DeviceEnergyData, error) {
	config := map[int]struct {
		PowerVarID  int64
		EnergyVarID int64
	}{
		1: {PowerVarID: 86, EnergyVarID: 81},
		2: {PowerVarID: 110, EnergyVarID: 107},
	}

	tagManager := core.GetTagManager()
	results := make([]*DeviceEnergyData, 0, len(config))
	for deviceID, cfg := range config {
		data := &DeviceEnergyData{
			DeviceID:   deviceID,
			PowerUnit:  "kW",
			EnergyUnit: "kWh",
		}

		if device, err := database.GetDeviceByID(deviceID); err == nil {
			data.DeviceName = device.DeviceName
		} else {
			data.DeviceName = fmt.Sprintf("设备%d", deviceID)
		}

		if powerTag, ok := tagManager.GetTag(cfg.PowerVarID); ok && powerTag != nil {
			data.RealTimePower = powerTag.GetValue()
			if powerTag.Unit != "" {
				data.PowerUnit = powerTag.Unit
			}
		}

		if consumption, err := GetTodayEnergyConsumption(cfg.EnergyVarID); err == nil {
			data.TodayConsumption = consumption
		}
		results = append(results, data)
	}
	return results, nil
}
