/*
 Navicat Premium Data Transfer

 Source Server         : mysql0
 Source Server Type    : MySQL
 Source Server Version : 80028
 Source Host           : localhost:3306
 Source Schema         : spd_jghj

 Target Server Type    : MySQL
 Target Server Version : 80028
 File Encoding         : 65001

 Date: 26/01/2026 16:43:01
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for pro_machine_sessions
-- ----------------------------
DROP TABLE IF EXISTS `pro_machine_sessions`;
CREATE TABLE `pro_machine_sessions`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `device_id` int NOT NULL COMMENT '登录设备ID',
  `team_id` int NOT NULL COMMENT '登录班组ID',
  `staff_ids` json NOT NULL COMMENT '当班员工ID列表 (如: [101, 102, 105])',
  `login_time` datetime NOT NULL COMMENT '上班/登录时间',
  `logout_time` datetime NULL DEFAULT NULL COMMENT '下班/登出时间 (NULL代表正在上班)',
  `duration_min` int NULL DEFAULT 0 COMMENT '上班时长(分钟)',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_device_active`(`device_id` ASC, `logout_time` ASC) USING BTREE,
  INDEX `idx_team_time`(`team_id` ASC, `login_time` ASC) USING BTREE,
  INDEX `idx_session_device_logout`(`device_id` ASC, `logout_time` ASC) USING BTREE,
  INDEX `idx_session_team_time`(`team_id` ASC, `login_time` ASC) USING BTREE,
  INDEX `idx_session_login`(`login_time` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 58 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【考勤表】设备登录与班次记录' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for pro_orders
-- ----------------------------
DROP TABLE IF EXISTS `pro_orders`;
CREATE TABLE `pro_orders`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '工单内部ID',
  `order_no` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '工单号 (ERP或MES下发的唯一编号)',
  `product_code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '产品型号/物料编码',
  `target_device_id` int NULL DEFAULT NULL COMMENT '计划生产设备ID (关联 sys_devices.id)',
  `plan_qty` int NOT NULL DEFAULT 0 COMMENT '【计划】计划生产总数',
  `actual_qty` int NULL DEFAULT 0 COMMENT '【实绩】实际生产总数 (OK数 + NG数)',
  `ok_qty` int NULL DEFAULT 0 COMMENT '【实绩】良品(OK)总数',
  `ng_qty` int NULL DEFAULT 0 COMMENT '【实绩】不良品(NG)总数',
  `status` tinyint NULL DEFAULT 0 COMMENT '工单状态: 0-待产, 1-生产中, 2-暂停, 3-已完工, 4-强制关闭',
  `start_time` datetime NULL DEFAULT NULL COMMENT '实际开工时间 (点击开始生产的时刻)',
  `end_time` datetime NULL DEFAULT NULL COMMENT '实际完工时间 (点击结束/完成的时刻)',
  `version` int NULL DEFAULT 0 COMMENT '乐观锁版本号 (防止多个协程同时更新导致计数错误)',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `used_seconds` int NULL DEFAULT 0 COMMENT '已使用秒数(累计)',
  `current_start_time` datetime NULL DEFAULT NULL COMMENT '当前开始时间(用于计算本次用时)',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `order_no`(`order_no` ASC) USING BTREE,
  INDEX `idx_status`(`status` ASC) USING BTREE COMMENT '用于快速查找正在生产中的工单',
  INDEX `idx_order`(`order_no` ASC) USING BTREE,
  INDEX `idx_device_status`(`target_device_id` ASC, `status` ASC) USING BTREE,
  INDEX `idx_order_device`(`target_device_id` ASC) USING BTREE,
  INDEX `idx_order_status`(`status` ASC) USING BTREE,
  INDEX `idx_order_created`(`created_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 13 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【核心表】生产工单主表 (记录计划与总进度)' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for pro_production_runs
-- ----------------------------
DROP TABLE IF EXISTS `pro_production_runs`;
CREATE TABLE `pro_production_runs`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '班次运行ID',
  `order_id` bigint NOT NULL COMMENT '关联工单ID (pro_orders.id)',
  `device_id` int NOT NULL DEFAULT 0 COMMENT '实际执行设备ID (关联 sys_devices.id)',
  `team_id` int NOT NULL COMMENT '执行班组ID (sys_teams.id)',
  `run_ok_qty` int NULL DEFAULT 0 COMMENT '本班次产出-良品数 (增量)',
  `run_ng_qty` int NULL DEFAULT 0 COMMENT '本班次产出-不良品数 (增量)',
  `start_time` datetime NOT NULL COMMENT '本班次开工时间',
  `end_time` datetime NULL DEFAULT NULL COMMENT '本班次结束/换班时间 (为空表示正在进行中)',
  `operator_ids` json NULL COMMENT '【关键】人员快照: 记录开工那一刻班组里的所有员工ID, 格式如 [101, 102, 105]',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注 (如: 设备故障暂停后重新开始)',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_run_device_time`(`device_id` ASC, `start_time` ASC) USING BTREE,
  INDEX `idx_run_team`(`team_id` ASC) USING BTREE,
  INDEX `idx_run_active`(`end_time` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 24 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【明细表】工单分班次执行记录 (记录具体谁干了多少)' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_alarm_records
-- ----------------------------
DROP TABLE IF EXISTS `sys_alarm_records`;
CREATE TABLE `sys_alarm_records`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `var_id` bigint NOT NULL COMMENT '变量ID',
  `var_name` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '快照: 变量名',
  `val` float NULL DEFAULT NULL COMMENT '数值报警:触发值; 系统报警:错误码',
  `alarm_type` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '类型: HH, H, L, LL, SYS(系统/设备故障)',
  `msg` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '报警内容',
  `start_time` datetime NOT NULL COMMENT '报警开始时间',
  `end_time` datetime NULL DEFAULT NULL COMMENT '报警恢复时间 (为空表示未恢复)',
  `ack_status` tinyint NULL DEFAULT 0 COMMENT '确认状态: 0-未确认',
  `limit_value` double NULL DEFAULT NULL COMMENT '被超过的阈值 (HH/H/L/LL对应的limit值)',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_time`(`start_time` ASC) USING BTREE,
  INDEX `idx_var`(`var_id` ASC) USING BTREE,
  INDEX `idx_alarm_records_limit_value`(`limit_value` ASC) USING BTREE,
  INDEX `idx_alarm_time`(`start_time` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 150201 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '报警历史记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_config_version
-- ----------------------------
DROP TABLE IF EXISTS `sys_config_version`;
CREATE TABLE `sys_config_version`  (
  `id` int NOT NULL,
  `version_code` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '版本号或UUID',
  `updated_at` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_data_history
-- ----------------------------
DROP TABLE IF EXISTS `sys_data_history`;
CREATE TABLE `sys_data_history`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `var_id` bigint NOT NULL,
  `val` float NULL DEFAULT NULL COMMENT '数值',
  `str_val` varchar(500) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '字符串值',
  `created_at` datetime(3) NOT NULL COMMENT '时间戳(精确到毫秒)',
  PRIMARY KEY (`id`, `created_at`) USING BTREE,
  INDEX `idx_var_time`(`var_id` ASC, `created_at` ASC) USING BTREE,
  INDEX `idx_history_perfect`(`var_id` ASC, `val` ASC, `created_at` ASC) USING BTREE,
  INDEX `idx_history_chart`(`var_id` ASC, `created_at` DESC) USING BTREE,
  INDEX `idx_stats_covering`(`var_id` ASC, `created_at` ASC, `val` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3444059 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '历史趋势数据表' ROW_FORMAT = Dynamic PARTITION BY RANGE (to_days(`created_at`))
PARTITIONS 4
(PARTITION `p_history` VALUES LESS THAN (739951) ENGINE = InnoDB MAX_ROWS = 0 MIN_ROWS = 0 ,
PARTITION `p202512` VALUES LESS THAN (739982) ENGINE = InnoDB MAX_ROWS = 0 MIN_ROWS = 0 ,
PARTITION `p202601` VALUES LESS THAN (740013) ENGINE = InnoDB MAX_ROWS = 0 MIN_ROWS = 0 ,
PARTITION `p_future` VALUES LESS THAN (MAXVALUE) ENGINE = InnoDB MAX_ROWS = 0 MIN_ROWS = 0 )
;

-- ----------------------------
-- Table structure for sys_device_status
-- ----------------------------
DROP TABLE IF EXISTS `sys_device_status`;
CREATE TABLE `sys_device_status`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `device_id` int NOT NULL COMMENT '设备ID',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '设备状态: 0-空闲, 1-运行, 2-故障',
  `start_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '状态开始时间',
  `end_time` timestamp NULL DEFAULT NULL COMMENT '状态结束时间(NULL表示当前状态)',
  `duration_min` int NULL DEFAULT 0 COMMENT '状态持续时长(分钟)',
  `extra_data` json NULL COMMENT '扩展数据(JSON格式,存储温度、湿度等变量)',
  `remark` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '备注',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_device_status`(`device_id` ASC, `start_time` ASC) USING BTREE,
  INDEX `idx_device_active`(`device_id` ASC, `end_time` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 94 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '设备状态历史记录表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_devices
-- ----------------------------
DROP TABLE IF EXISTS `sys_devices`;
CREATE TABLE `sys_devices`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `gateway_id` int NOT NULL COMMENT '所属网关ID',
  `device_code` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '设备编码 (英文唯一)',
  `device_name` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '设备显示名称',
  `identify_key` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT 'JSON根键名 (可选)',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_gw`(`gateway_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '逻辑设备表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_error_codes
-- ----------------------------
DROP TABLE IF EXISTS `sys_error_codes`;
CREATE TABLE `sys_error_codes`  (
  `error_code` int NOT NULL COMMENT '错误码',
  `error_msg` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '错误描述',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`error_code`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '系统错误代码表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_gateways
-- ----------------------------
DROP TABLE IF EXISTS `sys_gateways`;
CREATE TABLE `sys_gateways`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `gw_name` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '网关名称 (如: 1号车间采集器)',
  `status` tinyint NULL DEFAULT 1 COMMENT '状态: 1-启用, 0-禁用',
  `mqtt_broker` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '地址 (如: tcp://192.168.1.10:1883)',
  `mqtt_client_id` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '客户端ID (必须唯一)',
  `mqtt_user` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '用户名',
  `mqtt_pass` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '密码',
  `mqtt_topic` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '订阅主题 (如: factory/line1/#)',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = 'MQTT网关连接配置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_staff
-- ----------------------------
DROP TABLE IF EXISTS `sys_staff`;
CREATE TABLE `sys_staff`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '员工ID',
  `staff_code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '工号 (唯一标识，如: OP2025001)',
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '员工姓名',
  `current_team_id` int NULL DEFAULT NULL COMMENT '当前所属班组ID (注意: 这只是当前状态，人员流动历史请查询 sys_staff_history 表)',
  `is_active` tinyint NULL DEFAULT 1 COMMENT '在职状态: 1-在职, 0-离职 (离职后不要删除记录，改为0即可)',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '入职/录入时间',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `staff_code`(`staff_code` ASC) USING BTREE,
  INDEX `idx_team`(`current_team_id` ASC) USING BTREE,
  INDEX `idx_staff_team`(`current_team_id` ASC) USING BTREE,
  INDEX `idx_staff_active`(`is_active` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 15 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【基础表】员工人员名册' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_staff_history
-- ----------------------------
DROP TABLE IF EXISTS `sys_staff_history`;
CREATE TABLE `sys_staff_history`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `staff_id` int NOT NULL COMMENT '关联员工ID',
  `team_id` int NOT NULL COMMENT '发生变动的班组ID',
  `action_type` tinyint NOT NULL COMMENT '变动类型: 1-加入班组, 2-离开/调出班组',
  `happened_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '变动发生时间 (精确到秒)',
  `operator_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '操作人 (是谁在系统上点的调班)',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_staff_time`(`staff_id` ASC, `happened_at` ASC) USING BTREE COMMENT '用于查询某员工的调动轨迹',
  INDEX `idx_team_time`(`team_id` ASC, `happened_at` ASC) USING BTREE COMMENT '用于回溯某班组在某时刻的成员列表'
) ENGINE = InnoDB AUTO_INCREMENT = 66 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【日志表】人员班组调动历史记录' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_task_execution_logs
-- ----------------------------
DROP TABLE IF EXISTS `sys_task_execution_logs`;
CREATE TABLE `sys_task_execution_logs`  (
  `log_id` bigint NOT NULL AUTO_INCREMENT COMMENT '日志ID',
  `task_id` bigint NOT NULL COMMENT '任务ID',
  `task_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务名称',
  `execute_time` datetime NOT NULL COMMENT '执行时间',
  `success` tinyint NOT NULL COMMENT '是否成功: 1=成功, 0=失败',
  `error_msg` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '错误信息',
  `duration` int NULL DEFAULT NULL COMMENT '执行耗时(毫秒)',
  `result` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '执行结果',
  PRIMARY KEY (`log_id`) USING BTREE,
  INDEX `idx_task_time`(`task_id` ASC, `execute_time` ASC) USING BTREE,
  INDEX `idx_execute_time`(`execute_time` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 14945 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '任务执行日志表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_tasks
-- ----------------------------
DROP TABLE IF EXISTS `sys_tasks`;
CREATE TABLE `sys_tasks`  (
  `task_id` bigint NOT NULL AUTO_INCREMENT COMMENT '任务ID',
  `task_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '任务名称',
  `task_type` tinyint NOT NULL COMMENT '任务类型: 1=定时, 2=数据改变, 3=条件事件',
  `is_enabled` tinyint NOT NULL DEFAULT 1 COMMENT '是否启用: 1=启用, 0=禁用',
  `description` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '任务描述',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `cron_expr` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'Cron 表达式 (如: */5 * * * *)',
  `interval_sec` int NULL DEFAULT NULL COMMENT '简单间隔(秒), 优先使用 cron_expr',
  `last_run_time` datetime NULL DEFAULT NULL COMMENT '上次执行时间',
  `trigger_var_id` bigint NULL DEFAULT NULL COMMENT '触发的测点ID (关联 sys_variables.id)',
  `trigger_var_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '触发的测点名称',
  `change_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT 'ANY' COMMENT '变化类型: ANY, INCREASE, DECREASE, THRESHOLD, FALSE_TO_TRUE, TRUE_TO_FALSE',
  `change_threshold` float NULL DEFAULT NULL COMMENT '变化阈值 (用于 THRESHOLD 类型)',
  `condition_expr` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '条件表达式 (如: temp>50 AND pressure>100)',
  `action_type` tinyint NOT NULL COMMENT '动作类型: 1=HTTP请求, 2=MQTT发布, 3=数据库操作, 4=执行脚本, 5=写日志',
  `action_config` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '动作配置 (JSON格式)',
  PRIMARY KEY (`task_id`) USING BTREE,
  INDEX `idx_type_enabled`(`task_type` ASC, `is_enabled` ASC) USING BTREE,
  INDEX `idx_trigger_var`(`trigger_var_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 35 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '任务调度配置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_teams
-- ----------------------------
DROP TABLE IF EXISTS `sys_teams`;
CREATE TABLE `sys_teams`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '班组ID',
  `team_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '班组名称 (如: 激光焊早班A组)',
  `leader_name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '当前班长姓名 (冗余字段，方便显示)',
  `status` tinyint NULL DEFAULT 1 COMMENT '班组状态: 1-启用, 0-已解散/禁用',
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 5 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '【基础表】生产班组信息表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for sys_variables
-- ----------------------------
DROP TABLE IF EXISTS `sys_variables`;
CREATE TABLE `sys_variables`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `device_id` int NOT NULL COMMENT '所属设备ID',
  `var_name` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '变量名 (程序内部标识)',
  `display_name` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '显示名称 (如: 进水温度)',
  `data_type` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT 'FLOAT' COMMENT '数据类型: BOOL(布尔), INT16(短整型), INT32(整型), INT64(长整型/电能), FLOAT(浮点), DOUBLE(双精度), STRING(字符串)',
  `rw_mode` varchar(10) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT 'R' COMMENT '读写模式: R, W, RW',
  `unit` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '单位 (如: ℃)',
  `json_path` varchar(150) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT 'JSON提取路径',
  `scale_factor` float NULL DEFAULT 1 COMMENT '比例系数',
  `offset_val` float NULL DEFAULT 0 COMMENT '偏移量',
  `alarm_enable` tinyint NULL DEFAULT 1 COMMENT '报警开关',
  `limit_hh` float NULL DEFAULT NULL COMMENT '高高限 (HH)',
  `limit_h` float NULL DEFAULT NULL COMMENT '高限 (H)',
  `limit_l` float NULL DEFAULT NULL COMMENT '低限 (L)',
  `limit_ll` float NULL DEFAULT NULL COMMENT '低低限 (LL)',
  `deadband` float NULL DEFAULT 0 COMMENT '报警死区 (防止跳变)',
  `alarm_msg` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '自定义报警文案',
  `store_mode` tinyint NULL DEFAULT 1 COMMENT '存储模式: 0-不存, 1-变化, 2-定时, 3-混合',
  `store_cycle` int NULL DEFAULT 60 COMMENT '定时存储周期(秒)',
  `store_deadband` float NULL DEFAULT 0 COMMENT '存储死区(变化量超过此值才存)',
  `source_type` tinyint NULL DEFAULT 0 COMMENT '来源: 0-MQTT, 1-计算',
  `calc_rule` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '计算公式 (Go表达式)',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_device`(`device_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 118 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '变量点位配置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Procedure structure for auto_create_next_month_partition
-- ----------------------------
DROP PROCEDURE IF EXISTS `auto_create_next_month_partition`;
delimiter ;;
CREATE PROCEDURE `auto_create_next_month_partition`()
BEGIN
    DECLARE next_month_days INT;
    DECLARE partition_name VARCHAR(20);
    
    -- 计算下下个月的1号（作为切割点）
    SET next_month_days = TO_DAYS(DATE_ADD(NOW(), INTERVAL 2 MONTH));
    
    -- 生成分区名，例如 p202602
    SET partition_name = CONCAT('p', DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 1 MONTH), '%Y%m'));
    
    -- 如果不存在则创建
    IF NOT EXISTS (
        SELECT * FROM information_schema.partitions 
        WHERE TABLE_NAME = 'sys_data_history' 
        AND PARTITION_NAME = partition_name
    ) THEN
        SET @sql = CONCAT(
            'ALTER TABLE sys_data_history REORGANIZE PARTITION p_future INTO (',
            'PARTITION ', partition_name, ' VALUES LESS THAN (', next_month_days, '),',
            'PARTITION p_future VALUES LESS THAN MAXVALUE',
            ')'
        );
        
        PREPARE stmt FROM @sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
    END IF;
END
;;
delimiter ;

-- ----------------------------
-- Event structure for evt_auto_add_partition
-- ----------------------------
DROP EVENT IF EXISTS `evt_auto_add_partition`;
delimiter ;;
CREATE EVENT `evt_auto_add_partition`
ON SCHEDULE
EVERY '1' MONTH STARTS '2025-12-12 13:42:03'
DO CALL auto_create_next_month_partition()
;;
delimiter ;

SET FOREIGN_KEY_CHECKS = 1;
