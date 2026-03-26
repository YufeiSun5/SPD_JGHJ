-- 创建生产质量数据表
CREATE TABLE IF NOT EXISTS production_quality (
    id INT AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL COMMENT '生产日期',
    model_name VARCHAR(100) NOT NULL COMMENT '机型名称',
    total_count INT NOT NULL DEFAULT 0 COMMENT '总生产数量',
    defect_count INT NOT NULL DEFAULT 0 COMMENT '不良品数量',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    INDEX idx_date_model (date, model_name),
    INDEX idx_date (date)
) COMMENT='生产质量数据表';

-- 插入示例数据
INSERT INTO production_quality (date, model_name, total_count, defect_count) VALUES
-- iPhone15 数据
('2024-01-15', 'iPhone15', 1000, 25),
('2024-01-16', 'iPhone15', 1200, 30),
('2024-01-17', 'iPhone15', 950, 20),
('2024-01-18', 'iPhone15', 1100, 28),
('2024-01-19', 'iPhone15', 1050, 22),

-- SamsungS23 数据
('2024-01-15', 'SamsungS23', 800, 18),
('2024-01-16', 'SamsungS23', 900, 22),
('2024-01-17', 'SamsungS23', 850, 19),
('2024-01-18', 'SamsungS23', 920, 24),
('2024-01-19', 'SamsungS23', 880, 20),

-- 今天的数据（可以根据需要调整日期）
(CURDATE(), 'iPhone15', 1150, 32),
(CURDATE(), 'SamsungS23', 890, 21),

-- 昨天的数据
(DATE_SUB(CURDATE(), INTERVAL 1 DAY), 'iPhone15', 1080, 26),
(DATE_SUB(CURDATE(), INTERVAL 1 DAY), 'SamsungS23', 910, 23);

-- 创建更多历史数据（最近30天）
INSERT INTO production_quality (date, model_name, total_count, defect_count)
SELECT 
    DATE_SUB(CURDATE(), INTERVAL seq DAY) as date,
    model_name,
    FLOOR(900 + RAND() * 300) as total_count,
    FLOOR(15 + RAND() * 20) as defect_count
FROM (
    SELECT 0 seq UNION SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION 
    SELECT 5 UNION SELECT 6 UNION SELECT 7 UNION SELECT 8 UNION SELECT 9 UNION
    SELECT 10 UNION SELECT 11 UNION SELECT 12 UNION SELECT 13 UNION SELECT 14 UNION
    SELECT 15 UNION SELECT 16 UNION SELECT 17 UNION SELECT 18 UNION SELECT 19 UNION
    SELECT 20 UNION SELECT 21 UNION SELECT 22 UNION SELECT 23 UNION SELECT 24 UNION
    SELECT 25 UNION SELECT 26 UNION SELECT 27 UNION SELECT 28 UNION SELECT 29
) seq_table
CROSS JOIN (
    SELECT 'iPhone15' as model_name UNION 
    SELECT 'SamsungS23' UNION 
    SELECT 'HuaweiP60' UNION 
    SELECT 'XiaomiMi13'
) model_table
WHERE DATE_SUB(CURDATE(), INTERVAL seq DAY) NOT IN (
    SELECT DISTINCT date FROM production_quality 
    WHERE model_name = model_table.model_name
);
