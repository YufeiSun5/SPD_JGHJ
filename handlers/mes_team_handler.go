package handlers

import (
	"net/http"
	"strconv"

	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"

	"github.com/gin-gonic/gin"
)

// ========================================================
// 班组管理 API (sys_teams)
// ========================================================

// CreateTeam 创建班组
// POST /api/v1/teams
func CreateTeam(c *gin.Context) {
	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	team := &models.SysTeam{
		TeamName:   req.TeamName,
		LeaderName: req.LeaderName,
		Status:     1, // 默认启用
	}

	if err := database.CreateTeam(team); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "创建班组失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "创建班组成功",
		"data":    team,
	})
}

// GetTeam 获取单个班组
// GET /api/v1/teams/:id
func GetTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班组ID",
		})
		return
	}

	team, err := database.GetTeamByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "班组不存在",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": team,
	})
}

// GetAllTeams 获取所有班组
// GET /api/v1/teams?status=1
func GetAllTeams(c *gin.Context) {
	var status *int8
	if statusStr := c.Query("status"); statusStr != "" {
		val, err := strconv.ParseInt(statusStr, 10, 8)
		if err == nil {
			s := int8(val)
			status = &s
		}
	}

	teams, err := database.GetAllTeams(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "查询班组列表失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  teams,
		"total": len(teams),
	})
}

// UpdateTeam 更新班组
// PUT /api/v1/teams/:id
func UpdateTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班组ID",
		})
		return
	}

	var req models.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"details": err.Error(),
		})
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.TeamName != nil {
		updates["team_name"] = *req.TeamName
	}
	if req.LeaderName != nil {
		updates["leader_name"] = *req.LeaderName
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

	if err := database.UpdateTeam(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "更新班组失败",
			"details": err.Error(),
		})
		return
	}

	// 返回更新后的数据
	team, _ := database.GetTeamByID(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "更新班组成功",
		"data":    team,
	})
}

// DeleteTeam 删除班组
// DELETE /api/v1/teams/:id
func DeleteTeam(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的班组ID",
		})
		return
	}

	if err := database.DeleteTeam(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "删除班组失败",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "删除班组成功",
	})
}
