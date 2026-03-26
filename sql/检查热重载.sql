-- =============================================
-- 热重载配置检查 SQL 脚本
-- =============================================

-- 1. 检查 sys_config_version 表是否存在
SELECT table_name 
FROM information_schema.tables 
WHERE table_name = 'sys_config_version';

-- 2. 查看当前版本记录
SELECT * FROM sys_config_version;

-- 3. 如果表不存在或没有数据，执行以下创建和初始化
-- CREATE TABLE IF NOT EXISTS sys_config_version (
--     id INT PRIMARY KEY,
--     version_code VARCHAR(100) NOT NULL,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
-- );

-- INSERT INTO sys_config_version (id, version_code) 
-- VALUES (1, 'v1.0.0-initial')
-- ON CONFLICT (id) DO NOTHING;

-- =============================================
-- 4. 触发热重载 - 更新版本号
-- =============================================
UPDATE sys_config_version 
SET version_code = 'v1.0.1-test-reload',
    updated_at = CURRENT_TIMESTAMP
WHERE id = 1;

-- 5. 验证更新成功
SELECT * FROM sys_config_version;

-- =============================================
-- 6. 测试修改配置 + 触发重载
-- =============================================

-- 示例：修改温度报警阈值
-- UPDATE sys_variables 
-- SET limit_hh = 88.0, 
--     limit_h = 78.0 
-- WHERE var_name = 'temp_inlet';

-- 然后触发重载
-- UPDATE sys_config_version 
-- SET version_code = 'v1.0.2-change-alarm-limit',
--     updated_at = CURRENT_TIMESTAMP
-- WHERE id = 1;

-- =============================================
-- 7. 查询测点配置验证
-- =============================================
SELECT 
    var_name,
    display_name,
    alarm_enable,
    limit_hh,
    limit_h,
    limit_l,
    limit_ll
FROM sys_variables 
WHERE alarm_enable = true
ORDER BY var_name;

























