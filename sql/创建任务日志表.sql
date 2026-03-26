-- =====================================================
-- 创建任务执行日志表
-- =====================================================

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
-- 查询任务执行日志
-- =====================================================

-- 查看最近的任务执行记录
SELECT 
  log_id,
  task_name,
  execute_time,
  CASE success WHEN 1 THEN '成功' ELSE '失败' END as status,
  duration as '耗时(ms)',
  LEFT(result, 100) as result_preview,
  error_msg
FROM sys_task_execution_logs 
ORDER BY execute_time DESC 
LIMIT 20;

-- 查看可写变量1相关的任务执行记录
SELECT 
  log_id,
  task_name,
  execute_time,
  CASE success WHEN 1 THEN '✅成功' ELSE '❌失败' END as status,
  duration as '耗时(ms)',
  result,
  error_msg
FROM sys_task_execution_logs 
WHERE task_name LIKE '%可写变量1%'
ORDER BY execute_time DESC 
LIMIT 20;

-- 统计任务执行情况
SELECT 
  task_name,
  COUNT(*) as '执行次数',
  SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as '成功次数',
  SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as '失败次数',
  AVG(duration) as '平均耗时(ms)',
  MAX(execute_time) as '最后执行时间'
FROM sys_task_execution_logs
GROUP BY task_name
ORDER BY MAX(execute_time) DESC;

-- 查看设备状态记录（验证数据库任务是否成功）
SELECT 
  id,
  device_id,
  CASE status 
    WHEN 0 THEN '空闲'
    WHEN 1 THEN '运行'
    WHEN 2 THEN '故障'
  END as status_name,
  start_time,
  end_time,
  duration_min,
  remark
FROM sys_device_status 
WHERE device_id = 1 
ORDER BY id DESC 
LIMIT 10;

