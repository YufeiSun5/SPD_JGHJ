package core

import (
	"log"
	"time"

	"gin-mqtt-pgsql/database"
)

// ConfigReloader 配置热重载管理器
type ConfigReloader struct {
	lastVersion   string
	checkInterval time.Duration
	stopChan      chan struct{}
	isRunning     bool
}

var globalReloader *ConfigReloader

// InitConfigReloader 初始化配置重载器
func InitConfigReloader(checkInterval time.Duration) *ConfigReloader {
	globalReloader = &ConfigReloader{
		checkInterval: checkInterval,
		stopChan:      make(chan struct{}),
		isRunning:     false,
	}
	return globalReloader
}

// GetConfigReloader 获取全局配置重载器
func GetConfigReloader() *ConfigReloader {
	return globalReloader
}

// Start 启动配置监控协程
func (cr *ConfigReloader) Start() {
	if cr.isRunning {
		log.Println("[ConfigReloader] 配置监控已在运行")
		return
	}

	cr.isRunning = true
	log.Printf("[ConfigReloader] 启动配置监控协程 (检查间隔: %v)", cr.checkInterval)

	// 初始加载版本号
	version, err := cr.getCurrentVersion()
	if err != nil {
		log.Printf("[ConfigReloader] 无法获取当前版本: %v", err)
		cr.lastVersion = ""
	} else {
		cr.lastVersion = version
		log.Printf("[ConfigReloader] 当前配置版本: %s", version)
	}

	go cr.monitorLoop()
}

// Stop 停止配置监控
func (cr *ConfigReloader) Stop() {
	if !cr.isRunning {
		return
	}

	log.Println("[ConfigReloader] 停止配置监控...")
	close(cr.stopChan)
	cr.isRunning = false
}

// monitorLoop 监控循环
func (cr *ConfigReloader) monitorLoop() {
	ticker := time.NewTicker(cr.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cr.checkAndReload()
		case <-cr.stopChan:
			log.Println("[ConfigReloader] 监控协程已停止")
			return
		}
	}
}

// checkAndReload 检查版本并重载配置
func (cr *ConfigReloader) checkAndReload() {
	// 1. 获取最新版本号
	newVersion, err := cr.getCurrentVersion()
	if err != nil {
		log.Printf("[ConfigReloader] 获取版本失败: %v", err)
		return
	}

	// 2. 对比版本号
	if newVersion == cr.lastVersion {
		// 版本未变化，无需重载
		return
	}

	log.Printf("[ConfigReloader] 🔄 检测到配置更新: %s -> %s, 开始热重载...", cr.lastVersion, newVersion)

	// 3. 重载测点配置
	if err := cr.reloadVariables(); err != nil {
		log.Printf("[ConfigReloader] ❌ 重载测点配置失败: %v", err)
		return
	}

	// 4. 重载网关配置
	if err := cr.reloadGateways(); err != nil {
		log.Printf("[ConfigReloader] ❌ 重载网关配置失败: %v", err)
		return
	}

	// 5. 重载任务配置
	if err := cr.reloadTasks(); err != nil {
		log.Printf("[ConfigReloader] ❌ 重载任务配置失败: %v", err)
		return
	}

	// 6. 更新本地版本号
	cr.lastVersion = newVersion
	log.Printf("[ConfigReloader] ✅ 配置热重载成功! 当前版本: %s", newVersion)
}

// getCurrentVersion 从数据库获取当前配置版本
func (cr *ConfigReloader) getCurrentVersion() (string, error) {
	version, err := database.GetConfigVersion()
	if err != nil {
		return "", err
	}
	return version, nil
}

// reloadVariables 重载测点配置
func (cr *ConfigReloader) reloadVariables() error {
	log.Println("[ConfigReloader] 正在重载测点配置...")

	// 从数据库加载最新配置
	tags, err := database.LoadVariables()
	if err != nil {
		return err
	}

	// 更新到内存
	tagManager := GetTagManager()
	tagManager.LoadTags(tags)

	log.Printf("[ConfigReloader] ✅ 测点配置已更新 (%d个测点)", len(tags))
	return nil
}

// GatewayReloader 网关重载回调函数类型
type GatewayReloader func() error

var (
	gatewayReloadFunc GatewayReloader
	taskReloadFunc    func() error // 任务重载函数
)

// SetGatewayReloader 设置网关重载回调函数 (从main包注入)
func SetGatewayReloader(fn GatewayReloader) {
	gatewayReloadFunc = fn
}

// SetTaskReloader 设置任务重载函数
func SetTaskReloader(fn func() error) {
	taskReloadFunc = fn
}

// reloadGateways 重载网关配置
func (cr *ConfigReloader) reloadGateways() error {
	if gatewayReloadFunc == nil {
		log.Println("[ConfigReloader] ⚠️ 网关重载函数未设置,跳过网关重载")
		return nil
	}

	return gatewayReloadFunc()
}

// reloadTasks 重载任务配置
func (cr *ConfigReloader) reloadTasks() error {
	if taskReloadFunc == nil {
		log.Println("[ConfigReloader] ⚠️ 任务重载函数未设置,跳过任务重载")
		return nil
	}

	return taskReloadFunc()
}

// ForceReload 手动触发配置重载 (HTTP API调用)
func (cr *ConfigReloader) ForceReload() error {
	log.Println("[ConfigReloader] 🔧 手动触发配置重载...")

	// 强制重载,不检查版本
	if err := cr.reloadVariables(); err != nil {
		return err
	}

	if err := cr.reloadGateways(); err != nil {
		return err
	}

	if err := cr.reloadTasks(); err != nil {
		return err
	}

	// 更新版本号
	version, _ := cr.getCurrentVersion()
	cr.lastVersion = version

	log.Println("[ConfigReloader] ✅ 手动重载完成")
	return nil
}
