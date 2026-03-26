package config

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config 应用程序配置结构
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	MQTT     MQTTConfig     `json:"mqtt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `json:"port"`
	Mode string `json:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`
}

// MQTTConfig MQTT配置
type MQTTConfig struct {
	Broker               string `json:"broker"`
	Port                 int    `json:"port"`
	ClientID             string `json:"client_id"`
	Username             string `json:"username"`
	Password             string `json:"password"`
	KeepAlive            int    `json:"keep_alive"`             // 心跳间隔(秒)
	CleanSession         bool   `json:"clean_session"`          // 清理会话
	AutoReconnect        bool   `json:"auto_reconnect"`         // 自动重连
	MaxReconnectInterval int    `json:"max_reconnect_interval"` // 最大重连间隔(秒)
}

// LoadConfig 从环境变量或配置文件加载配置
func LoadConfig() *Config {
	// 尝试从 config.env 文件加载环境变量
	loadEnvFile("config.env")

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "gin_app"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		MQTT: MQTTConfig{
			Broker:               getEnv("MQTT_BROKER", "tcp://127.0.0.1"),
			Port:                 getEnvAsInt("MQTT_PORT", 1883),
			ClientID:             getEnv("MQTT_CLIENT_ID", "iot-gateway-client"),
			Username:             getEnv("MQTT_USERNAME", "root"),
			Password:             getEnv("MQTT_PASSWORD", ""),
			KeepAlive:            getEnvAsInt("MQTT_KEEP_ALIVE", 10),
			CleanSession:         getEnvAsBool("MQTT_CLEAN_SESSION", false),
			AutoReconnect:        getEnvAsBool("MQTT_AUTO_RECONNECT", true),
			MaxReconnectInterval: getEnvAsInt("MQTT_MAX_RECONNECT_INTERVAL", 30),
		},
	}

	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool 获取环境变量并转换为布尔值，如果不存在或转换失败则返回默认值
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// LoadEnvFileFromPath 从指定路径加载环境变量 (导出函数)
func LoadEnvFileFromPath(filename string) {
	loadEnvFile(filename)
}

// loadEnvFile 从文件加载环境变量
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// 文件不存在时使用默认配置
		return
	}
	defer file.Close()

	log.Printf("[配置] 从 %s 加载配置...", filename)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 解析 KEY=VALUE 格式
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			log.Printf("[配置] 跳过无效行 %d: %s", lineNum, line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 设置环境变量 (如果未设置)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[配置] 读取文件失败: %v", err)
	}
}
