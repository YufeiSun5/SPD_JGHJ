-- ========================================================
-- 更新任务配置：从原始SQL改为预定义操作
-- 日期: 2025-12-15
-- 说明: 将数据库操作任务改为调用数据访问层函数
-- ========================================================

-- 1. 查看当前使用原始SQL的任务
SELECT 
    task_id,
    task_name,
    action_type,
    action_config
FROM sys_tasks
WHERE action_type = 3  -- 数据库操作
  AND is_enabled = 1
ORDER BY task_id;

-- ========================================================
-- 2. 更新"可写变量1-数据库操作"任务 (task_id=17)
-- 从: INSERT INTO sys_device_status ...
-- 到: 调用 UpdateDeviceStatus 函数
-- ========================================================

UPDATE sys_tasks 
SET 
    task_name = '可写变量1-更新设备状态',
    description = '可写变量1变为1时，调用数据访问层更新设备状态为运行',
    action_config = '{
  "operation": "update_device_status",
  "op_params": {
    "device_id": 1,
    "status": 1,
    "remark": "可写变量1触发运行"
  }
}'
WHERE task_id = 17;

-- ========================================================
-- 3. 示例：创建一个"结束设备状态"的任务
-- 当可写变量1从1变为0时，结束设备当前状态
-- ========================================================

INSERT INTO sys_tasks 
(task_name, task_type, is_enabled, description, 
 trigger_var_id, trigger_var_name, change_type, 
 action_type, action_config, created_at, updated_at)
VALUES 
('可写变量1-结束设备状态', 2, 1, '可写变量1变为0时结束当前设备状态',
 1, '可写变量1', 'TRUE_TO_FALSE', 
 3,
 '{
  "operation": "end_device_status",
  "op_params": {
    "device_id": 1,
    "remark": "可写变量1触发停止"
  }
}',
 NOW(), NOW());

-- ========================================================
-- 4. 示例：为设备2创建类似的任务
-- ========================================================

INSERT INTO sys_tasks 
(task_name, task_type, is_enabled, description, 
 trigger_var_id, trigger_var_name, change_type, 
 action_type, action_config, created_at, updated_at)
VALUES 
('可写变量2-更新设备状态', 2, 1, '可写变量2变为1时更新设备2状态为运行',
 2, '可写变量2', 'FALSE_TO_TRUE', 
 3,
 '{
  "operation": "update_device_status",
  "op_params": {
    "device_id": 2,
    "status": 1,
    "remark": "可写变量2触发运行"
  }
}',
 NOW(), NOW());

-- ========================================================
-- 5. 验证更新结果
-- ========================================================

SELECT 
    task_id,
    task_name,
    task_type,
    trigger_var_name,
    change_type,
    action_type,
    action_config,
    is_enabled
FROM sys_tasks
WHERE action_type = 3  -- 数据库操作
ORDER BY task_id;

-- ========================================================
-- 6. 测试任务执行日志（运行一段时间后查看）
-- ========================================================

SELECT 
    log_id,
    task_id,
    task_name,
    execute_time,
    success,
    error_msg,
    duration,
    result
FROM task_execution_logs
WHERE task_id IN (SELECT task_id FROM sys_tasks WHERE action_type = 3)
ORDER BY execute_time DESC
LIMIT 20;

-- ========================================================
-- 说明：预定义操作 vs 原始SQL
-- ========================================================

-- ❌ 旧方式（原始SQL）：
-- {
--   "sql": "INSERT INTO sys_device_status (device_id, status, start_time, remark) VALUES (1, 1, NOW(), '可写变量1触发')"
-- }
-- 问题：
-- 1. 不会自动结束上一个状态
-- 2. 没有事务保护
-- 3. 业务逻辑分散

-- ✅ 新方式（预定义操作）：
-- {
--   "operation": "update_device_status",
--   "op_params": {
--     "device_id": 1,
--     "status": 1,
--     "remark": "可写变量1触发运行"
--   }
-- }
-- 优势：
-- 1. 自动结束上一个状态（设置 end_time 和 duration_min）
-- 2. 使用事务保证数据一致性
-- 3. 业务逻辑集中在数据访问层
-- 4. 代码可维护性高

-- ========================================================
-- 支持的预定义操作列表（当前已实现）
-- ========================================================

-- 1. update_device_status - 更新设备状态
--    参数: device_id (int), status (int8: 0=空闲,1=运行,2=故障), remark (string, 可选)
--
-- 2. end_device_status - 结束设备当前状态
--    参数: device_id (int), remark (string, 可选)
--
-- 未来可扩展：
-- 3. start_order - 开始工单
-- 4. complete_order - 完成工单
-- 5. record_production - 记录生产数据





















