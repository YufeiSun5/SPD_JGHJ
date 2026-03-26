-- =====================================================
-- 修复 MQTT 任务配置 - 使用正确的用户名和密码
-- =====================================================

-- 删除旧的MQTT任务
DELETE FROM sys_tasks WHERE task_name IN (
  '可写变量1-MQTT写入变量2',
  '可写变量1-自动复位'
);

-- =====================================================
-- 任务1：可写变量1变为1时，通过MQTT将可写变量2写为999
-- =====================================================
INSERT INTO sys_tasks (
  task_name, task_type, is_enabled, description,
  trigger_var_id, trigger_var_name, change_type,
  action_type, action_config
) VALUES (
  '可写变量1-MQTT写入变量2',
  2, 1, '可写变量1从0变为1时，通过MQTT将可写变量2写入为999',
  1, '可写变量1', 'FALSE_TO_TRUE',
  2,
  '{"topic":"setdata_S_KIO_Project","payload":"{\\"Writer\\":\\"sa\\",\\"WriteTime\\":\\"{{trigger_time}}\\",\\"Username\\":\\"sa\\",\\"Password\\":\\"C12E01F2A13FF5587E1E9E4AEDB8242D\\",\\"Qid\\":197296731,\\"PNs\\":{\\"1\\":\\"V\\",\\"2\\":\\"T\\",\\"3\\":\\"Q\\"},\\"PVs\\":{\\"1\\":999,\\"2\\":\\"{{trigger_time}}\\",\\"3\\":192},\\"Objs\\":[{\\"N\\":\\"可写变量2\\",\\"1\\":999,\\"2\\":123542,\\"3\\":192}]}","qos":1,"retain":false}'
);

-- =====================================================
-- 任务2：可写变量1变为1后，延迟1秒自动复位为0
-- =====================================================
INSERT INTO sys_tasks (
  task_name, task_type, is_enabled, description,
  trigger_var_id, trigger_var_name, change_type,
  action_type, action_config
) VALUES (
  '可写变量1-自动复位',
  2, 1, '可写变量1变为1后自动复位为0',
  1, '可写变量1', 'FALSE_TO_TRUE',
  2,
  '{"topic":"setdata_S_KIO_Project","payload":"{\\"Writer\\":\\"sa\\",\\"WriteTime\\":\\"{{trigger_time}}\\",\\"Username\\":\\"sa\\",\\"Password\\":\\"C12E01F2A13FF5587E1E9E4AEDB8242D\\",\\"Qid\\":197296732,\\"PNs\\":{\\"1\\":\\"V\\",\\"2\\":\\"T\\",\\"3\\":\\"Q\\"},\\"PVs\\":{\\"1\\":0,\\"2\\":\\"{{trigger_time}}\\",\\"3\\":192},\\"Objs\\":[{\\"N\\":\\"可写变量1\\",\\"1\\":0,\\"2\\":123542,\\"3\\":192}]}","qos":1,"retain":false}'
);

-- =====================================================
-- 触发热重载
-- =====================================================
UPDATE sys_config_version 
SET version_code = CONCAT('v', UNIX_TIMESTAMP()) 
WHERE id = 1;

-- =====================================================
-- 验证任务配置
-- =====================================================
SELECT 
  task_id,
  task_name,
  is_enabled,
  trigger_var_name,
  change_type,
  action_type,
  CASE action_type
    WHEN 1 THEN 'HTTP请求'
    WHEN 2 THEN 'MQTT发布'
    WHEN 3 THEN '数据库操作'
    WHEN 4 THEN '脚本执行'
    WHEN 5 THEN '日志写入'
  END AS action_name
FROM sys_tasks 
WHERE trigger_var_name = '可写变量1'
ORDER BY task_id;

