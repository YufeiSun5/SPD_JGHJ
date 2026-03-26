-- ========================================================
-- 创建设备1 NG计数功能
-- 日期: 2025-12-15
-- 说明: 设备1 NG脉冲触发，自动增加不良品数量
-- ========================================================

-- 1. 创建"设备1NG加1"变量
INSERT INTO `sys_variables` 
(`device_id`, `var_name`, `display_name`, `data_type`, `rw_mode`, `unit`, 
 `json_path`, `scale_factor`, `offset_val`, `alarm_enable`, 
 `store_mode`, `store_cycle`, `store_deadband`, `source_type`, `calc_rule`) 
VALUES 
(1, '设备1NG加1', '设备1NG脉冲', 'BOOL', 'RW', NULL, 
 'Objs.#(N=="设备1NG加1").1', 1, 0, 0, 
 1, 1, 0, 0, NULL);

-- 获取刚创建的变量ID（用于后续任务配置）
SELECT id, var_name FROM sys_variables WHERE var_name = '设备1NG加1';

-- ========================================================
-- 2. 创建NG计数任务（使用上面查询到的变量ID）
-- ========================================================

-- 注意：请将下面的 66 替换为上面查询到的实际变量ID
INSERT INTO sys_tasks 
(task_name, task_type, is_enabled, description, 
 trigger_var_id, trigger_var_name, change_type, 
 action_type, action_config, created_at, updated_at)
VALUES 
('设备1NG计数', 2, 1, 
 '设备1 NG脉冲0→1时，自动增加工单和班次的不良品数量',
 66,                    -- ⚠️ 请替换为实际的变量ID
 '设备1NG加1',          -- 触发变量名
 'FALSE_TO_TRUE',       -- 0→1触发
 3,                     -- 数据库操作
 '{
  "operation": "increment_production_qty",
  "op_params": {
    "device_id": 1,
    "ok_qty_delta": 0,
    "ng_qty_delta": 1
  }
}',
 NOW(), NOW());

-- ========================================================
-- 3. 验证创建结果
-- ========================================================

-- 查看变量
SELECT 
    id,
    var_name,
    display_name,
    data_type,
    store_mode,
    json_path
FROM sys_variables
WHERE var_name IN ('设备1产量加1', '设备1NG加1')
ORDER BY id;

-- 查看任务
SELECT 
    task_id,
    task_name,
    trigger_var_name,
    change_type,
    action_config,
    is_enabled
FROM sys_tasks
WHERE task_name IN ('设备1产量计数', '设备1NG计数')
ORDER BY task_id;

-- ========================================================
-- 4. 使用说明
-- ========================================================

-- 现在你有两个脉冲变量：
-- 
-- 1. "设备1产量加1" (ID=65)：
--    0→1 触发 → OK数量+1
--    更新: pro_production_runs.run_ok_qty +1
--          pro_orders.ok_qty +1
--          pro_orders.actual_qty +1
--
-- 2. "设备1NG加1" (ID=66)：
--    0→1 触发 → NG数量+1  
--    更新: pro_production_runs.run_ng_qty +1
--          pro_orders.ng_qty +1
--          pro_orders.actual_qty +1
--
-- 两个任务都会调用同一个数据访问层方法：
-- IncrementProductionQtyByDevice(deviceID, okDelta, ngDelta)
-- 
-- 该方法会：
-- ✅ 查找设备当前活动的运行记录
-- ✅ 使用事务同时更新班次和工单
-- ✅ 保证数据一致性

-- ========================================================
-- 5. 测试脉冲效果
-- ========================================================

-- 测试前：查看当前数量
SELECT 
    o.order_no,
    o.ok_qty,
    o.ng_qty,
    o.actual_qty,
    r.run_ok_qty,
    r.run_ng_qty
FROM pro_orders o
LEFT JOIN pro_production_runs r ON o.id = r.order_id AND r.end_time IS NULL
WHERE o.target_device_id = 1 AND o.status = 1;

-- 测试步骤：
-- 1. 在SCADA中修改"设备1产量加1": 0 → 1
--    预期: OK+1, 总数+1
--
-- 2. 在SCADA中修改"设备1NG加1": 0 → 1  
--    预期: NG+1, 总数+1

-- 测试后：再次查看数量变化
SELECT 
    o.order_no,
    o.ok_qty,
    o.ng_qty,
    o.actual_qty,
    r.run_ok_qty,
    r.run_ng_qty,
    CONCAT(ROUND(o.actual_qty * 100.0 / o.plan_qty, 1), '%') AS completion_rate
FROM pro_orders o
LEFT JOIN pro_production_runs r ON o.id = r.order_id AND r.end_time IS NULL
WHERE o.target_device_id = 1 AND o.status = 1;

-- ========================================================
-- 6. 查看任务执行日志
-- ========================================================

SELECT 
    task_name,
    execute_time,
    success,
    duration,
    result,
    error_msg
FROM sys_task_execution_logs
WHERE task_name IN ('设备1产量计数', '设备1NG计数')
ORDER BY execute_time DESC
LIMIT 10;

-- ========================================================
-- 7. PLC/SCADA 配置说明
-- ========================================================

-- 在你的 PLC/SCADA 系统中需要配置两个可写变量：
--
-- 变量1: "设备1产量加1"
--   - 数据类型: BOOL
--   - 触发方式: 生产一个合格品时，脉冲 0→1→0
--   - JSON路径: Objs.#(N=="设备1产量加1").1
--
-- 变量2: "设备1NG加1" 
--   - 数据类型: BOOL
--   - 触发方式: 检测到不良品时，脉冲 0→1→0
--   - JSON路径: Objs.#(N=="设备1NG加1").1
--
-- MQTT消息格式示例:
-- {
--   "Writer": "sa",
--   "WriteTime": "2025-12-15T10:30:00",
--   "Username": "sa", 
--   "Password": "...",
--   "Qid": 123456,
--   "PNs": {"1": "V", "2": "T", "3": "Q"},
--   "PVs": {"1": 1, "2": "2025-12-15T10:30:00", "3": 192},
--   "Objs": [
--     {"N": "设备1产量加1", "1": 1, "2": 123456, "3": 192},
--     {"N": "设备1NG加1", "1": 0, "2": 123456, "3": 192}
--   ]
-- }




















