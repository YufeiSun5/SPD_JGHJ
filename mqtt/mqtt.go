package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"gin-mqtt-pgsql/config"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var (
	Client        MQTT.Client
	messageCount  int64                       // 消息计数器
	isConnected   int32                       // 连接状态 (0=断开, 1=已连接)
	subscriptions = []string{"sensor/+/data"} // 需要订阅的主题列表
)

// InitMQTT 初始化MQTT客户端 - 按照IIoT网关架构要求
func InitMQTT(cfg *config.MQTTConfig) error {
	opts := MQTT.NewClientOptions()

	// 构建完整的broker地址
	var broker string
	if cfg.Broker == "tcp://127.0.0.1" {
		broker = fmt.Sprintf("%s:%d", cfg.Broker, cfg.Port)
	} else {
		broker = fmt.Sprintf("tcp://%s:%d", cfg.Broker, cfg.Port)
	}

	opts.AddBroker(broker)
	opts.SetClientID(cfg.ClientID)

	// 认证配置
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	// 关键稳定性配置 - 符合工业级要求
	opts.SetKeepAlive(time.Duration(cfg.KeepAlive) * time.Second)
	opts.SetCleanSession(cfg.CleanSession) // false确保重连后补收消息
	opts.SetAutoReconnect(cfg.AutoReconnect)
	opts.SetMaxReconnectInterval(time.Duration(cfg.MaxReconnectInterval) * time.Second)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetConnectTimeout(10 * time.Second)

	// 设置关键回调处理器
	opts.SetDefaultPublishHandler(defaultMessageHandler)
	opts.OnConnect = onConnectHandler
	opts.OnConnectionLost = onConnectionLostHandler

	Client = MQTT.NewClient(opts)

	log.Printf("[MQTT] 正在连接到 Broker: %s", broker)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	log.Println("[MQTT] 客户端连接成功，等待订阅完成...")
	return nil
}

// GetClient 获取MQTT客户端实例
func GetClient() MQTT.Client {
	return Client
}

// IsConnected 检查连接状态
func IsConnected() bool {
	return atomic.LoadInt32(&isConnected) == 1
}

// GetMessageCount 获取接收消息计数
func GetMessageCount() int64 {
	return atomic.LoadInt64(&messageCount)
}

// Publish 发布消息到指定主题 - 非阻塞设计
func Publish(topic string, payload interface{}) error {
	if Client == nil || !IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	// 使用QoS 0确保非阻塞
	token := Client.Publish(topic, 0, false, payload)

	// 异步等待，避免阻塞主流程
	go func() {
		if token.Wait() && token.Error() != nil {
			log.Printf("[MQTT] 发布消息失败: %v", token.Error())
		}
	}()

	return nil
}

// Subscribe 订阅主题
func Subscribe(topic string, callback MQTT.MessageHandler) error {
	if Client == nil {
		return fmt.Errorf("MQTT client not initialized")
	}

	token := Client.Subscribe(topic, 1, callback)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("订阅主题 %s 失败: %w", topic, token.Error())
	}

	log.Printf("[MQTT] 成功订阅主题: %s", topic)
	return nil
}

// 重新订阅所有主题 - 重连后调用
func resubscribeAll() {
	for _, topic := range subscriptions {
		if err := Subscribe(topic, sensorDataHandler); err != nil {
			log.Printf("[MQTT] 重新订阅失败 %s: %v", topic, err)
		}
	}
}

// 默认消息处理器 - 非阻塞设计
var defaultMessageHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	// 快速处理，避免阻塞MQTT内部协程
	go func() {
		atomic.AddInt64(&messageCount, 1)
		log.Printf("[MQTT] 收到消息 [%d]: Topic=%s, Payload=%s",
			atomic.LoadInt64(&messageCount), msg.Topic(), string(msg.Payload()))
	}()
}

// 传感器数据处理器 - 符合架构要求的高速通道
var sensorDataHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	// 非阻塞处理，立即放入logicChan（后续实现）
	go func() {
		count := atomic.AddInt64(&messageCount, 1)
		log.Printf("[传感器数据] [%d] Topic: %s, Data: %s",
			count, msg.Topic(), string(msg.Payload()))

		// TODO: 后续将数据放入 logicChan 供 LogicWorker 处理
		// logicChan <- &SensorData{Topic: msg.Topic(), Payload: msg.Payload(), Timestamp: time.Now()}
	}()
}

// 连接成功处理器 - 关键：重连后必须重新订阅
var onConnectHandler MQTT.OnConnectHandler = func(client MQTT.Client) {
	atomic.StoreInt32(&isConnected, 1)
	log.Println("[MQTT] ✅ 连接成功！开始重新订阅主题...")

	// 重连后立即重新订阅，防止订阅丢失
	resubscribeAll()

	log.Println("[MQTT] 🔄 所有订阅已恢复，系统就绪")
}

// 连接丢失处理器 - 记录断线情况
var onConnectionLostHandler MQTT.ConnectionLostHandler = func(client MQTT.Client, err error) {
	atomic.StoreInt32(&isConnected, 0)
	log.Printf("[MQTT] ❌ 连接丢失: %v - 自动重连中...", err)
}

// StartTestPublisher 启动测试发布协程 - 验证并发写能力
func StartTestPublisher() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		counter := 0
		for range ticker.C {
			if !IsConnected() {
				continue // 未连接时跳过发送
			}

			counter++
			testMsg := fmt.Sprintf(`{"cmd":"test","counter":%d,"timestamp":"%s"}`,
				counter, time.Now().Format("2006-01-02 15:04:05"))

			if err := Publish("sensor/test/set", testMsg); err != nil {
				log.Printf("[测试发布] 发送失败: %v", err)
			} else {
				log.Printf("[测试发布] 发送指令 #%d", counter)
			}
		}
	}()

	log.Println("[MQTT] 🚀 测试发布协程已启动 (每秒发送到 sensor/test/set)")
}

// StartKingIOTTestWriter 启动KingIOT写入测试协程 - 每5秒写入随机数据
func StartKingIOTTestWriter() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		qidCounter := int64(197296731)
		for range ticker.C {
			if !IsConnected() {
				continue // 未连接时跳过发送
			}

			qidCounter++
			// 生成各种类型的随机数据
			randomInt := rand.Intn(100)                             // int类型
			randomFloat := 20.0 + rand.Float64()*15.0               // float类型
			randomStatus := rand.Intn(2)                            // int类型
			randomString := fmt.Sprintf("test_data_%d", qidCounter) // string类型

			// 构建KingIOT写入数据格式
			writeData := map[string]interface{}{
				"Writer":    "test_user",
				"WriteTime": time.Now().Format("2006-01-02 15:04:05.000 -0700"),
				"Username":  "sa",
				"Password":  "C12E01F2A13FF5587E1E9E4AEDB8242D",
				"Qid":       qidCounter,
				"PNs": map[string]string{
					"1": "V", // int类型参数
					"2": "T", // 时间类型参数
					"3": "Q", // int类型参数
					"4": "F", // float类型参数
					"5": "S", // string类型参数
				},
				"PVs": map[string]interface{}{
					"1": randomInt,                                          // 写入int值
					"2": time.Now().Format("2006-01-02 15:04:05.000 -0700"), // 写入时间
					"3": randomStatus,                                       // 写入状态int
					"4": randomFloat,                                        // 写入float值
					"5": randomString,                                       // 写入string值
				},
				"Objs": []map[string]interface{}{
					{
						"N": "可写变量1",
						"1": 20 + rand.Intn(50), // int
					},
					{
						"N": "可写变量2",
						"1": 10.5 + rand.Float64()*20.0, // float
					},
					{
						"N": "可写变量3",
						"1": fmt.Sprintf("hello:%d", rand.Intn(100)), // string
					},
				},
			}

			// 序列化为JSON
			payload, err := json.Marshal(writeData)
			if err != nil {
				log.Printf("[KingIOT写入测试] JSON序列化失败: %v", err)
				continue
			}

			// 发布到主题 setdata_S_KIO_Project
			topic := "setdata_S_KIO_Project"
			if err := Publish(topic, string(payload)); err != nil {
				log.Printf("[KingIOT写入测试] 发送失败: %v", err)
			} else {
				log.Printf("[KingIOT写入测试] ✅ 写入成功 Qid=%d, Int=%d, Float=%.2f, Str=%s",
					qidCounter, randomInt, randomFloat, randomString)
			}
		}
	}()

	log.Println("[MQTT] 🚀 KingIOT写入测试协程已启动 (每5秒写入到 setdata_S_KIO_Project)")
}
