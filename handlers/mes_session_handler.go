package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 设备登录/班次管理 API (pro_machine_sessions)
// ========================================================

// DeviceLogin 设备登录/上班打卡
// POST /api/v1/sessions/login
func DeviceLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 验证员工列表不能为空
	if len(req.StaffIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "当班员工列表不能为空",
		})
		return
	}

	// 序列化员工ID列表为JSON
	staffIDsJSON, err := json.Marshal(req.StaffIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "序列化员工列表失败",
			"details": err.Error(),
		})
		return
	}

	session := &models.ProMachineSession{
		DeviceID: req.DeviceID,
		TeamID:   req.TeamID,
		StaffIDs: string(staffIDsJSON),
	}

	if err := database.DeviceLogin(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "登录失败",
			"details": err.Error(),
		})
		return
	}

	// 返回包含班组信息的记录
	session, _ = database.GetSessionByID(session.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "上班签到成功",
		"data":    session,
	})
}

// DeviceLogout 设备登出/下班打卡
// POST /api/v1/sessions/logout
func DeviceLogout(c *gin.Context) {
	var req models.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	session, err := database.DeviceLogout(req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "登出失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "下班签退成功",
		"data":    session,
	})
}

// GetActiveSession 获取设备当前活动班次
// GET /api/v1/sessions/active/:device_id
func GetActiveSession(c *gin.Context) {
	deviceID, err := strconv.Atoi(c.Param("device_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的设备ID",
		})
		return
	}

	session, err := database.GetActiveSession(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "该设备当前没有活动班次",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": session,
	})
}

// GetSessionStats 获取班次统计信息
// GET /api/v1/sessions/:id/stats
func GetSessionStats(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班次ID",
		})
		return
	}

	stats, err := database.GetSessionStats(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取统计信息失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}

// GetSessionHistory 获取班次历史记录
// GET /api/v1/sessions/history?device_id=1&team_id=2&start_date=2025-12-01&end_date=2025-12-31
func GetSessionHistory(c *gin.Context) {
	var deviceID, teamID *int
	var startDate, endDate *time.Time

	// 解析设备ID
	if deviceIDStr := c.Query("device_id"); deviceIDStr != "" {
		val, err := strconv.Atoi(deviceIDStr)
		if err == nil {
			deviceID = &val
		}
	}

	// 解析班组ID
	if teamIDStr := c.Query("team_id"); teamIDStr != "" {
		val, err := strconv.Atoi(teamIDStr)
		if err == nil {
			teamID = &val
		}
	}

	// 解析开始日期（使用本地时区）
	if startStr := c.Query("start_date"); startStr != "" {
		val, err := time.ParseInLocation("2006-01-02", startStr, time.Local)
		if err == nil {
			startDate = &val
		}
	}

	// 解析结束日期（使用本地时区）
	if endStr := c.Query("end_date"); endStr != "" {
		val, err := time.ParseInLocation("2006-01-02", endStr, time.Local)
		if err == nil {
			// 设置为当天的 23:59:59
			endTime := val.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &endTime
		}
	}

	sessions, err := database.GetSessionHistory(deviceID, teamID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询班次历史失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  sessions,
		"total": len(sessions),
	})
}

// GetStaffAttendance 获取员工出勤记录
// GET /api/v1/sessions/attendance/:staff_id?start_date=2025-12-01&end_date=2025-12-31
func GetStaffAttendance(c *gin.Context) {
	staffID, err := strconv.Atoi(c.Param("staff_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的员工ID",
		})
		return
	}

	var startDate, endDate *time.Time

	// 解析日期范围（使用本地时区）
	if startStr := c.Query("start_date"); startStr != "" {
		val, err := time.ParseInLocation("2006-01-02", startStr, time.Local)
		if err == nil {
			startDate = &val
		}
	}

	if endStr := c.Query("end_date"); endStr != "" {
		val, err := time.ParseInLocation("2006-01-02", endStr, time.Local)
		if err == nil {
			endTime := val.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &endTime
		}
	}

	sessions, err := database.GetStaffAttendance(staffID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询出勤记录失败",
			"details": err.Error(),
		})
		return
	}

	// 统计出勤信息
	totalSessions := len(sessions)
	totalMinutes := 0
	for _, session := range sessions {
		totalMinutes += session.DurationMin
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sessions,
		"stats": gin.H{
			"total_sessions": totalSessions,
			"total_hours":    float64(totalMinutes) / 60.0,
			"total_days":     totalSessions, // 简化计算，每个 session 算一天
		},
	})
}

// UpdateSessionStaff 更新班次的员工列表
// PUT /api/v1/sessions/:id/staff
func UpdateSessionStaff(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班次ID",
		})
		return
	}

	var req struct {
		StaffIDs []int `json:"staff_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	if err := database.UpdateSessionStaff(id, req.StaffIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新员工列表失败",
			"details": err.Error(),
		})
		return
	}

	// 返回更新后的记录
	session, _ := database.GetSessionByID(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "更新员工列表成功",
		"data":    session,
	})
}
