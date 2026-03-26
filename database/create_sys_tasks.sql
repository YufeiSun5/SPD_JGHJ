-- =====================================================
-- 任务调度系统表结构 (MySQL)
-- =====================================================

-- 任务表
CREATE TABLE IF NOT EXISTS `sys_tasks` (
  `task_id` bigint NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `task_name` varchar(100) NOT NULL COMMENT '任务名称',
  `task_type` tinyint NOT NULL COMMENT '任务类型: 1=定时, 2=数据改变, 3=条件事件',
  `is_enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用: 1=启用, 0=禁用',
  `description` varchar(255) DEFAULT NULL COMMENT '任务描述',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  -- 定时任务配置 (task_type=1 时使用)
  `cron_expr` varchar(50) DEFAULT NULL COMMENT 'Cron 表达式 (如: */5 * * * *)',
  `interval_sec` int DEFAULT NULL COMMENT '简单间隔(秒), 优先使用 cron_expr',
  `last_run_time` datetime DEFAULT NULL COMMENT '上次执行时间',
  
  -- 数据改变任务配置 (task_type=2 时使用)
  `trigger_var_id` bigint DEFAULT NULL COMMENT '触发的测点ID (关联 sys_variables.id)',
  `trigger_var_name` varchar(100) DEFAULT NULL COMMENT '触发的测点名称',
  `change_type` varchar(20) DEFAULT 'ANY' COMMENT '变化类型: ANY, INCREASE, DECREASE, THRESHOLD, FALSE_TO_TRUE, TRUE_TO_FALSE',
  `change_threshold` float DEFAULT NULL COMMENT '变化阈值 (用于 THRESHOLD 类型)',
  
  -- 条件事件任务配置 (task_type=3 时使用)
  `condition_expr` varchar(500) DEFAULT NULL COMMENT '条件表达式 (如: temp>50 AND pressure>100)',
  
  -- 任务动作配置 (所有类型共用)
  `action_type` tinyint NOT NULL COMMENT '动作类型: 1=HTTP请求, 2=MQTT发布, 3=数据库操作, 4=执行脚本, 5=写日志',
  `action_config` text NOT NULL COMMENT '动作配置 (JSON格式)',
  
  PRIMARY KEY (`task_id`),
  KEY `idx_type_enabled` (`task_type`, `is_enabled`),
  KEY `idx_trigger_var` (`trigger_var_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务调度配置表';

-- 任务执行日志表
CREATE TABLE IF NOT EXISTS `sys_task_execution_logs` (
  `log_id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `task_id` bigint NOT NULL COMMENT '任务ID',
  `task_name` varchar(100) NOT NULL COMMENT '任务名称',
  `execute_time` datetime NOT NULL COMMENT '执行时间',
  `success` tinyint NOT NULL COMMENT '是否成功: 1=成功, 0=失败',
  `error_msg` text COMMENT '错误信息',
  `duration` int DEFAULT NULL COMMENT '执行耗时(毫秒)',
  `result` text COMMENT '执行结果',
  
  PRIMARY KEY (`log_id`),
  KEY `idx_task_time` (`task_id`, `execute_time`),
  KEY `idx_execute_time` (`execute_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行日志表';

-- =====================================================
-- 示例数据
-- =====================================================

-- 示例1: 定时任务 - 每5分钟执行一次日志记录
INSERT INTO `sys_tasks` (
  `task_name`, `task_type`, `is_enabled`, `description`,
  `interval_sec`, `action_type`, `action_config`
) VALUES (
  '定时统计任务', 1, 1, '每5分钟统计一次数据',
  300,
  5,
  '{"log_level":"INFO","message":"定时统计任务执行: 时间={{trigger_time}}"}'
);

-- 示例2: 数据改变任务 - 设备状态从 false 变为 true 时更新设备表
-- 假设 sys_variables 中有一个 ID=100 的布尔变量 "device_running"
INSERT INTO `sys_tasks` (
  `task_name`, `task_type`, `is_enabled`, `description`,
  `trigger_var_id`, `trigger_var_name`, `change_type`,
  `action_type`, `action_config`
) VALUES (
  '设备启动状态更新', 2, 1, '当设备运行状态从false变为true时，更新设备状态表',
  100, 'device_running', 'FALSE_TO_TRUE',
  3,
  '{
    "sql": "INSERT INTO sys_device_status (device_id, status, start_time) VALUES ({{device_id}}, 1, NOW())",
    "params": {"device_id": 1}
  }'
);

-- 示例3: 数据改变任务 - 设备状态从 true 变为 false 时关闭设备记录
INSERT INTO `sys_tasks` (
  `task_name`, `task_type`, `is_enabled`, `description`,
  `trigger_var_id`, `trigger_var_name`, `change_type`,
  `action_type`, `action_config`
) VALUES (
  '设备停止状态更新', 2, 1, '当设备运行状态从true变为false时，结束设备运行记录',
  100, 'device_running', 'TRUE_TO_FALSE',
  3,
  '{
    "sql": "UPDATE sys_device_status SET end_time=NOW(), duration_min=TIMESTAMPDIFF(MINUTE,start_time,NOW()) WHERE device_id={{device_id}} AND end_time IS NULL",
    "params": {"device_id": 1}
  }'
);

-- 示例4: 数据改变任务 - 温度变化超过5度时发送HTTP告警
INSERT INTO `sys_tasks` (
  `task_name`, `task_type`, `is_enabled`, `description`,
  `trigger_var_id`, `trigger_var_name`, `change_type`, `change_threshold`,
  `action_type`, `action_config`
) VALUES (
  '温度变化告警', 2, 0, '温度变化超过5度时发送HTTP告警',
  101, '温度传感器1', 'THRESHOLD', 5.0,
  1,
  '{
    "url": "http://localhost:8080/api/alerts",
    "method": "POST",
    "headers": {"Content-Type": "application/json"},
    "body": "{\\"type\\":\\"temperature_change\\",\\"old\\":{{old_value}},\\"new\\":{{new_value}},\\"change\\":{{change}}}",
    "timeout": 10
  }'
);

-- 示例5: 条件事件任务 - 温度>50且压力>100时发送MQTT消息
INSERT INTO `sys_tasks` (
  `task_name`, `task_type`, `is_enabled`, `description`,
  `condition_expr`, `action_type`, `action_config`
) VALUES (
  '高温高压告警', 3, 0, '温度>50且压力>100时发送MQTT消息',
  'temp>50 AND pressure>100',
  2,
  '{
    "topic": "alerts/critical",
    "payload": "{\\"alert\\":\\"high_temp_pressure\\",\\"temp\\":{{temp}},\\"pressure\\":{{pressure}}}",
    "qos": 1,
    "retain": false
  }'
);

-- 示例6: 执行脚本任务 - 设备故障时调用外部脚本
INSERT INTO `sys_tasks` (
  `task_name`, `task_type`, `is_enabled`, `description`,
  `trigger_var_id`, `trigger_var_name`, `change_type`,
  `action_type`, `action_config`
) VALUES (
  '设备故障处理脚本', 2, 0, '设备故障时调用外部脚本进行处理',
  102, 'device_fault', 'FALSE_TO_TRUE',
  4,
  '{
    "script_type": "bash",
    "script_path": "/scripts/handle_device_fault.sh",
    "args": ["{{device_id}}", "{{var_id}}"],
    "timeout": 60
  }'
);

