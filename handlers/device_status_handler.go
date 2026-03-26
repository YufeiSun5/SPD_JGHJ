package handlers

import (
	"net/http"
	"strconv"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 设备状态管理 API
// ========================================================

// UpdateDeviceStatus 更新设备状态
// POST /api/device-status
func UpdateDeviceStatus(c *gin.Context) {
	var req models.UpdateDeviceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 验证状态值
	if req.Status < 0 || req.Status > 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "状态值必须是 0(空闲), 1(运行) 或 2(故障)"})
		return
	}

	status, err := database.UpdateDeviceStatus(req.DeviceID, req.Status, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备状态已更新",
		"data":    status,
	})
}

// GetDeviceCurrentStatus 获取设备当前状态
// GET /api/device-status/:device_id/current
func GetDeviceCurrentStatus(c *gin.Context) {
	deviceIDStr := c.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "设备ID格式错误"})
		return
	}

	status, err := database.GetDeviceCurrentStatus(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "设备当前无状态记录"})
		return
	}

	// 计算当前状态持续时长
	duration := int(time.Since(status.StartTime).Minutes())

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":           status.ID,
			"device_id":    status.DeviceID,
			"device_name":  status.Device.DeviceName,
			"status":       status.Status,
			"status_name":  getStatusName(status.Status),
			"start_time":   status.StartTime,
			"duration_min": duration,
			"remark":       status.Remark,
		},
	})
}

// GetDeviceStatusHistory 获取设备状态历史
// GET /api/device-status/:device_id/history
func GetDeviceStatusHistory(c *gin.Context) {
	deviceIDStr := c.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "设备ID格式错误"})
		return
	}

	// 解析查询参数（使用本地时区）
	var startTime, endTime *time.Time
	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.ParseInLocation(time.RFC3339, startStr, time.Local); err == nil {
			startTime = &t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.ParseInLocation(time.RFC3339, endStr, time.Local); err == nil {
			endTime = &t
		}
	}

	statusList, err := database.GetDeviceStatusHistory(deviceID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 格式化响应数据
	formattedList := make([]gin.H, 0, len(statusList))
	for _, status := range statusList {
		duration := 0
		if status.EndTime != nil {
			duration = int(status.EndTime.Sub(status.StartTime).Minutes())
		} else {
			duration = int(time.Since(status.StartTime).Minutes())
		}

		formattedList = append(formattedList, gin.H{
			"id":           status.ID,
			"device_id":    status.DeviceID,
			"device_name":  status.Device.DeviceName,
			"status":       status.Status,
			"status_name":  getStatusName(status.Status),
			"start_time":   status.StartTime,
			"end_time":     status.EndTime,
			"duration_min": duration,
			"remark":       status.Remark,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  formattedList,
		"count": len(formattedList),
	})
}

// GetAllDevicesStatus 获取所有设备当前状态
// GET /api/device-status/all
func GetAllDevicesStatus(c *gin.Context) {
	statusList, err := database.GetAllDevicesStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 格式化响应数据
	formattedList := make([]gin.H, 0, len(statusList))
	for _, status := range statusList {
		duration := int(time.Since(status.StartTime).Minutes())

		formattedList = append(formattedList, gin.H{
			"id":           status.ID,
			"device_id":    status.DeviceID,
			"device_name":  status.Device.DeviceName,
			"status":       status.Status,
			"status_name":  getStatusName(status.Status),
			"start_time":   status.StartTime,
			"duration_min": duration,
			"remark":       status.Remark,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  formattedList,
		"count": len(formattedList),
	})
}

// GetDeviceStatusSummary 获取设备状态统计
// GET /api/device-status/:device_id/summary
func GetDeviceStatusSummary(c *gin.Context) {
	deviceIDStr := c.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "设备ID格式错误"})
		return
	}

	summary, err := database.GetDeviceStatusSummary(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": summary,
	})
}

// GetAllDevicesStatusSummary 获取所有设备状态统计
// GET /api/device-status/summary
func GetAllDevicesStatusSummary(c *gin.Context) {
	summaries, err := database.GetAllDevicesStatusSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  summaries,
		"count": len(summaries),
	})
}

// EndDeviceStatus 结束设备当前状态
// POST /api/device-status/:device_id/end
func EndDeviceStatus(c *gin.Context) {
	deviceIDStr := c.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "设备ID格式错误"})
		return
	}

	var req struct {
		Remark *string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	if err := database.EndDeviceStatus(deviceID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "设备状态已结束",
	})
}

// GetStatusStatistics 获取设备状态统计（指定时间范围）
// GET /api/device-status/:device_id/statistics
func GetStatusStatistics(c *gin.Context) {
	deviceIDStr := c.Param("device_id")
	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "设备ID格式错误"})
		return
	}

	// 解析时间参数，默认为今天
	startTime := time.Now().Truncate(24 * time.Hour)
	endTime := time.Now()

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.ParseInLocation(time.RFC3339, startStr, time.Local); err == nil {
			startTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.ParseInLocation(time.RFC3339, endStr, time.Local); err == nil {
			endTime = t
		}
	}

	statistics, err := database.GetStatusStatistics(deviceID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": statistics,
	})
}

// ========================================================
// 辅助函数
// ========================================================

// getStatusName 获取状态名称
func getStatusName(status int8) string {
	statusNames := map[int8]string{
		0: "空闲",
		1: "运行",
		2: "故障",
	}
	if name, exists := statusNames[status]; exists {
		return name
	}
	return "未知"
}






