# MQTT 稳定性测试指南

## 验收标准测试

### 1. 基础功能测试

启动程序：
```bash
go run main.go
```

预期输出：
```
🚀 启动 IIoT 网关系统...
📋 配置加载完成: MQTT=tcp://127.0.0.1:1883, DB=localhost:5432
✅ 数据库连接成功 (或 ⚠️ 数据库连接失败，继续运行)
[MQTT] 正在连接到 Broker: tcp://127.0.0.1:1883
[MQTT] 客户端连接成功，等待订阅完成...
[MQTT] ✅ 连接成功！开始重新订阅主题...
[MQTT] 成功订阅主题: sensor/+/data
[MQTT] 🔄 所有订阅已恢复，系统就绪
[MQTT] 🚀 测试发布协程已启动 (每秒发送到 sensor/test/set)
🌐 HTTP服务器启动: http://localhost:8080
📡 MQTT监听: sensor/+/data
📤 测试发布: sensor/test/set (每秒)
🔧 验收测试: 请拔网线或关闭MQTT Broker测试自动重连
```

### 2. 并发读写测试

程序运行后会自动：
- **读取**: 监听 `sensor/+/data` 主题
- **写入**: 每秒向 `sensor/test/set` 发送测试消息

### 3. 断线重连测试

#### 测试步骤：
1. 程序正常运行时，关闭MQTT Broker或拔网线
2. 观察日志输出断线信息
3. 恢复网络/重启Broker
4. 验证自动重连和订阅恢复

#### 预期行为：
- 断线时：`[MQTT] ❌ 连接丢失: ... - 自动重连中...`
- 重连时：`[MQTT] ✅ 连接成功！开始重新订阅主题...`
- 程序不会崩溃，继续运行

### 4. API 测试

访问以下端点验证系统状态：

```bash
# 基础状态
curl http://localhost:8080/

# 详细状态
curl http://localhost:8080/status

# MQTT统计
curl http://localhost:8080/api/v1/mqtt/stats

# 手动发布消息
curl -X POST http://localhost:8080/api/v1/mqtt/publish \
  -H "Content-Type: application/json" \
  -d '{"topic": "sensor/manual/data", "message": "手动测试消息"}'
```

### 5. 性能监控

程序每30秒输出系统状态：
```
📊 系统状态 - MQTT连接: true, 消息计数: 156, 运行时长: 2m30s
```

## 配置说明

根据你的MQTT配置，系统使用以下参数：
- **Broker**: tcp://127.0.0.1:1883
- **用户名**: root
- **密码**: (空)
- **心跳**: 10秒
- **自动重连**: 启用
- **清理会话**: 禁用 (确保重连后补收消息)

## 故障排除

1. **连接失败**: 检查MQTT Broker是否运行在127.0.0.1:1883
2. **认证失败**: 确认用户名密码配置正确
3. **订阅失败**: 检查网络连接和Broker权限设置

## 下一步开发

当前版本完成了MQTT核心稳定性构建，后续可以：
1. 实现7-Worker模型的其他协程
2. 添加内存Tag管理 (map[string]*Tag)
3. 实现logicChan和各种Worker
4. 集成TimescaleDB时序数据存储
