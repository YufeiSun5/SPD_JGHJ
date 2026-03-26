-- ===================================
-- 报警记录表 - 添加阈值字段
-- ===================================
-- 用途: 记录触发报警时被超过的阈值
-- 例如: 温度86℃触发HH报警，limit_value=85.0
-- ===================================

-- 添加 limit_value 列
ALTER TABLE sys_alarm_records 
ADD COLUMN limit_value DOUBLE PRECISION NULL 
COMMENT '被超过的阈值 (HH/H/L/LL对应的limit值)';

-- 创建索引以加速查询
CREATE INDEX idx_alarm_records_limit_value ON sys_alarm_records(limit_value);

-- 示例查询: 查看所有报警及其阈值
/*
SELECT 
    id,
    var_name,
    alarm_type,
    val AS '实际值',
    limit_value AS '阈值',
    (val - limit_value) AS '超出值',
    start_time,
    end_time,
    CASE 
        WHEN end_time IS NULL THEN '报警中'
        ELSE '已恢复'
    END AS '状态'
FROM sys_alarm_records
ORDER BY start_time DESC
LIMIT 20;
*/

