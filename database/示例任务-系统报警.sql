-- 设备1系统错误信息 - 记录系统报警
INSERT INTO `sys_tasks` (
    `task_name`, 
    `task_type`, 
    `is_enabled`, 
    `description`, 
    `trigger_var_id`, 
    `trigger_var_name`, 
    `change_type`, 
    `action_type`, 
    `action_config`
) VALUES (
    '设备1系统错误信息-记录系统报警',
    2,
    1,
    '当设备1系统错误信息变化时，查询错误码表并记录系统报警',
    70,
    '设备1系统错误信息',
    'ANY',
    3,
    '{
  "operation": "log_system_alarm",
  "op_params": {
    "var_id": 70,
    "var_name": "设备1",
    "error_code": "{{new_value}}"
  }
}'
);

-- 设备2系统错误信息 - 记录系统报警
INSERT INTO `sys_tasks` (
    `task_name`, 
    `task_type`, 
    `is_enabled`, 
    `description`, 
    `trigger_var_id`, 
    `trigger_var_name`, 
    `change_type`, 
    `action_type`, 
    `action_config`
) VALUES (
    '设备2系统错误信息-记录系统报警',
    2,
    1,
    '当设备2系统错误信息变化时，查询错误码表并记录系统报警',
    69,
    '设备2系统错误信息',
    'ANY',
    3,
    '{
  "operation": "log_system_alarm",
  "op_params": {
    "var_id": 69,
    "var_name": "设备2",
    "error_code": "{{new_value}}"
  }
}'
);







