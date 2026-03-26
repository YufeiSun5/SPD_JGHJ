-- ===================================
-- 设备状态初始化数据
-- ===================================

-- 说明：这个脚本会初始化一些设备状态数据用于测试
-- 前提：确保 sys_devices 表中已经有设备数据

-- 1. 为现有设备初始化状态（示例：假设设备ID为1-6）
-- 运行中的设备
INSERT INTO sys_device_status (device_id, status, start_time, end_time, duration_min, remark)
VALUES 
(1, 1, DATE_SUB(NOW(), INTERVAL 2 HOUR), NULL, 0, '正常运行中'),
(2, 1, DATE_SUB(NOW(), INTERVAL 3 HOUR), NULL, 0, '正常运行中');

-- 空闲的设备
INSERT INTO sys_device_status (device_id, status, start_time, end_time, duration_min, remark)
VALUES 
(3, 0, DATE_SUB(NOW(), INTERVAL 30 MINUTE), NULL, 0, '待命中');

-- 故障的设备
INSERT INTO sys_device_status (device_id, status, start_time, end_time, duration_min, remark)
VALUES 
(4, 2, DATE_SUB(NOW(), INTERVAL 1 HOUR), NULL, 0, '传感器异常');

-- 2. 添加一些历史状态记录（用于24小时甘特图展示）
-- 设备1的历史记录
INSERT INTO sys_device_status (device_id, status, start_time, end_time, duration_min, remark)
VALUES 
-- 昨天的记录
(1, 0, DATE_SUB(NOW(), INTERVAL 24 HOUR), DATE_SUB(NOW(), INTERVAL 22 HOUR), 120, '夜班空闲'),
(1, 1, DATE_SUB(NOW(), INTERVAL 22 HOUR), DATE_SUB(NOW(), INTERVAL 18 HOUR), 240, '夜班生产'),
(1, 0, DATE_SUB(NOW(), INTERVAL 18 HOUR), DATE_SUB(NOW(), INTERVAL 16 HOUR), 120, '交接班'),
(1, 1, DATE_SUB(NOW(), INTERVAL 16 HOUR), DATE_SUB(NOW(), INTERVAL 10 HOUR), 360, '白班生产'),
(1, 2, DATE_SUB(NOW(), INTERVAL 10 HOUR), DATE_SUB(NOW(), INTERVAL 9 HOUR), 60, '设备故障'),
(1, 1, DATE_SUB(NOW(), INTERVAL 9 HOUR), DATE_SUB(NOW(), INTERVAL 2 HOUR), 420, '恢复生产');

-- 设备2的历史记录
INSERT INTO sys_device_status (device_id, status, start_time, end_time, duration_min, remark)
VALUES 
(2, 0, DATE_SUB(NOW(), INTERVAL 24 HOUR), DATE_SUB(NOW(), INTERVAL 20 HOUR), 240, '待机'),
(2, 1, DATE_SUB(NOW(), INTERVAL 20 HOUR), DATE_SUB(NOW(), INTERVAL 15 HOUR), 300, '连续生产'),
(2, 0, DATE_SUB(NOW(), INTERVAL 15 HOUR), DATE_SUB(NOW(), INTERVAL 14 HOUR), 60, '午休'),
(2, 1, DATE_SUB(NOW(), INTERVAL 14 HOUR), DATE_SUB(NOW(), INTERVAL 3 HOUR), 660, '下午生产');

-- 3. 今日统计数据示例（最近24小时内的各状态时长）
-- 设备5和6今日有多次状态切换
INSERT INTO sys_device_status (device_id, status, start_time, end_time, duration_min, remark)
VALUES 
-- 设备5
(5, 1, DATE_SUB(NOW(), INTERVAL 8 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR), 120, '上午生产'),
(5, 0, DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 4 HOUR), 120, '换模具'),
(5, 1, DATE_SUB(NOW(), INTERVAL 4 HOUR), DATE_SUB(NOW(), INTERVAL 1 HOUR), 180, '下午生产'),
(5, 0, DATE_SUB(NOW(), INTERVAL 1 HOUR), NULL, 0, '当前空闲'),

-- 设备6
(6, 1, DATE_SUB(NOW(), INTERVAL 10 HOUR), DATE_SUB(NOW(), INTERVAL 7 HOUR), 180, '正常运行'),
(6, 2, DATE_SUB(NOW(), INTERVAL 7 HOUR), DATE_SUB(NOW(), INTERVAL 6 HOUR), 60, '急停维修'),
(6, 1, DATE_SUB(NOW(), INTERVAL 6 HOUR), DATE_SUB(NOW(), INTERVAL 3 HOUR), 180, '恢复运行'),
(6, 0, DATE_SUB(NOW(), INTERVAL 3 HOUR), DATE_SUB(NOW(), INTERVAL 2 HOUR), 60, '等待工单'),
(6, 1, DATE_SUB(NOW(), INTERVAL 2 HOUR), NULL, 0, '当前运行中');

-- 4. 查询验证 - 查看当前所有设备状态
SELECT 
    ds.device_id,
    d.device_name,
    d.device_code,
    ds.status,
    CASE ds.status 
        WHEN 0 THEN '空闲'
        WHEN 1 THEN '运行'
        WHEN 2 THEN '故障'
        ELSE '未知'
    END as status_name,
    ds.start_time,
    ds.end_time,
    CASE 
        WHEN ds.end_time IS NULL THEN TIMESTAMPDIFF(MINUTE, ds.start_time, NOW())
        ELSE ds.duration_min
    END as duration_min,
    ds.remark
FROM sys_device_status ds
INNER JOIN sys_devices d ON ds.device_id = d.id
WHERE ds.end_time IS NULL
ORDER BY ds.device_id;

-- 5. 统计查询 - 今日各设备状态汇总
SELECT 
    d.id as device_id,
    d.device_name,
    SUM(CASE WHEN ds.status = 1 THEN 
        CASE 
            WHEN ds.end_time IS NULL THEN TIMESTAMPDIFF(MINUTE, ds.start_time, NOW())
            ELSE ds.duration_min
        END
        ELSE 0 
    END) as running_min,
    SUM(CASE WHEN ds.status = 0 THEN 
        CASE 
            WHEN ds.end_time IS NULL THEN TIMESTAMPDIFF(MINUTE, ds.start_time, NOW())
            ELSE ds.duration_min
        END
        ELSE 0 
    END) as idle_min,
    SUM(CASE WHEN ds.status = 2 THEN 
        CASE 
            WHEN ds.end_time IS NULL THEN TIMESTAMPDIFF(MINUTE, ds.start_time, NOW())
            ELSE ds.duration_min
        END
        ELSE 0 
    END) as fault_min
FROM sys_devices d
LEFT JOIN sys_device_status ds ON d.id = ds.device_id 
    AND ds.start_time >= DATE_SUB(NOW(), INTERVAL 24 HOUR)
GROUP BY d.id, d.device_name
ORDER BY d.id;

-- ===================================
-- 清理脚本（如需重新初始化，先运行此部分）
-- ===================================
/*
-- 清空设备状态表
TRUNCATE TABLE sys_device_status;

-- 或者只删除测试数据
DELETE FROM sys_device_status WHERE device_id <= 6;
*/

























