package database

import (
	"fmt"
	"log"
	"time"

	"gin-mqtt-pgsql/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.DatabaseConfig) error {
	// MySQL DSN 格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var err error

	// 配置自定义 logger，设置慢查询阈值为 1 秒
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags), // 使用标准日志输出
		logger.Config{
			SlowThreshold:             1 * time.Second, // 慢查询阈值：1 秒
			LogLevel:                  logger.Warn,     // 日志级别：只记录警告和错误（包括慢查询）
			IgnoreRecordNotFoundError: true,            // 忽略 ErrRecordNotFound 错误
			Colorful:                  true,            // 彩色输出
		},
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 关闭自动创建外键
		DisableForeignKeyConstraintWhenMigrating: true,
		// 使用自定义 logger
		Logger: newLogger,
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	// 配置连接池参数
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 🎯 针对 IIoT 系统特点调优
	sqlDB.SetMaxOpenConns(50)                  // 限制最大连接数，防止 MySQL 崩溃
	sqlDB.SetMaxIdleConns(10)                  // 保持10个空闲连接，减少握手开销
	sqlDB.SetConnMaxLifetime(time.Hour)        // 1小时后重建连接，防止 MySQL 超时断开
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲10分钟的连接自动关闭

	log.Println("✅ 数据库连接成功 (MySQL)")
	log.Printf("📊 连接池配置: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		50, 10, time.Hour)
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
