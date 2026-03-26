-- =====================================================
-- 快速修复：确保表存在并插入测试任务
-- =====================================================

-- 步骤1：确保表存在（执行一次即可）
CREATE TABLE IF NOT EXISTS `sys_tasks` (
  `task_id` bigint NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `task_name` varchar(100) NOT NULL COMMENT '任务名称',
  `task_type` tinyint NOT NULL COMMENT '任务类型: 1=定时, 2=数据改变, 3=条件事件',
  `is_enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用',
  `description` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `cron_expr` varchar(50) DEFAULT NULL,
  `interval_sec` int DEFAULT NULL,
  `last_run_time` datetime DEFAULT NULL,
  `trigger_var_id` bigint DEFAULT NULL,
  `trigger_var_name` varchar(100) DEFAULT NULL,
  `change_type` varchar(20) DEFAULT 'ANY',
  `change_threshold` float DEFAULT NULL,
  `condition_expr` varchar(500) DEFAULT NULL,
  `action_type` tinyint NOT NULL,
  `action_config` text NOT NULL,
  PRIMARY KEY (`task_id`),
  KEY `idx_type_enabled` (`task_type`, `is_enabled`),
  KEY `idx_trigger_var` (`trigger_var_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务调度配置表';

CREATE TABLE IF NOT EXISTS `sys_task_execution_logs` (
  `log_id` bigint NOT NULL AUTO_INCREMENT,
  `task_id` bigint NOT NULL,
  `task_name` varchar(100) NOT NULL,
  `execute_time` datetime NOT NULL,
  `success` tinyint NOT NULL,
  `error_msg` text,
  `duration` int DEFAULT NULL,
  `result` text,
  PRIMARY KEY (`log_id`),
  KEY `idx_task_time` (`task_id`, `execute_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 步骤2：清空旧任务
DELETE FROM sys_tasks WHERE trigger_var_name = '可写变量1';

-- 步骤3：插入简单的日志测试任务（最容易验证）
INSERT INTO sys_tasks (
  task_name, task_type, is_enabled, description,
  trigger_var_id, trigger_var_name, change_type,
  action_type, action_config
) VALUES (
  '可写变量1-日志测试',
  2, 1, '可写变量1从0变为1时输出日志',
  1, '可写变量1', 'FALSE_TO_TRUE',
  5,
  '{"log_level": "INFO", "message": "🔥🔥🔥 可写变量1触发成功！old={{old_value}}, new={{new_value}}"}'
);

-- 步骤4：插入数据库操作任务
INSERT INTO sys_tasks (
  task_name, task_type, is_enabled, description,
  trigger_var_id, trigger_var_name, change_type,
  action_type, action_config
) VALUES (
  '可写变量1-数据库操作',
  2, 1, '可写变量1变为1时记录设备状态',
  1, '可写变量1', 'FALSE_TO_TRUE',
  3,
  '{"sql": "INSERT INTO sys_device_status (device_id, status, start_time, remark) VALUES (1, 1, NOW(), ''可写变量1触发'')"}'
);

-- 步骤5：插入MQTT发布任务
INSERT INTO sys_tasks (
  task_name, task_type, is_enabled, description,
  trigger_var_id, trigger_var_name, change_type,
  action_type, action_config
) VALUES (
  '可写变量1-MQTT写入变量2',
  2, 1, '可写变量1变为1时写入可写变量2为999',
  1, '可写变量1', 'FALSE_TO_TRUE',
  2,
  '{"topic": "setdata_S_KIO_Project", "payload": "{\"Writer\":\"AutoTask\",\"WriteTime\":\"2025-12-13 23:40:00\",\"Username\":\"admin\",\"Password\":\"admin\",\"Qid\":1001,\"PNs\":{\"1\":\"V\"},\"PVs\":{\"1\":999},\"Objs\":[{\"N\":\"可写变量2\",\"1\":999}]}", "qos": 1, "retain": false}'
);

-- 步骤6：验证任务已插入
SELECT task_id, task_name, is_enabled, trigger_var_id, trigger_var_name, change_type, action_type
FROM sys_tasks 
WHERE trigger_var_name = '可写变量1';

-- 步骤7：触发热重载
UPDATE sys_config_version 
SET version_code = CONCAT('v', UNIX_TIMESTAMP()) 
WHERE id = 1;

-- 步骤8：检查配置版本
SELECT id, version_code, updated_at FROM sys_config_version WHERE id = 1;

