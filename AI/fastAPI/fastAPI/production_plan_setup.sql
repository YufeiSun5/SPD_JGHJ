-- 生产计划表
CREATE TABLE IF NOT EXISTS production_plan (
    id INT AUTO_INCREMENT PRIMARY KEY,
    plan_date DATE NOT NULL COMMENT '计划日期',
    model_name VARCHAR(100) NOT NULL COMMENT '机型名称',
    planned_quantity INT NOT NULL DEFAULT 0 COMMENT '计划生产数量',
    actual_quantity INT NOT NULL DEFAULT 0 COMMENT '实际生产数量',
    production_line VARCHAR(50) COMMENT '生产线',
    shift_type ENUM('morning', 'afternoon', 'night') COMMENT '班次',
    status ENUM('pending', 'in_progress', 'completed', 'cancelled') DEFAULT 'pending' COMMENT '状态',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_plan_date_model (plan_date, model_name),
    INDEX idx_production_line (production_line),
    INDEX idx_status (status)
) COMMENT='生产计划表';

-- 插入示例数据
INSERT INTO production_plan (plan_date, model_name, planned_quantity, actual_quantity, production_line, shift_type, status) VALUES
-- 今天的计划
(CURDATE(), 'iPhone15', 1000, 950, 'Line-A', 'morning', 'completed'),
(CURDATE(), 'iPhone15', 800, 780, 'Line-B', 'afternoon', 'completed'),
(CURDATE(), 'SamsungS23', 600, 0, 'Line-C', 'night', 'pending'),

-- 明天的计划
(DATE_ADD(CURDATE(), INTERVAL 1 DAY), 'iPhone15', 1200, 0, 'Line-A', 'morning', 'pending'),
(DATE_ADD(CURDATE(), INTERVAL 1 DAY), 'SamsungS23', 900, 0, 'Line-B', 'afternoon', 'pending'),
(DATE_ADD(CURDATE(), INTERVAL 1 DAY), 'HuaweiP60', 700, 0, 'Line-C', 'night', 'pending'),

-- 本周的历史计划
(DATE_SUB(CURDATE(), INTERVAL 1 DAY), 'iPhone15', 1100, 1080, 'Line-A', 'morning', 'completed'),
(DATE_SUB(CURDATE(), INTERVAL 1 DAY), 'SamsungS23', 850, 840, 'Line-B', 'afternoon', 'completed'),
(DATE_SUB(CURDATE(), INTERVAL 2 DAY), 'iPhone15', 950, 920, 'Line-A', 'morning', 'completed'),
(DATE_SUB(CURDATE(), INTERVAL 2 DAY), 'HuaweiP60', 600, 580, 'Line-C', 'night', 'completed');

-- 生成更多历史数据（最近30天）
INSERT INTO production_plan (plan_date, model_name, planned_quantity, actual_quantity, production_line, shift_type, status)
SELECT 
    DATE_SUB(CURDATE(), INTERVAL seq DAY) as plan_date,
    model_name,
    FLOOR(800 + RAND() * 400) as planned_quantity,
    FLOOR(750 + RAND() * 350) as actual_quantity,
    production_line,
    shift_type,
    'completed' as status
FROM (
    SELECT 3 seq UNION SELECT 4 UNION SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION 
    SELECT 8 UNION SELECT 9 UNION SELECT 10 UNION SELECT 11 UNION SELECT 12 UNION
    SELECT 13 UNION SELECT 14 UNION SELECT 15 UNION SELECT 16 UNION SELECT 17 UNION
    SELECT 18 UNION SELECT 19 UNION SELECT 20 UNION SELECT 21 UNION SELECT 22 UNION
    SELECT 23 UNION SELECT 24 UNION SELECT 25 UNION SELECT 26 UNION SELECT 27 UNION
    SELECT 28 UNION SELECT 29 UNION SELECT 30
) seq_table
CROSS JOIN (
    SELECT 'iPhone15' as model_name, 'Line-A' as production_line, 'morning' as shift_type UNION 
    SELECT 'SamsungS23', 'Line-B', 'afternoon' UNION 
    SELECT 'HuaweiP60', 'Line-C', 'night' UNION 
    SELECT 'XiaomiMi13', 'Line-A', 'afternoon'
) plan_table;
