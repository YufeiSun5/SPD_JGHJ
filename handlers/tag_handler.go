package handlers

import (
	"net/http"
	"strconv"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"

	"github.com/gin-gonic/gin"
)

// ========================================================
// IOT测点管理 API (sys_variables)
// ========================================================

// GetRealtimeValue 获取测点实时值
// GET /api/v1/tags/realtime/:var_id
func GetRealtimeValue(c *gin.Context) {
	varID, err := strconv.ParseInt(c.Param("var_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的变量ID"})
		return
	}

	tagManager := core.GetTagManager()
	tag, exists := tagManager.GetTag(varID)

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "测点不存在",
		})
		return
	}

	// 根据数据类型返回不同字段
	response := gin.H{
		"var_name":         tag.VarName,
		"display_name":     tag.DisplayName,
		"data_type":        tag.DataType,
		"unit":             tag.Unit,
		"last_update_time": tag.GetLastUpdateTime().Format("2006-01-02 15:04:05"),
	}

	if tag.DataType == "STRING" || tag.DataType == "TEXT" {
		response["value"] = tag.GetStringValue()
	} else {
		response["value"] = tag.GetValue()
		response["alarm_state"] = tag.AlarmState
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// GetAllRealtimeValues 获取所有测点实时值
// GET /api/v1/tags/realtime
func GetAllRealtimeValues(c *gin.Context) {
	tagManager := core.GetTagManager()
	allTags := tagManager.GetAllTags()

	response := make([]gin.H, 0, len(allTags))
	for _, tag := range allTags {
		item := gin.H{
			"var_id":           tag.VarID, // 添加ID
			"var_name":         tag.VarName,
			"display_name":     tag.DisplayName,
			"data_type":        tag.DataType,
			"unit":             tag.Unit,
			"store_mode":       tag.StoreMode, // 添加存储模式
			"last_update_time": tag.GetLastUpdateTime().Format("2006-01-02 15:04:05"),
		}

		if tag.DataType == "STRING" || tag.DataType == "TEXT" {
			item["value"] = tag.GetStringValue()
		} else {
			item["value"] = tag.GetValue()
			item["alarm_state"] = tag.AlarmState
		}

		response = append(response, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  response,
		"total": len(response),
	})
}

// GetHistoryData 获取测点历史数据
// GET /api/v1/tags/history/:var_id?start_time=2025-01-01&end_time=2025-12-31&limit=1000
func GetHistoryData(c *gin.Context) {
	varID, err := strconv.ParseInt(c.Param("var_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的测点ID",
		})
		return
	}

	// 解析时间参数
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	limitStr := c.DefaultQuery("limit", "1000")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 10000 {
		limit = 1000 // 默认最多返回1000条
	}

	// 构建查询
	query := database.DB.Table("sys_data_history").
		Where("var_id = ?", varID).
		Order("created_at DESC").
		Limit(limit)

	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	// 执行查询
	var results []struct {
		Val       *float64  `gorm:"column:val"`
		StrVal    *string   `gorm:"column:str_val"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}

	if err := query.Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询历史数据失败",
			"details": err.Error(),
		})
		return
	}

	// 格式化返回数据
	data := make([]gin.H, len(results))
	for i, row := range results {
		item := gin.H{
			"timestamp": row.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if row.StrVal != nil {
			item["value"] = *row.StrVal
			item["type"] = "string"
		} else if row.Val != nil {
			item["value"] = *row.Val
			item["type"] = "numeric"
		}
		data[i] = item
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   data,
		"total":  len(data),
		"var_id": varID,
	})
}

// GetAlarmRecords 获取报警记录
// GET /api/v1/alarms?var_id=1&start_time=2025-01-01&ack_status=0
func GetAlarmRecords(c *gin.Context) {
	query := database.DB.Table("sys_alarm_records").Order("start_time DESC")

	// 筛选条件
	if varIDStr := c.Query("var_id"); varIDStr != "" {
		varID, err := strconv.ParseInt(varIDStr, 10, 64)
		if err == nil {
			query = query.Where("var_id = ?", varID)
		}
	}

	if startTime := c.Query("start_time"); startTime != "" {
		query = query.Where("start_time >= ?", startTime)
	}

	if endTime := c.Query("end_time"); endTime != "" {
		query = query.Where("start_time <= ?", endTime)
	}

	if ackStatusStr := c.Query("ack_status"); ackStatusStr != "" {
		ackStatus, err := strconv.Atoi(ackStatusStr)
		if err == nil {
			query = query.Where("ack_status = ?", ackStatus)
		}
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize > 200 {
		pageSize = 200
	}

	offset := (page - 1) * pageSize

	var total int64
	query.Count(&total)

	var records []database.AlarmRecord
	if err := query.Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询报警记录失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// AckAlarm 确认报警
// POST /api/v1/alarms/:id/ack
func AckAlarm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的报警记录ID",
		})
		return
	}

	result := database.DB.Model(&database.AlarmRecord{}).
		Where("id = ?", id).
		Update("ack_status", 1)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "确认报警失败",
			"details": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "报警记录不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "报警已确认",
	})
}

// GetTagConfig 获取测点配置
// GET /api/v1/tags/config/:var_id
func GetTagConfig(c *gin.Context) {
	varID, err := strconv.ParseInt(c.Param("var_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的变量ID"})
		return
	}

	tagManager := core.GetTagManager()
	tag, exists := tagManager.GetTag(varID)

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "测点不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tag,
	})
}

// ReloadTags 重新加载测点配置
// POST /api/v1/tags/reload
func ReloadTags(c *gin.Context) {
	tags, err := database.LoadVariables()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "加载测点配置失败",
			"details": err.Error(),
		})
		return
	}

	tagManager := core.GetTagManager()
	tagManager.LoadTags(tags)

	c.JSON(http.StatusOK, gin.H{
		"message": "测点配置已重新加载",
		"total":   len(tags),
	})
}
