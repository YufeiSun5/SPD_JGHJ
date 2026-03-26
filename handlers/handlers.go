package handlers

import (
	"encoding/json"
	"net/http"

	"gin-mqtt-pgsql/mqtt"

	"github.com/gin-gonic/gin"
)

// MQTTPublishRequest MQTT发布请求结构
type MQTTPublishRequest struct {
	Topic   string      `json:"topic" binding:"required"`
	Message interface{} `json:"message" binding:"required"`
}

// SetDataRequest KingIOT写入数据请求结构
type SetDataRequest struct {
	Writer    string                   `json:"Writer" binding:"required"`
	WriteTime string                   `json:"WriteTime" binding:"required"`
	Username  string                   `json:"Username" binding:"required"`
	Password  string                   `json:"Password" binding:"required"`
	Qid       int64                    `json:"Qid" binding:"required"`
	PNs       map[string]string        `json:"PNs" binding:"required"`
	PVs       map[string]interface{}   `json:"PVs" binding:"required"`
	Objs      []map[string]interface{} `json:"Objs" binding:"required"`
}

// KingIOTWriteRequest KingIOT通用写入请求
type KingIOTWriteRequest struct {
	ProjectName string         `json:"project_name" binding:"required"` // 项目名称，如 "S_KIO_Project"
	Data        SetDataRequest `json:"data" binding:"required"`
}

// PublishMQTTMessage 发布MQTT消息的API处理器
func PublishMQTTMessage(c *gin.Context) {
	var req MQTTPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := mqtt.Publish(req.Topic, req.Message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to publish message",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message published successfully",
		"topic":   req.Topic,
	})
}

// GetStatus 获取系统状态
func GetStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "running",
		"services": gin.H{
			"database": "connected",
			"mqtt":     "connected",
		},
	})
}

// WriteKingIOTData 向KingIOT项目写入数据
func WriteKingIOTData(c *gin.Context) {
	var req KingIOTWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// 构建主题名称: setdata_{ProjectName}
	topic := "setdata_" + req.ProjectName

	// 将数据序列化为JSON
	payload, err := json.Marshal(req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to serialize data",
			"details": err.Error(),
		})
		return
	}

	// 发布到MQTT
	if err := mqtt.Publish(topic, string(payload)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to publish to MQTT",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Data written successfully",
		"topic":        topic,
		"project_name": req.ProjectName,
		"qid":          req.Data.Qid,
	})
}

// SetKingIOTDataDirect 直接写入KingIOT数据（简化版，直接接收完整payload）
func SetKingIOTDataDirect(c *gin.Context) {
	// 获取项目名称（从URL参数或查询参数）
	projectName := c.Param("project")
	if projectName == "" {
		projectName = c.Query("project")
	}
	if projectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Project name is required",
		})
		return
	}

	// 读取原始JSON数据
	var data SetDataRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid data format",
			"details": err.Error(),
		})
		return
	}

	// 构建主题名称
	topic := "setdata_" + projectName

	// 序列化数据
	payload, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to serialize data",
			"details": err.Error(),
		})
		return
	}

	// 发布到MQTT
	if err := mqtt.Publish(topic, string(payload)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to publish to MQTT",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Data written successfully",
		"topic":        topic,
		"project_name": projectName,
		"qid":          data.Qid,
		"writer":       data.Writer,
	})
}
