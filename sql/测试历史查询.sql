-- 测试历史查询功能
-- 确保数据库中有历史数据

-- 1. 检查变量配置
SELECT 
    id as var_id,
    var_name,
    display_name,
    unit,
    store_mode,
    CASE store_mode
        WHEN 0 THEN '不存储'
        WHEN 1 THEN '变动存储'
        WHEN 2 THEN '定时存储'
        WHEN 3 THEN '混合存储'
    END as store_mode_desc
FROM sys_variables
ORDER BY id;

-- 2. 检查历史数据表
SELECT 
    COUNT(*) as total_records,
    MIN(created_at) as earliest_time,
    MAX(created_at) as latest_time
FROM sys_data_history;

-- 3. 按变量统计历史数据
SELECT 
    v.id as var_id,
    v.var_name,
    v.display_name,
    COUNT(h.id) as record_count,
    MIN(h.created_at) as earliest_record,
    MAX(h.created_at) as latest_record
FROM sys_variables v
LEFT JOIN sys_data_history h ON v.id = h.var_id
GROUP BY v.id, v.var_name, v.display_name
ORDER BY v.id;

-- 4. 查看最近的历史数据（示例）
SELECT 
    h.id,
    h.var_id,
    v.var_name,
    v.display_name,
    COALESCE(h.val::text, h.str_val) as value,
    h.created_at
FROM sys_data_history h
JOIN sys_variables v ON h.var_id = v.id
ORDER BY h.created_at DESC
LIMIT 20;

-- 5. 如果没有历史数据，插入一些测试数据
-- 注意：只有在确认没有数据时才执行以下INSERT
/*
-- 假设 var_id=1 的变量存在
INSERT INTO sys_data_history (var_id, val, created_at)
SELECT 
    1 as var_id,
    (random() * 100)::numeric(10,2) as val,
    NOW() - (n || ' minutes')::interval as created_at
FROM generate_series(1, 100) as n;

-- 为其他变量也插入一些数据
INSERT INTO sys_data_history (var_id, val, created_at)
SELECT 
    v.id as var_id,
    (random() * 100)::numeric(10,2) as val,
    NOW() - (n || ' minutes')::interval as created_at
FROM sys_variables v
CROSS JOIN generate_series(1, 50) as n
WHERE v.store_mode > 0  -- 只为启用存储的变量插入数据
LIMIT 500;
*/

























