-- 为 pro_orders 表添加时间跟踪字段（安全版本 - 检查字段是否存在）
-- 用于准确统计订单的实际用时（支持暂停/继续）

-- 检查并添加 used_seconds 字段
SET @col_exists = 0;
SELECT COUNT(*) INTO @col_exists 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'pro_orders' 
  AND COLUMN_NAME = 'used_seconds';

SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE pro_orders ADD COLUMN used_seconds INT DEFAULT 0 COMMENT ''已使用秒数(累计)''', 
    'SELECT ''Column used_seconds already exists'' AS Info');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 检查并添加 current_start_time 字段
SET @col_exists = 0;
SELECT COUNT(*) INTO @col_exists 
FROM INFORMATION_SCHEMA.COLUMNS 
WHERE TABLE_SCHEMA = DATABASE() 
  AND TABLE_NAME = 'pro_orders' 
  AND COLUMN_NAME = 'current_start_time';

SET @sql = IF(@col_exists = 0, 
    'ALTER TABLE pro_orders ADD COLUMN current_start_time DATETIME COMMENT ''当前开始时间(用于计算本次用时)''', 
    'SELECT ''Column current_start_time already exists'' AS Info');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 为现有数据设置默认值
UPDATE pro_orders SET used_seconds = 0 WHERE used_seconds IS NULL;

