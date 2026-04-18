-- CN: sys_variables 采集层防抖与启动首帧快照配置列，生产库可手工执行；应用启动也会幂等补列。
-- EN: sys_variables acquisition debounce and startup snapshot columns; safe for manual production execution, and app startup also adds them idempotently.
-- JP: sys_variables の収集デバウンス/起動初回スナップショット列。本番で手動実行可能で、アプリ起動時にも冪等追加されます。

SET @schema_name = DATABASE();

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE sys_variables ADD COLUMN suspicious_value DOUBLE NULL DEFAULT NULL COMMENT ''可疑抖动值，NULL=关闭采集防抖''',
    'SELECT ''sys_variables.suspicious_value already exists'''
  )
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'sys_variables'
    AND COLUMN_NAME = 'suspicious_value'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE sys_variables ADD COLUMN debounce_threshold DOUBLE NULL DEFAULT 5 COMMENT ''采集防抖起滤阈值，LastValidValue 大于该值才拦截可疑值''',
    'SELECT ''sys_variables.debounce_threshold already exists'''
  )
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'sys_variables'
    AND COLUMN_NAME = 'debounce_threshold'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @sql = (
  SELECT IF(
    COUNT(*) = 0,
    'ALTER TABLE sys_variables ADD COLUMN startup_snapshot_enable TINYINT NULL DEFAULT NULL COMMENT ''启动首帧快照开关，NULL=兼容旧行为，1=开启，0=关闭''',
    'SELECT ''sys_variables.startup_snapshot_enable already exists'''
  )
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 'sys_variables'
    AND COLUMN_NAME = 'startup_snapshot_enable'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
