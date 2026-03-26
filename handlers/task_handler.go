package handlers

import (
	"net/http"
	"strconv"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"
	"gin-mqtt-pgsql/workers"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 任务管理 API (tasks)
// ========================================================

// GetAllTasks 获取所有任务
// GET /api/v1/tasks?is_enabled=1
func GetAllTasks(c *gin.Context) {
	query := database.DB.Model(&models.Task{})

	// 筛选条件
	if isEnabledStr := c.Query("is_enabled"); isEnabledStr != "" {
		isEnabled, err := strconv.ParseBool(isEnabledStr)
		if err == nil {
			query = query.Where("is_enabled = ?", isEnabled)
		}
	}

	if taskTypeStr := c.Query("task_type"); taskTypeStr != "" {
		taskType, err := strconv.Atoi(taskTypeStr)
		if err == nil {
			query = query.Where("task_type = ?", taskType)
		}
	}

	var tasks []models.Task
	if err := query.Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询任务列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tasks,
		"total": len(tasks),
	})
}

// GetTask 获取单个任务
// GET /api/v1/tasks/:id
func GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
		})
		return
	}

	task, err := database.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "任务不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": task,
	})
}

// CreateTask 创建任务
// POST /api/v1/tasks
func CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建任务失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建任务成功",
		"data":    task,
	})
}

// UpdateTask 更新任务
// PUT /api/v1/tasks/:id
func UpdateTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	result := database.DB.Model(&models.Task{}).Where("task_id = ?", id).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新任务失败",
			"details": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	// 返回更新后的数据
	task, _ := database.GetTaskByID(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "更新任务成功",
		"data":    task,
	})
}

// DeleteTask 删除任务
// DELETE /api/v1/tasks/:id
func DeleteTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
		})
		return
	}

	result := database.DB.Delete(&models.Task{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除任务失败",
			"details": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除任务成功",
	})
}

// EnableTask 启用任务
// POST /api/v1/tasks/:id/enable
func EnableTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
		})
		return
	}

	result := database.DB.Model(&models.Task{}).
		Where("task_id = ?", id).
		Update("is_enabled", true)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "启用任务失败",
			"details": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已启用",
	})
}

// DisableTask 禁用任务
// POST /api/v1/tasks/:id/disable
func DisableTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
		})
		return
	}

	result := database.DB.Model(&models.Task{}).
		Where("task_id = ?", id).
		Update("is_enabled", false)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "禁用任务失败",
			"details": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "任务不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已禁用",
	})
}

// GetTaskLogs 获取任务执行日志
// GET /api/v1/tasks/:id/logs?page=1&page_size=50
func GetTaskLogs(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
		})
		return
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
	database.DB.Model(&models.TaskExecutionLog{}).Where("task_id = ?", id).Count(&total)

	var logs []models.TaskExecutionLog
	if err := database.DB.Where("task_id = ?", id).
		Order("execute_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询任务日志失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      logs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ManualTriggerTask 手动触发任务
// POST /api/v1/tasks/:id/trigger
// 用于前端按钮直接触发任务，无需等待MQTT数据变动
func ManualTriggerTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务ID",
		})
		return
	}

	// 可选参数：网关ID和触发数据
	var req struct {
		GatewayID   int                    `json:"gateway_id"`   // 可选，默认0
		TriggerData map[string]interface{} `json:"trigger_data"` // 可选，自定义触发数据
	}
	c.ShouldBindJSON(&req)

	// 获取任务调度器实例
	scheduler := workers.GetTaskScheduler()
	if scheduler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "任务调度器未初始化",
		})
		return
	}

	// 手动触发任务
	err = scheduler.ManualTriggerTask(id, req.GatewayID, req.TriggerData)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "触发任务失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已成功触发",
		"task_id": id,
	})
}
