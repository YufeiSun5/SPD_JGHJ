-- 为 pro_orders 表添加时间跟踪字段
-- 用于准确统计订单的实际用时（支持暂停/继续）

-- 添加已使用秒数字段（累计）
ALTER TABLE pro_orders ADD COLUMN used_seconds INT DEFAULT 0 COMMENT '已使用秒数(累计)';

-- 添加当前开始时间字段（用于计算本次用时）
ALTER TABLE pro_orders ADD COLUMN current_start_time DATETIME COMMENT '当前开始时间(用于计算本次用时)';

-- 为现有数据设置默认值（如果表中已有数据）
UPDATE pro_orders SET used_seconds = 0 WHERE used_seconds IS NULL;

