package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gin-mqtt-pgsql/database"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 配置管理 API
// ========================================================

// GetAllVariables 获取所有变量配置
// GET /api/v1/config/variables
func GetAllVariables(c *gin.Context) {
	var variables []database.VariableRow

	result := database.DB.Table("sys_variables").Order("id").Find(&variables)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询变量配置失败",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  variables,
		"total": len(variables),
	})
}

// UpdateVariable 更新单个变量配置
// PUT /api/v1/config/variables/:id
func UpdateVariable(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的变量ID",
		})
		return
	}

	var updateData database.VariableRow
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	// 更新数据库
	result := database.DB.Table("sys_variables").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"device_id":      updateData.DeviceID,
			"var_name":       updateData.VarName,
			"display_name":   updateData.DisplayName,
			"json_path":      updateData.JSONPath,
			"data_type":      updateData.DataType,
			"rw_mode":        updateData.RWMode,
			"unit":           updateData.Unit,
			"scale_factor":   updateData.ScaleFactor,
			"offset_val":     updateData.OffsetVal,
			"store_mode":     updateData.StoreMode,
			"store_cycle":    updateData.StoreCycle,
			"store_deadband": updateData.StoreDeadband,
			"alarm_enable":   updateData.AlarmEnable,
			"limit_hh":       updateData.LimitHH,
			"limit_h":        updateData.LimitH,
			"limit_l":        updateData.LimitL,
			"limit_ll":       updateData.LimitLL,
			"deadband":       updateData.Deadband,
			"alarm_msg":      updateData.AlarmMsg,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新变量配置失败",
			"details": result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "变量不存在",
		})
		return
	}

	// 更新配置版本以触发热重载
	newVersion := fmt.Sprintf("v%d", time.Now().Unix())
	if err := database.UpdateConfigVersion(newVersion); err != nil {
		// 记录错误但不影响主流程
		c.JSON(http.StatusOK, gin.H{
			"message": "变量配置已更新，但触发热重载失败",
			"warning": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "变量配置已更新",
	})
}

// BatchUpdateVariables 批量更新变量配置
// PUT /api/v1/config/variables/batch
func BatchUpdateVariables(c *gin.Context) {
	var request struct {
		Variables []database.VariableRow `json:"variables"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		return
	}

	if len(request.Variables) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "没有要更新的变量",
		})
		return
	}

	// 使用事务批量更新
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, variable := range request.Variables {
		result := tx.Table("sys_variables").
			Where("id = ?", variable.ID).
			Updates(map[string]interface{}{
				"device_id":      variable.DeviceID,
				"var_name":       variable.VarName,
				"display_name":   variable.DisplayName,
				"json_path":      variable.JSONPath,
				"data_type":      variable.DataType,
				"rw_mode":        variable.RWMode,
				"unit":           variable.Unit,
				"scale_factor":   variable.ScaleFactor,
				"offset_val":     variable.OffsetVal,
				"store_mode":     variable.StoreMode,
				"store_cycle":    variable.StoreCycle,
				"store_deadband": variable.StoreDeadband,
				"alarm_enable":   variable.AlarmEnable,
				"limit_hh":       variable.LimitHH,
				"limit_h":        variable.LimitH,
				"limit_l":        variable.LimitL,
				"limit_ll":       variable.LimitLL,
				"deadband":       variable.Deadband,
				"alarm_msg":      variable.AlarmMsg,
			})

		if result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "批量更新失败",
				"details": result.Error.Error(),
			})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "提交事务失败",
			"details": err.Error(),
		})
		return
	}

	// 更新配置版本以触发热重载
	newVersion := fmt.Sprintf("v%d", time.Now().Unix())
	if err := database.UpdateConfigVersion(newVersion); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "批量更新成功，但触发热重载失败",
			"warning": err.Error(),
			"count":   len(request.Variables),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "批量更新成功",
		"count":   len(request.Variables),
	})
}

// GetGateways 获取所有网关配置
// GET /api/v1/config/gateways
func GetGateways(c *gin.Context) {
	var gateways []database.GatewayConfig

	result := database.DB.Table("sys_gateways").Order("id").Find(&gateways)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询网关配置失败",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  gateways,
		"total": len(gateways),
	})
}

// GetDevices 获取所有设备配置
// GET /api/v1/config/devices
func GetDevices(c *gin.Context) {
	var devices []struct {
		ID        int    `gorm:"column:id"`
		GatewayID int    `gorm:"column:gateway_id"`
		DevName   string `gorm:"column:dev_name"`
		DevType   string `gorm:"column:dev_type"`
		Status    int    `gorm:"column:status"`
	}

	result := database.DB.Table("sys_devices").Order("id").Find(&devices)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询设备配置失败",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  devices,
		"total": len(devices),
	})
}





