-- 测试能耗数据查询
-- 用于验证实时功率和今日电能功能

-- 1. 查看能耗相关变量配置
SELECT 
    id as var_id,
    device_id,
    var_name,
    display_name,
    data_type,
    unit,
    store_mode,
    store_cycle
FROM sys_variables
WHERE id IN (81, 86, 107, 110)
ORDER BY device_id, id;

-- 2. 查看今日历史数据记录数
SELECT 
    var_id,
    COUNT(*) as record_count,
    MIN(val) as min_val,
    MAX(val) as max_val,
    MAX(val) - MIN(val) as consumption,
    MIN(created_at) as first_time,
    MAX(created_at) as last_time
FROM sys_data_history
WHERE var_id IN (81, 107)  -- 电能变量
  AND created_at >= CURDATE()
  AND val IS NOT NULL
GROUP BY var_id;

-- 3. 查看最近的功率数据（设备1）
SELECT 
    val as power_value,
    created_at
FROM sys_data_history
WHERE var_id = 86  -- #1总有功功率
  AND created_at >= CURDATE()
ORDER BY created_at DESC
LIMIT 10;

-- 4. 查看最近的功率数据（设备2）
SELECT 
    val as power_value,
    created_at
FROM sys_data_history
WHERE var_id = 110  -- #2总有功功率
  AND created_at >= CURDATE()
ORDER BY created_at DESC
LIMIT 10;

-- 5. 查看最近的电能数据（设备1）
SELECT 
    val as energy_value,
    created_at
FROM sys_data_history
WHERE var_id = 81  -- #1当前总有功电能
  AND created_at >= CURDATE()
ORDER BY created_at DESC
LIMIT 10;

-- 6. 查看最近的电能数据（设备2）
SELECT 
    val as energy_value,
    created_at
FROM sys_data_history
WHERE var_id = 107  -- #2当前正向总有功电能
  AND created_at >= CURDATE()
ORDER BY created_at DESC
LIMIT 10;

-- 7. 测试今日电能消耗查询（模拟后端逻辑）
SELECT 
    81 as var_id,
    '#1当前总有功电能' as var_name,
    MAX(val) as max_val,
    MIN(val) as min_val,
    MAX(val) - MIN(val) as today_consumption
FROM sys_data_history
WHERE var_id = 81
  AND created_at >= CURDATE()
  AND val IS NOT NULL

UNION ALL

SELECT 
    107 as var_id,
    '#2当前正向总有功电能' as var_name,
    MAX(val) as max_val,
    MIN(val) as min_val,
    MAX(val) - MIN(val) as today_consumption
FROM sys_data_history
WHERE var_id = 107
  AND created_at >= CURDATE()
  AND val IS NOT NULL;

-- 8. 验证索引使用情况（性能测试）
EXPLAIN SELECT MAX(val), MIN(val)
FROM sys_data_history
WHERE var_id = 81 
  AND created_at >= CURDATE()
  AND val IS NOT NULL;


























