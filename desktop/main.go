package main

import (
	"embed"
	"log"
	"net"
	"os"
	"time"

	"gin-mqtt-pgsql/config"
	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/gateway"
	"gin-mqtt-pgsql/models"
	"gin-mqtt-pgsql/workers"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:dist
var assets embed.FS

func main() {
	log.Println("🚀 启动 IIoT 桌面监控客户端...")

	// ── 单实例锁（Single-instance lock / シングルインスタンスロック） ──────────────
	// CN: 尝试监听本机固定端口；若端口已被占用，说明另一实例正在运行，直接退出。
	//     使用 TCP 端口（而非文件锁）可避免进程崩溃后锁文件残留的问题。
	// EN: Bind to a fixed local port; if already in use, another instance is running → exit.
	//     TCP port avoids stale lock-file issues after a crash.
	// JP: 固定ローカルポートをバインド。既に使用中なら別インスタンスが起動中 → 終了。
	//     TCP ポートはクラッシュ後のロックファイル残留問題を回避できる。
	const singleInstanceAddr = "127.0.0.1:47391"
	sil, err := net.Listen("tcp", singleInstanceAddr)
	if err != nil {
		log.Printf("⚠️  检测到程序已有一个实例在运行（%s 已被占用），禁止重复启动，本进程退出。", singleInstanceAddr)
		os.Exit(1)
	}
	defer sil.Close()

	// 初始化后端系统
	initBackend()

	// 创建Wails应用
	app := NewApp()

	log.Println("📦 准备启动 Wails 窗口...")

	err = wails.Run(&options.App{
		Title:     "IIoT 网关监控",
		Width:     1440,
		Height:    800,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.Startup,
		Bind: []interface{}{
			app,
		},
		WindowStartState: options.Normal,
		CSSDragProperty:  "--wails-draggable",
		CSSDragValue:     "drag",
	})

	if err != nil {
		log.Printf("❌ Wails启动失败: %v", err)
		log.Println("💡 提示: 请确保已安装 WebView2 Runtime")
		log.Println("💡 下载地址: https://developer.microsoft.com/en-us/microsoft-edge/webview2/")
	} else {
		log.Println("✅ Wails 窗口已关闭")
	}
}

// initBackend 初始化后端系统 (精简版)
func initBackend() {
	// 先尝试加载当前目录的配置文件，然后尝试上层目录
	config.LoadEnvFileFromPath("config.env")
	config.LoadEnvFileFromPath("../config.env")

	cfg := config.LoadConfig()

	log.Printf("📋 数据库配置: Host=%s Port=%d DBName=%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)

	// 🔧 修改: 数据库连接失败时不退出，允许界面启动
	if err := database.InitDatabase(&cfg.Database); err != nil {
		log.Printf("⚠️ 数据库连接失败: %v", err)
		log.Println("💡 应用将以离线模式启动，部分功能不可用")
		// 初始化基础组件，但不加载数据
		core.InitChannels()
		core.InitTagManager()
		core.GetTagManager().LoadTags([]*models.Tag{}) // 空测点列表
		log.Println("✅ 离线模式已就绪（无数据）")
		return
	}

	// 自动建表 / 补全缺失字段（幂等，已存在的表不会被破坏）
	// Auto-migrate all MES tables on startup; idempotent, existing tables are not dropped.
	// 起動時に全MESTテーブルを自動マイグレーション。冪等処理で既存テーブルは保持。
	if err := database.InitMESDatabase(); err != nil {
		log.Printf("⚠️ 数据库表结构初始化失败: %v", err)
	}

	core.InitChannels()
	core.InitTagManager()

	tags, err := database.LoadVariables()
	if err != nil {
		log.Printf("⚠️ 加载测点失败: %v", err)
		tags = []*models.Tag{} // 空列表
	}
	core.GetTagManager().LoadTags(tags)
	log.Printf("✅ 加载 %d 个测点", len(tags))

	workers.StartLogicWorkers(20)
	workers.StartChangeWorkers(5)
	workers.StartCycleWorkers(5)
	workers.StartScanner()
	workers.StartAlarmWorkers(3)

	// 启动任务调度系统
	workers.InitTaskScheduler()
	workers.StartTaskScheduler()
	workers.StartEventProcessors(3)

	// 加载任务配置
	tasks, err := database.LoadTasks()
	if err != nil {
		log.Printf("⚠️ 加载任务配置失败: %v", err)
	} else {
		workers.GetTaskScheduler().LoadTasks(tasks)
		log.Printf("✅ 加载 %d 个任务配置", len(tasks))
	}

	gateway.InitGatewayManager()
	if err := gateway.GetGatewayManager().LoadAndStartAll(); err != nil {
		log.Printf("⚠️ 加载网关失败: %v", err)
	}

	// 启动配置热重载监控 (每10秒检查一次版本变化)
	reloader := core.InitConfigReloader(10 * time.Second)

	// 注入网关重载函数
	core.SetGatewayReloader(func() error {
		log.Println("[ConfigReloader] 正在重载网关配置...")
		gwManager := gateway.GetGatewayManager()
		currentStatus := gwManager.GetStatus()
		newConfigs, err := database.LoadGateways()
		if err != nil {
			return err
		}

		newConfigMap := make(map[int]database.GatewayConfig)
		for _, config := range newConfigs {
			newConfigMap[config.ID] = config
		}

		for _, newConfig := range newConfigs {
			if _, exists := currentStatus[newConfig.ID]; !exists {
				log.Printf("[ConfigReloader] 🆕 新增网关: ID=%d, Name=%s", newConfig.ID, newConfig.Name)
				if err := gwManager.StartGateway(newConfig); err != nil {
					log.Printf("[ConfigReloader] 启动新网关失败: %v", err)
				}
			}
		}

		for gwID := range currentStatus {
			if _, exists := newConfigMap[gwID]; !exists {
				log.Printf("[ConfigReloader] 🗑️  删除网关: ID=%d", gwID)
				gwManager.StopGateway(gwID)
			}
		}

		log.Printf("[ConfigReloader] ✅ 网关配置已更新")
		return nil
	})

	// 注入任务重载函数
	core.SetTaskReloader(func() error {
		log.Println("[ConfigReloader] 正在重载任务配置...")
		tasks, err := database.LoadTasks()
		if err != nil {
			return err
		}
		workers.GetTaskScheduler().LoadTasks(tasks)
		log.Printf("[ConfigReloader] ✅ 任务配置已更新 (%d个任务)", len(tasks))
		return nil
	})

	// ⭐ 启动热重载监控
	reloader.Start()

	log.Println("✅ 后端系统已就绪 (含热重载)")
}
