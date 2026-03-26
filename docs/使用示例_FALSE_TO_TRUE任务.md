# 数据改变任务使用示例：设备状态监控

## 场景说明

你有一个设备运行状态变量 `device_running`（布尔类型），希望实现：
- 当状态从 `false` 变为 `true` 时：在 `sys_device_status` 表插入设备启动记录
- 当状态从 `true` 变为 `false` 时：更新 `sys_device_status` 表，结束当前运行记录并计算持续时长

## 步骤1: 配置变量（sys_variables 表）

首先在 `sys_variables` 表中创建一个布尔类型的变量：

```sql
INSERT INTO `sys_variables` (
  `device_id`, 
  `var_name`, 
  `display_name`, 
  `data_type`, 
  `json_path`,
  `store_mode`
) VALUES (
  1,                      -- 设备ID
  'device_running',       -- 变量名（程序内部标识）
  '设备运行状态',         -- 显示名称
  'BOOL',                 -- 数据类型：布尔
  'status.running',       -- JSON提取路径
  1                       -- 存储模式：变化存储
);
-- 假设这条记录的 ID 为 100
```

## 步骤2: 创建数据改变任务（sys_tasks 表）

### 任务1: 设备启动时插入记录

```sql
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
  '设备启动状态更新',                     -- 任务名称
  2,                                      -- 任务类型：2=数据改变任务
  1,                                      -- 启用
  '当设备运行状态从false变为true时，插入设备启动记录',
  100,                                    -- 触发变量ID（对应上面的变量）
  'device_running',                       -- 触发变量名
  'FALSE_TO_TRUE',                        -- 变化类型：false→true
  3,                                      -- 动作类型：3=数据库操作
  '{
    "sql": "INSERT INTO sys_device_status (device_id, status, start_time) VALUES (?, 1, NOW())",
    "params": {"device_id": 1}
  }'
);
```

### 任务2: 设备停止时更新记录

```sql
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
  '设备停止状态更新',                     -- 任务名称
  2,                                      -- 任务类型：2=数据改变任务
  1,                                      -- 启用
  '当设备运行状态从true变为false时，结束设备运行记录',
  100,                                    -- 触发变量ID
  'device_running',                       -- 触发变量名
  'TRUE_TO_FALSE',                        -- 变化类型：true→false
  3,                                      -- 动作类型：3=数据库操作
  '{
    "sql": "UPDATE sys_device_status SET end_time=NOW(), duration_min=TIMESTAMPDIFF(MINUTE,start_time,NOW()) WHERE device_id=? AND end_time IS NULL",
    "params": {"device_id": 1}
  }'
);
```

## 步骤3: 重启系统或热重载配置

任务配置会在系统启动时自动加载，也可以通过修改 `sys_config` 表的 `config_version` 触发热重载：

```sql
UPDATE `sys_config` SET `config_value` = CONCAT('v', UNIX_TIMESTAMP()) WHERE `config_key` = 'config_version';
```

## 步骤4: 测试

向 MQTT Broker 发送消息：

```json
// 设备启动（false → true）
{
  "status": {
    "running": true
  }
}

// 设备停止（true → false）
{
  "status": {
    "running": false
  }
}
```

系统会自动：
1. LogicWorker Pass 1: 更新内存中的 `device_running` 变量值
2. LogicWorker Pass 2: 检测到值从 `false→true`，触发 TaskScheduler
3. TaskScheduler: 匹配到 `FALSE_TO_TRUE` 类型的任务，发送到 EventChan
4. EventProcessor: 执行数据库操作，插入/更新 `sys_device_status` 表

## 查看执行日志

```sql
-- 查看任务执行日志
SELECT * FROM sys_task_execution_logs 
WHERE task_name LIKE '%设备%状态%' 
ORDER BY execute_time DESC 
LIMIT 10;

-- 查看设备状态记录
SELECT * FROM sys_device_status 
WHERE device_id = 1 
ORDER BY start_time DESC 
LIMIT 10;
```

## 支持的其他变化类型

除了 `FALSE_TO_TRUE` 和 `TRUE_TO_FALSE`，系统还支持：

- `ANY`: 任意变化
- `INCREASE`: 数值增加
- `DECREASE`: 数值减少
- `THRESHOLD`: 变化量超过指定阈值（配合 `change_threshold` 字段）

## 支持的动作类型

- `1`: HTTP 请求（发送到外部 API）
- `2`: MQTT 发布（发布消息到 MQTT Broker）
- `3`: 数据库操作（执行 SQL 语句）
- `4`: 脚本执行（调用 bash/python/powershell 脚本）
- `5`: 日志写入（写入日志文件）

## 模板变量

`action_config` 中的 SQL 或其他配置支持模板变量替换：

- `{{var_id}}`: 变量ID
- `{{old_value}}`: 变化前的值
- `{{new_value}}`: 变化后的值
- `{{change}}`: 变化量（new_value - old_value）
- `{{trigger_time}}`: 触发时间

示例：
```json
{
  "sql": "INSERT INTO logs (var_id, old_val, new_val, change_time) VALUES ({{var_id}}, {{old_value}}, {{new_value}}, '{{trigger_time}}')"
}
```

## 注意事项

1. **变量数据类型**：布尔类型的变量在内存中存储为 `float64`，`false=0.0`, `true=1.0`
2. **触发频率**：每次值变化只触发一次，避免重复触发
3. **执行顺序**：任务在 LogicWorker Pass 2 中执行，确保所有变量都已更新到最新值
4. **并发安全**：TaskScheduler 使用 RWMutex 保护，EventProcessor 池化处理，不会阻塞主流程
5. **失败处理**：执行失败会记录到 `sys_task_execution_logs` 表，不影响其他任务

## 架构优势

✅ **双循环一致性**：Pass 1 建立快照，Pass 2 执行任务，脚本读取的变量是同一时刻的值  
✅ **非阻塞设计**：任务通过 Channel 异步处理，不阻塞 MQTT 消息接收  
✅ **可扩展性**：支持多种触发类型和动作类型，轻松扩展新功能  
✅ **可观测性**：所有任务执行都有日志记录，方便调试和监控

