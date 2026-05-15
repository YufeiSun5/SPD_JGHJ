package service

import (
	"time"

	"gin-mqtt-pgsql/database"
)

// GetHourlyProduction 获取今日按小时统计的产量。
// CN: 这些函数目前仍是 database 层聚合透传，放进 service 后 Wails 与未来 Screen API 可共用入口。
// EN: These functions currently pass through database aggregations; service centralizes them for Wails and future Screen APIs.
// JP: これらは現状 database 集計の透過呼び出しだが、service に集約して Wails と将来の Screen API で共用する。
func GetHourlyProduction(deviceID *int) ([]database.HourlyProduction, error) {
	return database.GetHourlyProduction(deviceID)
}

func GetStaffEfficiency(startTime, endTime *time.Time) ([]database.StaffEfficiency, error) {
	return database.GetStaffEfficiency(startTime, endTime)
}

func GetDeviceUtilizationTrend(deviceID *int) ([]database.DeviceUtilizationTrend, error) {
	return database.GetDeviceUtilizationTrend(deviceID)
}

func GetRealHourlyProduction() ([]database.HourlyProductionPulse, error) {
	return database.GetHourlyProductionPulse(nil)
}

func GetHourlyProductionAccurate(configs []database.DeviceVarConfig) ([]database.HourlyProductionAccurate, error) {
	ldShifts, err := GetShiftsForLogicalDay()
	logicalDate := ""
	if err == nil && len(ldShifts) > 0 {
		logicalDate = ldShifts[0].LogicalDate
	}
	return database.GetHourlyProductionAccurate(configs, logicalDate)
}

func GetMonthlyProductionAccurate(configs []database.DeviceVarConfig) ([]database.MonthlyProductionAccurate, error) {
	return database.GetMonthlyProductionAccurate(configs)
}

func GetMonthlyQualityByOrder() ([]database.DeviceQualityStat, error) {
	return database.GetMonthlyQualityByOrder()
}

func GetMonthlyDailyQualityTrend() ([]database.DailyQualityTrend, error) {
	return database.GetMonthlyDailyQualityTrend()
}

func GetDailyQualityByRun() ([]database.DeviceQualityStat, error) {
	return database.GetDailyQualityByRun()
}

func GetActiveOrderQuality() ([]database.DeviceQualityStat, error) {
	return database.GetActiveOrderQuality()
}
