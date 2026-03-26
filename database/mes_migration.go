package database

import (
	"fmt"
	"log"

	"gin-mqtt-pgsql/models"
)

// MigrateMESTables 创建MES系统所有表（MySQL优化版）
func MigrateMESTables() error {
	log.Println("[Database] 开始创建MES系统表结构...")

	// 使用GORM的AutoMigrate自动创建表
	// GORM会自动处理MySQL的字段类型和索引
	// 注意：sys_devices 表已存在（IOT系统使用），不需要迁移
	if err := DB.AutoMigrate(
		&models.SysDeviceStatus{},   // 设备状态表
		&models.SysTeam{},           // 班组表
		&models.SysStaff{},          // 员工表
		&models.SysStaffHistory{},   // 员工调动历史
		&models.ProOrder{},          // 工单表
		&models.ProProductionRun{},  // 生产运行记录
		&models.ProMachineSession{}, // 设备登录/班次记录表
		&models.Task{},              // 任务配置表
		&models.TaskExecutionLog{},  // 任务执行日志表
	); err != nil {
		return fmt.Errorf("创建MES表结构失败: %w", err)
	}

	log.Println("[Database] ✅ MES系统表结构已创建")
	return nil
}

// InitMESDatabase 初始化MES数据库（仅创建表结构）
func InitMESDatabase() error {
	log.Println("[Database] 🔧 开始初始化MES数据库...")

	// 创建表结构（不插入示例数据，使用真实数据）
	if err := MigrateMESTables(); err != nil {
		return err
	}

	log.Println("[Database] ✅ MES数据库初始化完成")
	return nil
}
