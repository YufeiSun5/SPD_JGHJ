-- ===================================
-- IIoT 网关系统 - 示例数据
-- ===================================

-- 1. 插入网关配置
INSERT INTO sys_gateways (gw_name, status, mqtt_broker, mqtt_client_id, mqtt_user, mqtt_pass, mqtt_topic)
VALUES 
('1号车间采集器', 1, 'tcp://127.0.0.1:1883', 'gateway_workshop_01', 'admin', 'admin123', 'factory/workshop1/#'),
('2号车间采集器', 1, 'tcp://127.0.0.1:1883', 'gateway_workshop_02', 'admin', 'admin123', 'factory/workshop2/#');

-- 2. 插入逻辑设备
INSERT INTO sys_devices (gateway_id, device_code, device_name, identify_key)
VALUES 
(1, 'PLC_01', '1号PLC控制器', 'plc01'),
(1, 'SENSOR_01', '温度传感器组', 'sensors'),
(2, 'PLC_02', '2号PLC控制器', 'plc02');

-- 3. 插入测点变量配置
-- 3.1 温度测点 (变动存储 + 报警)
INSERT INTO sys_variables 
(device_id, var_name, display_name, unit, json_path, scale_factor, offset_val, 
 alarm_enable, limit_hh, limit_h, limit_l, limit_ll, deadband, alarm_msg,
 store_mode, store_cycle, store_deadband, source_type)
VALUES 
(2, 'temp_inlet', '进水温度', '℃', 'temp_inlet', 1.0, 0.0,
 1, 85.0, 75.0, 5.0, 0.0, 2.0, '进水温度异常',
 1, 60, 0.5, 0),

(2, 'temp_outlet', '出水温度', '℃', 'temp_outlet', 1.0, 0.0,
 1, 90.0, 80.0, 10.0, 5.0, 2.0, '出水温度异常',
 1, 60, 0.5, 0);

-- 3.2 压力测点 (定时存储 + 报警)
INSERT INTO sys_variables 
(device_id, var_name, display_name, unit, json_path, scale_factor, offset_val, 
 alarm_enable, limit_hh, limit_h, limit_l, limit_ll, deadband, alarm_msg,
 store_mode, store_cycle, store_deadband, source_type)
VALUES 
(1, 'pressure_main', '主管道压力', 'MPa', 'pressure', 1.0, 0.0,
 1, 10.0, 8.0, 2.0, 1.0, 0.5, '主管道压力异常',
 2, 30, 0.0, 0);

-- 3.3 流量测点 (混合存储: 变动+定时)
INSERT INTO sys_variables 
(device_id, var_name, display_name, unit, json_path, scale_factor, offset_val, 
 alarm_enable, limit_hh, limit_h, limit_l, limit_ll, deadband, alarm_msg,
 store_mode, store_cycle, store_deadband, source_type)
VALUES 
(1, 'flow_rate', '流量', 'm³/h', 'flow', 1.0, 0.0,
 0, 0.0, 0.0, 0.0, 0.0, 0.0, NULL,
 3, 60, 1.0, 0);

-- 3.4 开关状态测点 (变动存储, 无报警)
INSERT INTO sys_variables 
(device_id, var_name, display_name, unit, json_path, scale_factor, offset_val, 
 alarm_enable, store_mode, store_cycle, store_deadband, source_type)
VALUES 
(1, 'pump_status', '水泵运行状态', '', 'pump_on', 1.0, 0.0,
 0, 1, 60, 0.1, 0),

(1, 'valve_status', '阀门开度', '%', 'valve_open', 1.0, 0.0,
 0, 1, 60, 1.0, 0);

-- 4. 插入配置版本记录
INSERT INTO sys_config_version (id, version_code) 
VALUES (1, 'v1.0.0-initial');

-- ===================================
-- 测试数据模拟
-- ===================================

-- 模拟MQTT消息格式:
/*
Topic: factory/workshop1/data

Payload:
{
  "plc01": {
    "pressure": 6.5,
    "flow": 125.3,
    "pump_on": 1,
    "valve_open": 75.5
  },
  "sensors": {
    "temp_inlet": 45.2,
    "temp_outlet": 62.8
  }
}
*/

-- 查询已配置的测点
SELECT 
    v.var_name,
    v.display_name,
    v.json_path,
    d.device_name,
    g.gw_name,
    CASE v.store_mode
        WHEN 0 THEN '不存储'
        WHEN 1 THEN '变动存储'
        WHEN 2 THEN '定时存储'
        WHEN 3 THEN '混合存储'
    END as store_mode_desc,
    v.alarm_enable
FROM sys_variables v
INNER JOIN sys_devices d ON v.device_id = d.id
INNER JOIN sys_gateways g ON d.gateway_id = g.id
WHERE g.status = 1
ORDER BY g.id, d.id, v.id;



























