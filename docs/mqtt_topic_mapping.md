# MQTT 主题映射规划

## SCADA 与 IIoT网关 主题对接

### 1. 数据上报 (SCADA → 网关)
```
主题格式: sensor/{设备类型}/{设备ID}/data
示例:
- sensor/plc/device001/data     # PLC设备数据
- sensor/temperature/temp01/data # 温度传感器
- sensor/pressure/press01/data   # 压力传感器
```

### 2. 控制指令 (网关 → SCADA)  
```
主题格式: sensor/{设备类型}/{设备ID}/cmd
示例:
- sensor/plc/device001/cmd      # PLC控制指令
- sensor/valve/valve01/cmd      # 阀门控制
```

### 3. 系统状态
```
- system/scada/status          # SCADA系统状态
- system/gateway/status        # 网关系统状态
- system/heartbeat            # 心跳消息
```

## SCADA 配置建议

### 订阅主题配置:
```
sensor/+/cmd        # 接收网关的控制指令
sensor/test/set     # 接收网关的测试消息
system/+/status     # 接收系统状态消息
```

### 发布主题配置:
```
sensor/plc/+/data   # 发布PLC数据
sensor/+/+/data     # 发布所有传感器数据
system/scada/status # 发布SCADA状态
```

## 数据格式标准

### 传感器数据格式:
```json
{
  "device_id": "device001",
  "device_type": "plc", 
  "timestamp": "2025-12-11T15:30:00Z",
  "values": {
    "temperature": 25.6,
    "pressure": 1.2,
    "status": "running"
  },
  "quality": "good",
  "source": "scada"
}
```

### 控制指令格式:
```json
{
  "cmd_id": "cmd_001",
  "device_id": "device001", 
  "action": "set_value",
  "parameters": {
    "target_temp": 30.0,
    "mode": "auto"
  },
  "timestamp": "2025-12-11T15:30:00Z",
  "source": "gateway"
}
```




























