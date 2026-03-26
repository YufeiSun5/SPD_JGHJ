-- ========================================================
-- MES系统性能优化 - 添加索引 (MySQL版本)
-- 执行时间: 根据数据量，预计1-5分钟
-- 注意: 如果索引已存在会报错，可以忽略
-- ========================================================

-- 1. 员工表索引
ALTER TABLE sys_staff ADD INDEX idx_staff_team (current_team_id);
ALTER TABLE sys_staff ADD INDEX idx_staff_active (is_active);

-- 2. 工单表索引
ALTER TABLE pro_orders ADD INDEX idx_order_device (target_device_id);
ALTER TABLE pro_orders ADD INDEX idx_order_status (status);
ALTER TABLE pro_orders ADD INDEX idx_order_created (created_at);

-- 3. 生产运行记录表索引
ALTER TABLE pro_production_runs ADD INDEX idx_run_device_time (device_id, start_time);
ALTER TABLE pro_production_runs ADD INDEX idx_run_team (team_id);
ALTER TABLE pro_production_runs ADD INDEX idx_run_active (end_time);

-- 4. 班次记录表索引
ALTER TABLE pro_machine_sessions ADD INDEX idx_session_device_logout (device_id, logout_time);
ALTER TABLE pro_machine_sessions ADD INDEX idx_session_team_time (team_id, login_time);
ALTER TABLE pro_machine_sessions ADD INDEX idx_session_login (login_time);

-- 查看索引创建结果
SHOW INDEX FROM sys_staff;
SHOW INDEX FROM pro_orders;
SHOW INDEX FROM pro_production_runs;
SHOW INDEX FROM pro_machine_sessions;



















