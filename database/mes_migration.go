package database

import (
	"fmt"
	"log"

	"gin-mqtt-pgsql/models"
)

// MigrateMESTables 创建MES系统所有表（MySQL优化版）
func MigrateMESTables() error {
	log.Println("[Database] 开始创建MES系统表结构...")

	// CN: AutoMigrate 只追加列，不删除，安全幂等。
	//     SysShiftSchedule 必须先于 SysShift（因为 SysShift.ScheduleID 是外键）。
	//     SysDevice 追加 schedule_id 列（设备关联时间安排组，多对一）。
	// EN: AutoMigrate only adds columns, never drops. SysShiftSchedule must precede SysShift
	//     (SysShift.ScheduleID FK). SysDevice gains schedule_id (many devices → one schedule).
	// JP: AutoMigrate は列追加のみ。SysShiftSchedule は外部キーのため SysShift より先に処理。
	//     SysDevice には schedule_id 列（多対一でスケジュールに紐づく）を追加。
	if err := DB.AutoMigrate(
		&models.SysShiftSchedule{}, // 时间安排组（新增表，须在 SysShift 之前）
		&models.SysShift{},         // 班次配置（追加 schedule_id 列）
		&models.SysShiftBreak{},    // 班次内休息时间段
		&models.SysDevice{},        // 设备表（追加 schedule_id 列）
		&models.SysDeviceStatus{},  // 设备状态表
		&models.SysTeam{},          // 班组表
		&models.SysStaff{},         // 员工表
		&models.SysStaffHistory{},  // 员工调动历史
		&models.ProOrder{},         // 工单表
		&models.ProProductionRun{}, // 生产运行记录
		&models.ProMachineSession{},  // 设备登录/班次记录表
		&models.ProShiftSnapshot{},  // 班次生产快照（追溯用）
		&models.Task{},              // 任务配置表
		&models.TaskExecutionLog{},  // 任务执行日志表
	); err != nil {
		return fmt.Errorf("创建MES表结构失败: %w", err)
	}

	// 数据迁移：若无时间安排组，把已有班次归入一个默认组
	// CN: 首次部署新版本时自动执行，防止历史班次游离（schedule_id=0）。
	// EN: Runs once on first new-version deploy; prevents orphan shifts (schedule_id=0).
	// JP: 新バージョン初回デプロイ時に1度実行し、孤立シフト（schedule_id=0）を防ぐ。
	var scheduleCount int64
	DB.Model(&models.SysShiftSchedule{}).Count(&scheduleCount)
	if scheduleCount == 0 {
		var orphanCount int64
		DB.Model(&models.SysShift{}).Where("schedule_id = 0 OR schedule_id IS NULL").Count(&orphanCount)
		if orphanCount > 0 {
			defaultSched := models.SysShiftSchedule{Name: "默认时间安排", SortOrder: 0, IsActive: true}
			if err := DB.Create(&defaultSched).Error; err == nil {
				DB.Model(&models.SysShift{}).
					Where("schedule_id = 0 OR schedule_id IS NULL").
					Update("schedule_id", defaultSched.ID)
				log.Printf("[Database] ✅ 已将 %d 个历史班次迁移至「%s」(id=%d)", orphanCount, defaultSched.Name, defaultSched.ID)
			}
		}
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
