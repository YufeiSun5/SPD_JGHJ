-- 创建系统错误代码表
CREATE TABLE `sys_error_codes` (
  `error_code` int NOT NULL COMMENT '错误码',
  `error_msg` varchar(255) NOT NULL COMMENT '错误描述',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`error_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统错误代码表';

-- 插入示例错误码
INSERT INTO `sys_error_codes` (`error_code`, `error_msg`) VALUES
(1001, '设备通信超时'),
(1002, '传感器数据异常'),
(1003, '电机过载'),
(2001, 'PLC连接断开'),
(2002, 'Modbus读取失败'),
(3001, '数据库连接失败'),
(3002, '内存不足');
