package gateway

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"gin-mqtt-pgsql/core"
	"gin-mqtt-pgsql/database"
	"gin-mqtt-pgsql/models"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// GatewayManager 多网关管理器
type GatewayManager struct {
	mu       sync.RWMutex
	gateways map[int]*Gateway // key: gateway_id
}

var globalGatewayManager *GatewayManager

// InitGatewayManager 初始化全局网关管理器
func InitGatewayManager() *GatewayManager {
	globalGatewayManager = &GatewayManager{
		gateways: make(map[int]*Gateway),
	}
	return globalGatewayManager
}

// GetGatewayManager 获取全局网关管理器
func GetGatewayManager() *GatewayManager {
	return globalGatewayManager
}

// Gateway 单个网关实例
type Gateway struct {
	ID       int
	Name     string
	Config   database.GatewayConfig
	Client   MQTT.Client
	IsActive bool
	mu       sync.RWMutex
}

// StartGateway 启动单个网关 (独立goroutine)
func (gm *GatewayManager) StartGateway(config database.GatewayConfig) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	// 检查是否已存在
	if _, exists := gm.gateways[config.ID]; exists {
		return fmt.Errorf("网关 %d 已存在", config.ID)
	}

	// 创建网关实例
	gateway := &Gateway{
		ID:       config.ID,
		Name:     config.Name,
		Config:   config,
		IsActive: false,
	}

	// 启动独立协程运行网关
	go gateway.run()

	gm.gateways[config.ID] = gateway
	log.Printf("[GatewayManager] 网关 %d (%s) 已启动", config.ID, config.Name)

	return nil
}

// StopGateway 停止指定网关
func (gm *GatewayManager) StopGateway(gatewayID int) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gateway, exists := gm.gateways[gatewayID]
	if !exists {
		return fmt.Errorf("网关 %d 不存在", gatewayID)
	}

	gateway.stop()
	delete(gm.gateways, gatewayID)

	log.Printf("[GatewayManager] 网关 %d 已停止", gatewayID)
	return nil
}

// LoadAndStartAll 从数据库加载并启动所有网关
func (gm *GatewayManager) LoadAndStartAll() error {
	configs, err := database.LoadGateways()
	if err != nil {
		return fmt.Errorf("加载网关配置失败: %w", err)
	}

	log.Printf("[GatewayManager] 加载到 %d 个网关配置", len(configs))

	for _, config := range configs {
		if err := gm.StartGateway(config); err != nil {
			log.Printf("[GatewayManager] 启动网关 %d 失败: %v", config.ID, err)
			continue
		}
	}

	return nil
}

// GetStatus 获取所有网关状态
func (gm *GatewayManager) GetStatus() map[int]bool {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	status := make(map[int]bool)
	for id, gw := range gm.gateways {
		gw.mu.RLock()
		status[id] = gw.IsActive
		gw.mu.RUnlock()
	}

	return status
}

// GetGateway 获取指定ID的网关
func (gm *GatewayManager) GetGateway(gatewayID int) (*Gateway, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	gw, exists := gm.gateways[gatewayID]
	if !exists {
		return nil, fmt.Errorf("网关 %d 不存在", gatewayID)
	}

	gw.mu.RLock()
	isActive := gw.IsActive
	gw.mu.RUnlock()

	if !isActive {
		return nil, fmt.Errorf("网关 %d 未激活", gatewayID)
	}

	return gw, nil
}

// GetFirstActiveGateway 获取第一个活跃的网关
func (gm *GatewayManager) GetFirstActiveGateway() (*Gateway, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	for _, gw := range gm.gateways {
		gw.mu.RLock()
		isActive := gw.IsActive
		gw.mu.RUnlock()

		if isActive {
			return gw, nil
		}
	}

	return nil, fmt.Errorf("没有活跃的网关")
}

// Publish 通过网关发布MQTT消息
func (gw *Gateway) Publish(topic string, payload interface{}, qos byte, retain bool) error {
	gw.mu.RLock()
	client := gw.Client
	isActive := gw.IsActive
	gw.mu.RUnlock()

	if !isActive || client == nil || !client.IsConnected() {
		return fmt.Errorf("网关未连接")
	}

	token := client.Publish(topic, qos, retain, payload)

	// 异步等待，避免阻塞
	go func() {
		if token.Wait() && token.Error() != nil {
			log.Printf("[Gateway-%d] MQTT发布失败: %v", gw.ID, token.Error())
		}
	}()

	return nil
}

// run 网关运行主循环 (独立goroutine)
func (gw *Gateway) run() {
	log.Printf("[Gateway-%d] 协程启动: %s", gw.ID, gw.Name)

	// 创建MQTT客户端配置
	opts := MQTT.NewClientOptions()
	opts.AddBroker(gw.Config.Broker)

	// 🔧 为ClientID添加随机后缀，避免多实例或重启时的ID冲突
	// 格式: 原始ID_时间戳纳秒_随机数 (例: Client_Go_001_1702345678123456789_8234)
	uniqueClientID := fmt.Sprintf("%s_%d_%d",
		gw.Config.ClientID,
		time.Now().UnixNano(),
		rand.Intn(10000))
	opts.SetClientID(uniqueClientID)

	log.Printf("[Gateway-%d] 🔧 MQTT ClientID: %s", gw.ID, uniqueClientID)

	if gw.Config.Username != "" {
		opts.SetUsername(gw.Config.Username)
	}
	if gw.Config.Password != "" {
		opts.SetPassword(gw.Config.Password)
	}

	// 连接配置
	// 🔧 架构说明：SCADA和本网关都作为客户端连接到同一个MQTT Broker
	// - SCADA设置：心跳30秒（SCADA→Broker）
	// - 网关设置：心跳60秒是安全的默认值（网关→Broker）
	// - 如果Broker有特殊要求，需要查看Broker日志确定超时原因
	opts.SetKeepAlive(60 * time.Second)            // 标准的60秒心跳（大多数Broker默认值）
	opts.SetCleanSession(false)                    // 断线重连后补收消息
	opts.SetAutoReconnect(true)                    // 启用自动重连
	opts.SetMaxReconnectInterval(30 * time.Second) // 最大重连间隔30秒
	opts.SetPingTimeout(20 * time.Second)          // 心跳超时20秒（宽松一些）
	opts.SetConnectTimeout(30 * time.Second)       // 连接超时30秒
	opts.SetWriteTimeout(10 * time.Second)         // 写入超时10秒

	// 设置回调
	opts.OnConnect = gw.onConnect
	opts.OnConnectionLost = gw.onConnectionLost

	// 🔧 禁用TCP层的KeepAlive，只使用MQTT层的KeepAlive
	// 避免TCP和MQTT的KeepAlive冲突
	opts.SetTLSConfig(nil) // 确保使用普通TCP
	opts.SetKeepAlive(60 * time.Second)

	// 创建客户端
	gw.Client = MQTT.NewClient(opts)

	log.Printf("[Gateway-%d] 🔧 MQTT配置: KeepAlive=60s, PingTimeout=20s, AutoReconnect=true", gw.ID)

	// 连接
	log.Printf("[Gateway-%d] 正在连接: %s", gw.ID, gw.Config.Broker)
	if token := gw.Client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("[Gateway-%d] 连接失败: %v", gw.ID, token.Error())
		return
	}

	// 保持运行 (MQTT库会自动维护连接)
	for {
		time.Sleep(10 * time.Second)

		// 检查是否需要停止
		gw.mu.RLock()
		if !gw.IsActive && gw.Client != nil && !gw.Client.IsConnected() {
			gw.mu.RUnlock()
			break
		}
		gw.mu.RUnlock()
	}

	log.Printf("[Gateway-%d] 协程退出", gw.ID)
}

// stop 停止网关
func (gw *Gateway) stop() {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	if gw.Client != nil && gw.Client.IsConnected() {
		gw.Client.Disconnect(250)
	}

	gw.IsActive = false
	log.Printf("[Gateway-%d] 已断开连接", gw.ID)
}

// onConnect 连接成功回调
func (gw *Gateway) onConnect(client MQTT.Client) {
	gw.mu.Lock()
	gw.IsActive = true
	gw.mu.Unlock()

	log.Printf("[Gateway-%d] ✅ 连接成功，订阅主题: %s", gw.ID, gw.Config.Topic)

	// 订阅主题
	token := client.Subscribe(gw.Config.Topic, 1, gw.messageHandler)
	token.Wait()
	if token.Error() != nil {
		log.Printf("[Gateway-%d] 订阅失败: %v", gw.ID, token.Error())
		return
	}

	log.Printf("[Gateway-%d] ✅ 订阅成功", gw.ID)
}

// onConnectionLost 连接丢失回调
func (gw *Gateway) onConnectionLost(client MQTT.Client, err error) {
	gw.mu.Lock()
	gw.IsActive = false
	gw.mu.Unlock()

	log.Printf("[Gateway-%d] ❌ 连接丢失: %v - 自动重连中...", gw.ID, err)
}

// messageHandler MQTT消息处理器 - 非阻塞设计
func (gw *Gateway) messageHandler(client MQTT.Client, msg MQTT.Message) {
	// 立即返回，避免阻塞MQTT内部协程
	go func() {
		// 输出原始MQTT消息
		//log.Printf("[Gateway-%d] 📨 收到MQTT消息: Topic=%s, Payload=%s", gw.ID, msg.Topic(), string(msg.Payload()))

		mqttMsg := &models.MQTTMessage{
			GatewayID: gw.ID,
			Topic:     msg.Topic(),
			Payload:   msg.Payload(),
			Timestamp: time.Now(),
		}

		// 非阻塞发送到logicChan
		select {
		case core.LogicChan <- mqttMsg:
			// 发送成功
		default:
			log.Printf("[Gateway-%d] LogicChan已满，丢弃消息: %s", gw.ID, msg.Topic())
		}
	}()
}
