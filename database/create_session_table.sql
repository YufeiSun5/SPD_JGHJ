-- ========================================================
-- 设备登录/班次记录表 (考勤表)
-- ========================================================

DROP TABLE IF EXISTS `pro_machine_sessions`;

CREATE TABLE `pro_machine_sessions` (
  `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
  
  -- 1. 谁在哪里
  `device_id` INT NOT NULL COMMENT '登录设备ID',
  `team_id` INT NOT NULL COMMENT '登录班组ID',
  
  -- 2. 都有谁 (这是这一班次的具体出勤名单)
  -- 比如班组里有5个人，今天只有4个人来上班，这里就只存这4个人的ID
  `staff_ids` JSON NOT NULL COMMENT '当班员工ID列表 (如: [101, 102, 105])',
  
  -- 3. 时间段 (打卡时间)
  `login_time` DATETIME NOT NULL COMMENT '上班/登录时间',
  `logout_time` DATETIME DEFAULT NULL COMMENT '下班/登出时间 (NULL代表正在上班)',
  
  -- 4. 统计 (可选，方便算工时)
  `duration_min` INT DEFAULT 0 COMMENT '上班时长(分钟)',
  
  -- 索引：快速查某台设备当前是谁在登录
  INDEX `idx_device_active` (`device_id`, `logout_time`),
  -- 索引：查某班组的历史出勤
  INDEX `idx_team_time` (`team_id`, `login_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='【考勤表】设备登录与班次记录';

-- 插入示例数据
INSERT INTO `pro_machine_sessions` (`device_id`, `team_id`, `staff_ids`, `login_time`, `logout_time`, `duration_min`) VALUES
(1, 1, '[1, 2]', '2025-12-12 08:00:00', '2025-12-12 20:00:00', 720),
(1, 2, '[3]', '2025-12-11 20:00:00', '2025-12-12 08:00:00', 720);



























