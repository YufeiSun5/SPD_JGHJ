package handlers

import (
	"net/http"
	"strconv"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 人员管理 API (sys_staff)
// ========================================================

// CreateStaff 创建员工
// POST /api/v1/staff
func CreateStaff(c *gin.Context) {
	var req models.CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	staff := &models.SysStaff{
		StaffCode:     req.StaffCode,
		Name:          req.Name,
		CurrentTeamID: req.CurrentTeamID,
		IsActive:      1, // 默认在职
	}

	if req.IsActive != nil {
		staff.IsActive = *req.IsActive
	}

	if err := database.CreateStaff(staff); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建员工失败",
			"details": err.Error(),
		})
		return
	}

	// 重新加载包含班组信息
	staff, _ = database.GetStaffByID(staff.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "创建员工成功",
		"data":    staff,
	})
}

// GetStaff 获取单个员工
// GET /api/v1/staff/:id
func GetStaff(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的员工ID",
		})
		return
	}

	staff, err := database.GetStaffByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "员工不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": staff,
	})
}

// GetStaffByCode 根据工号获取员工
// GET /api/v1/staff/code/:staff_code
func GetStaffByCode(c *gin.Context) {
	staffCode := c.Param("staff_code")

	staff, err := database.GetStaffByCode(staffCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "员工不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": staff,
	})
}

// GetAllStaff 获取所有员工
// GET /api/v1/staff?team_id=1&is_active=1
func GetAllStaff(c *gin.Context) {
	var teamID *int
	var isActive *int8

	if teamIDStr := c.Query("team_id"); teamIDStr != "" {
		val, err := strconv.Atoi(teamIDStr)
		if err == nil {
			teamID = &val
		}
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		val, err := strconv.ParseInt(isActiveStr, 10, 8)
		if err == nil {
			a := int8(val)
			isActive = &a
		}
	}

	staffList, err := database.GetAllStaff(teamID, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询员工列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  staffList,
		"total": len(staffList),
	})
}

// UpdateStaff 更新员工
// PUT /api/v1/staff/:id
func UpdateStaff(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的员工ID",
		})
		return
	}

	var req models.UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.CurrentTeamID != nil {
		updates["current_team_id"] = *req.CurrentTeamID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "没有需要更新的字段",
		})
		return
	}

	if err := database.UpdateStaff(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新员工失败",
			"details": err.Error(),
		})
		return
	}

	// 返回更新后的数据
	staff, _ := database.GetStaffByID(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "更新员工成功",
		"data":    staff,
	})
}

// DeleteStaff 删除员工
// DELETE /api/v1/staff/:id
func DeleteStaff(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的员工ID",
		})
		return
	}

	if err := database.DeleteStaff(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除员工失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除员工成功",
	})
}

// TransferStaff 调动员工
// POST /api/v1/staff/transfer
func TransferStaff(c *gin.Context) {
	var req models.TransferStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	if err := database.TransferStaff(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "员工调动失败",
			"details": err.Error(),
		})
		return
	}

	// 返回更新后的员工信息
	staff, _ := database.GetStaffByID(req.StaffID)
	c.JSON(http.StatusOK, gin.H{
		"message": "员工调动成功",
		"data":    staff,
	})
}

// GetStaffHistory 获取员工调动历史
// GET /api/v1/staff/:id/history
func GetStaffHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的员工ID",
		})
		return
	}

	history, err := database.GetStaffHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询员工历史失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  history,
		"total": len(history),
	})
}
