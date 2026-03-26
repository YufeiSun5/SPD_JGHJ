-- ========================================================
-- 创建班次考勤表
-- ========================================================

USE spd_jghj;

-- 删除旧表（如果存在）
DROP TABLE IF EXISTS `pro_machine_sessions`;

-- 创建班次记录表
CREATE TABLE `pro_machine_sessions` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `device_id` int NOT NULL COMMENT '登录设备ID',
  `team_id` int NOT NULL COMMENT '登录班组ID',
  `staff_ids` json NOT NULL COMMENT '当班员工ID列表 (如: [101, 102, 105])',
  `login_time` datetime NOT NULL COMMENT '上班/登录时间',
  `logout_time` datetime DEFAULT NULL COMMENT '下班/登出时间 (NULL代表正在上班)',
  `duration_min` int DEFAULT '0' COMMENT '上班时长(分钟)',
  PRIMARY KEY (`id`),
  KEY `idx_device_active` (`device_id`,`logout_time`),
  KEY `idx_team_time` (`team_id`,`login_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='【考勤表】设备登录与班次记录';

-- 插入一条测试数据（可选）
-- INSERT INTO `pro_machine_sessions` (`device_id`, `team_id`, `staff_ids`, `login_time`, `logout_time`, `duration_min`) 
-- VALUES (1, 2, '[1, 2]', '2025-12-12 08:00:00', '2025-12-12 17:00:00', 540);

SELECT '✅ 班次记录表创建成功！' as status;



























