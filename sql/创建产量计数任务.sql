-- ========================================================
-- 创建产量计数任务
-- 日期: 2025-12-15
-- 说明: 设备1产量脉冲触发，自动更新工单和班次产量
-- ========================================================

-- 1. 确认产量脉冲变量存在
SELECT 
    id,
    var_name,
    display_name,
    data_type,
    store_mode,
    json_path
FROM sys_variables
WHERE var_name = '设备1产量加1';

-- 预期结果: id=65, data_type=BOOL, store_mode=1

-- ========================================================
-- 2. 创建产量计数任务（调用数据访问层）
-- ========================================================

INSERT INTO sys_tasks 
(task_name, task_type, is_enabled, description, 
 trigger_var_id, trigger_var_name, change_type, 
 action_type, action_config, created_at, updated_at)
VALUES 
('设备1产量计数', 2, 1, 
 '设备1产量脉冲0→1时，自动更新工单和班次的产量（调用数据访问层）',
 65,                    -- 触发变量ID
 '设备1产量加1',        -- 触发变量名
 'FALSE_TO_TRUE',       -- 0→1触发
 3,                     -- 数据库操作
 '{
  "operation": "increment_production_qty",
  "op_params": {
    "device_id": 1,
    "ok_qty_delta": 1,
    "ng_qty_delta": 0
  }
}',
 NOW(), NOW());

-- ========================================================
-- 3. 验证任务创建成功
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
WHERE task_name = '设备1产量计数';

-- ========================================================
-- 4. 使用说明
-- ========================================================

-- 步骤1: 在"人员管理"页面进行班次登记
--   - 选择设备1
--   - 选择班组
--   - 选择在班人员
--   - 点击"开始" → 创建 pro_machine_sessions 记录

-- 步骤2: 在"生产管理"页面点击工单的"开工"按钮
--   - 自动读取班次信息（设备、班组、人员）
--   - 自动暂停该设备的其他"生产中"工单
--   - 创建 pro_production_runs 记录
--   - 工单状态改为"生产中"

-- 步骤3: PLC 发送产量脉冲
--   - 设备1产量加1: 0 → 1
--   - 触发任务: increment_production_qty
--   - 自动更新:
--     • pro_production_runs.run_ok_qty +1 (班次产量)
--     • pro_orders.ok_qty +1 (工单良品数)
--     • pro_orders.actual_qty +1 (工单总产量)
--   - 使用事务保证数据一致性 ✅

-- ========================================================
-- 5. 检查班次和工单状态
-- ========================================================

-- 查看设备1的活动班次
SELECT 
    s.id,
    s.device_id,
    d.device_name,
    s.team_id,
    t.team_name,
    s.staff_ids,
    s.login_time,
    s.logout_time
FROM pro_machine_sessions s
LEFT JOIN sys_devices d ON s.device_id = d.id
LEFT JOIN sys_teams t ON s.team_id = t.id
WHERE s.device_id = 1 
  AND s.logout_time IS NULL;  -- 活动班次

-- 查看设备1的当前运行记录
SELECT 
    r.id,
    r.order_id,
    o.order_no,
    r.device_id,
    r.team_id,
    t.team_name,
    r.run_ok_qty,
    r.run_ng_qty,
    r.start_time,
    r.end_time,
    r.operator_ids
FROM pro_production_runs r
LEFT JOIN pro_orders o ON r.order_id = o.id
LEFT JOIN sys_teams t ON r.team_id = t.id
WHERE r.device_id = 1 
  AND r.end_time IS NULL;  -- 活动运行记录

-- 查看设备1的工单状态
SELECT 
    id,
    order_no,
    product_code,
    plan_qty,
    actual_qty,
    ok_qty,
    ng_qty,
    CONCAT(ROUND(actual_qty * 100.0 / plan_qty, 1), '%') AS completion_rate,
    status,
    CASE status
        WHEN 0 THEN '待产'
        WHEN 1 THEN '生产中'
        WHEN 2 THEN '暂停'
        WHEN 3 THEN '完工'
        WHEN 4 THEN '关闭'
    END AS status_name,
    start_time
FROM pro_orders
WHERE target_device_id = 1
ORDER BY id DESC;

-- ========================================================
-- 6. 测试产量累加（模拟PLC脉冲）
-- ========================================================

-- 方式1: 直接修改变量值触发任务（推荐）
-- 在你的SCADA软件中修改"设备1产量加1"变量: 0 → 1

-- 方式2: 手动执行增量（用于调试）
-- 调用数据访问层方法（需要有活动运行记录）
-- UPDATE pro_production_runs 
-- SET run_ok_qty = run_ok_qty + 1 
-- WHERE device_id = 1 AND end_time IS NULL;

-- UPDATE pro_orders o
-- INNER JOIN pro_production_runs r ON o.id = r.order_id
-- SET o.ok_qty = o.ok_qty + 1, o.actual_qty = o.actual_qty + 1
-- WHERE r.device_id = 1 AND r.end_time IS NULL;

-- ========================================================
-- 7. 查看任务执行日志
-- ========================================================

SELECT 
    log_id,
    task_name,
    execute_time,
    success,
    duration,
    result,
    error_msg
FROM sys_task_execution_logs
WHERE task_name = '设备1产量计数'
ORDER BY execute_time DESC
LIMIT 10;





















