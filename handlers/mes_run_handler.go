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
// 生产运行记录 API (pro_production_runs)
// ========================================================

// StartProduction 开始生产
// POST /api/v1/production/start
func StartProduction(c *gin.Context) {
	var req models.StartProductionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	run, err := database.StartProduction(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "开始生产失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "开始生产成功",
		"data":    run,
	})
}

// GetProductionRun 获取单个运行记录
// GET /api/v1/production/runs/:id
func GetProductionRun(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的运行记录ID",
		})
		return
	}

	run, err := database.GetProductionRunByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "运行记录不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": run,
	})
}

// GetActiveRun 获取设备当前运行记录
// GET /api/v1/production/active/:device_id
func GetActiveRun(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("device_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备ID",
		})
		return
	}

	run, err := database.GetActiveRun(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "该设备当前没有运行记录",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": run,
	})
}

// UpdateProductionRun 更新运行记录产量
// PUT /api/v1/production/runs/:id
func UpdateProductionRun(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的运行记录ID",
		})
		return
	}

	var req models.UpdateProductionRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	okQtyDelta := 0
	ngQtyDelta := 0

	if req.RunOkQty != nil {
		okQtyDelta = *req.RunOkQty
	}
	if req.RunNgQty != nil {
		ngQtyDelta = *req.RunNgQty
	}

	if okQtyDelta == 0 && ngQtyDelta == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "至少需要更新良品或不良品数量",
		})
		return
	}

	if err := database.UpdateProductionRun(id, okQtyDelta, ngQtyDelta); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新运行记录失败",
			"details": err.Error(),
		})
		return
	}

	// 返回更新后的数据
	run, _ := database.GetProductionRunByID(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "更新运行记录成功",
		"data":    run,
	})
}

// EndProduction 结束生产
// POST /api/v1/production/end
func EndProduction(c *gin.Context) {
	var req models.EndProductionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	if err := database.EndProduction(req.RunID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "结束生产失败",
			"details": err.Error(),
		})
		return
	}

	// 返回更新后的数据
	run, _ := database.GetProductionRunByID(req.RunID)
	c.JSON(http.StatusOK, gin.H{
		"message": "结束生产成功",
		"data":    run,
	})
}

// GetDeviceRuns 获取设备运行历史
// GET /api/v1/production/device/:device_id/runs?start_time=2025-01-01&end_time=2025-12-31
func GetDeviceRuns(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("device_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备ID",
		})
		return
	}

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

	runs, err := database.GetDeviceRuns(deviceID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询设备运行历史失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  runs,
		"total": len(runs),
	})
}
