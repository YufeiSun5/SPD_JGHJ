package handlers

import (
	"net/http"
	"strconv"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 工单管理 API (pro_orders)
// ========================================================

// CreateOrder 创建工单
// POST /api/v1/orders
func CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	order := &models.ProOrder{
		OrderNo:        req.OrderNo,
		ProductCode:    req.ProductCode,
		TargetDeviceID: req.TargetDeviceID,
		PlanQty:        req.PlanQty,
		ActualQty:      0,
		OkQty:          0,
		NgQty:          0,
		Status:         0, // 待产
		Version:        0,
	}

	if err := database.CreateOrder(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建工单失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建工单成功",
		"data":    order,
	})
}

// GetOrder 获取单个工单
// GET /api/v1/orders/:id
func GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的工单ID",
		})
		return
	}

	order, err := database.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "工单不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": order,
	})
}

// GetOrderByNo 根据工单号获取工单
// GET /api/v1/orders/no/:order_no
func GetOrderByNo(c *gin.Context) {
	orderNo := c.Param("order_no")

	order, err := database.GetOrderByNo(orderNo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "工单不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": order,
	})
}

// GetAllOrders 获取所有工单
// GET /api/v1/orders?status=1&device_id=1
func GetAllOrders(c *gin.Context) {
	var status *int8
	var deviceID *int

	if statusStr := c.Query("status"); statusStr != "" {
		val, err := strconv.ParseInt(statusStr, 10, 8)
		if err == nil {
			s := int8(val)
			status = &s
		}
	}

	if deviceIDStr := c.Query("device_id"); deviceIDStr != "" {
		val, err := strconv.Atoi(deviceIDStr)
		if err == nil {
			deviceID = &val
		}
	}

	orders, err := database.GetAllOrders(status, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询工单列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  orders,
		"total": len(orders),
	})
}

// UpdateOrder 更新工单
// PUT /api/v1/orders/:id
func UpdateOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的工单ID",
		})
		return
	}

	var req models.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.ProductCode != nil {
		updates["product_code"] = *req.ProductCode
	}
	if req.TargetDeviceID != nil {
		updates["target_device_id"] = *req.TargetDeviceID
	}
	if req.PlanQty != nil {
		updates["plan_qty"] = *req.PlanQty
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "没有需要更新的字段",
		})
		return
	}

	if err := database.UpdateOrder(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新工单失败",
			"details": err.Error(),
		})
		return
	}

	// 返回更新后的数据
	order, _ := database.GetOrderByID(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "更新工单成功",
		"data":    order,
	})
}

// DeleteOrder 删除工单
// DELETE /api/v1/orders/:id
func DeleteOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的工单ID",
		})
		return
	}

	if err := database.DeleteOrder(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除工单失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除工单成功",
	})
}

// GetOrderSummary 获取工单汇总信息
// GET /api/v1/orders/:id/summary
func GetOrderSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的工单ID",
		})
		return
	}

	summary, err := database.GetOrderSummary(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取工单汇总失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": summary,
	})
}

// GetOrderRuns 获取工单的所有运行记录
// GET /api/v1/orders/:id/runs
func GetOrderRuns(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的工单ID",
		})
		return
	}

	runs, err := database.GetOrderRuns(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询运行记录失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  runs,
		"total": len(runs),
	})
}
