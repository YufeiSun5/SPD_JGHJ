package handlers

import (
	"net/http"
	"strconv"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/workers"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 统计分析 API
// ========================================================

// GetHourlyProduction 获取今日按小时统计的产量
// GET /api/v1/statistics/hourly-production?device_id=1
func GetHourlyProduction(c *gin.Context) {
	var deviceID *int
	if deviceIDStr := c.Query("device_id"); deviceIDStr != "" {
		id, err := strconv.Atoi(deviceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的设备ID",
			})
			return
		}
		deviceID = &id
	}

	results, err := database.GetHourlyProduction(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询小时产量失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"total": len(results),
	})
}

// GetStaffEfficiency 获取员工绩效统计
// GET /api/v1/statistics/staff-efficiency?start_time=2025-12-01&end_time=2025-12-31
func GetStaffEfficiency(c *gin.Context) {
	var startTime, endTime *time.Time

	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", startTimeStr, time.Local); err == nil {
			startTime = &t
		}
	}

	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", endTimeStr, time.Local); err == nil {
			endTime = &t
		}
	}

	results, err := database.GetStaffEfficiency(startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询员工绩效失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"total": len(results),
	})
}

// GetDeviceUtilizationTrend 获取设备利用率趋势
// GET /api/v1/statistics/utilization-trend?device_id=1
func GetDeviceUtilizationTrend(c *gin.Context) {
	var deviceID *int
	if deviceIDStr := c.Query("device_id"); deviceIDStr != "" {
		id, err := strconv.Atoi(deviceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的设备ID",
			})
			return
		}
		deviceID = &id
	}

	results, err := database.GetDeviceUtilizationTrend(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询利用率趋势失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  results,
		"total": len(results),
	})
}

// GetSystemMonitor 获取系统监控数据 (通道队列长度、任务统计)
// GET /api/v1/statistics/system-monitor
func GetSystemMonitor(c *gin.Context) {
	// 获取通道状态
	channelStats := core.GetChannelStats()

	// 获取任务统计
	scheduler := workers.GetTaskScheduler()
	var taskStats map[string]int
	if scheduler != nil {
		taskStats = scheduler.GetTaskCount()
	} else {
		taskStats = map[string]int{
			"scheduled":   0,
			"data_change": 0,
			"condition":   0,
		}
	}

	// 计算通道容量和使用率
	channelCapacity := map[string]int{
		"logic_chan":   2000,
		"change_chan":  500,
		"cycle_chan":   500,
		"alarm_chan":   200,
		"sse_chan":     100,
		"event_chan":   300,
		"trigger_chan": 10000, // 🔥 超大缓冲，不丢弃任务
	}

	channelUsage := make(map[string]float64)
	for name, current := range channelStats {
		capacity := channelCapacity[name]
		if capacity > 0 {
			channelUsage[name] = float64(current) / float64(capacity) * 100
		}
	}

	// 使用核心模块的健康检查
	alerts := core.CheckChannelHealth()

	c.JSON(http.StatusOK, gin.H{
		"channel_stats":    channelStats,
		"channel_capacity": channelCapacity,
		"channel_usage":    channelUsage,
		"task_stats":       taskStats,
		"alerts":           alerts,
		"timestamp":        time.Now(),
	})
}
