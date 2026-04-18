# [后端规则] Go + Wails + MQTT + MySQL SCADA

## 长期事实

1. 这是桌面项目，桌面窗口启动时同时初始化后端核心。
2. Wails 绑定方法集中在 desktop/app.go，前端通过 window.go.main.App 和 wailsjs 调用。
3. MQTT 是现场实时数据主入口，数据库主要负责配置、历史、报警、任务、工单、人员班组等持久化。

## 修改后端时的默认规则

1. 不要先入为主新增 HTTP API 作为主链路。
2. 不要把数据库写入、HTTP 调用、AI 调用这类慢 IO 塞回 MQTT 回调或 LogicWorker 热路径。
3. 保持 TagManager、Worker 池、TaskScheduler、EventProcessor、GatewayManager 的边界清晰。
4. 多网关配置以数据库表为准，不要把 config.env 中的旧 MQTT 配置重新当成主配置中心。
5. 产量计数 SQL 必须使用 last_nonzero_val 三档防回跳规则（val=prev→0, val=0且prev≠0→last_nonzero, else→val-prev），不允许退回双分支 CASE。涉及函数：GetHourlyProductionAccurate、GetMonthlyProductionAccurate、GetShiftWindowProduction、GetHourlyOEE、GetHourlyOEEWithSQL。
6. MySQL DSN 必须包含连接超时参数（`timeout=5s&readTimeout=30s&writeTimeout=30s`），防止 MySQL 不可达时卡死启动流程。
7. 产量累计计数器的假回落必须优先在采集层处理：`models.Tag.ApplyNumericSample()` 在 `LogicWorker` 更新内存前执行，`suspicious_value` 为空时关闭防抖，启动首帧可疑值必须先用历史最近非零值判定，不能直接被 `InitOnly` 写入 `sys_data_history`。详细规则见 `docs/采集层防抖与启动快照说明.md`。
8. 工单数量口径固定为：`actual_qty` = 产量累计计数器增量（总产量），`ng_qty` = NG+1/NG-1 脉冲净值，`ok_qty` = `actual_qty - ng_qty`。不要把产量累计计数器增量当作 OK 写入；历史修复时先修 `actual_qty`，NG 已确认正确时不覆盖 `ng_qty`，最后再按公式修 `ok_qty`。

## 任务调度规则（TaskScheduler）

### 条件任务（TaskType=3）边沿触发规则
- **条件任务必须使用边沿触发**：只在条件从 `false→true` 跳变时执行一次，条件持续为真时不重复触发。
- 实现方式：`TaskScheduler.conditionLastResult map[int64]bool` 记录每个任务的上次求值结果；
  `checkScheduledTasks` 中判断 `result == true && prevResult == false` 才调用 `triggerTask`。
- **禁止**改为"条件为真即触发"的电平模式——会导致日志表（`sys_task_execution_logs`）无限写入、磁盘撑满、应用卡死。
- 示例故障：设备离线检测任务（interval_sec=5，条件=`GetVarQuality(x)==0`），设备持续离线时每5秒写一条，数小时内可积累百万行。

### 日志表维护
- `sys_task_execution_logs` 是高频写入表，生产环境需定期清理旧数据（建议保留7天，超出即删）。
- 紧急止血 SQL：`DELETE FROM sys_task_execution_logs WHERE execute_time < NOW() - INTERVAL 7 DAY; OPTIMIZE TABLE sys_task_execution_logs;`

## 改动后至少检查

1. desktop/main.go 启动链路是否仍能初始化数据库、通道、任务调度与网关。
2. desktop/app.go 中的方法名是否与前端调用保持一致。
3. 若涉及 SQL、配置或热重载，至少保留一条可人工复查的验证路径。
4. 若涉及任务调度逻辑，确认条件任务仍为边沿触发，不会退化为电平触发。
5. 若涉及采集层产量防抖，确认 `10,0,12`、`10,0,1`、`1,0,1`、启动历史 `77` 首帧 `0` 下一帧 `77` 等场景仍被测试覆盖。
6. 若涉及工单历史修复 SQL，必须默认使用事务预览和 `ROLLBACK` 收尾，人工确认后再改 `COMMIT`；复查至少包含 `ok_qty + ng_qty = actual_qty` 与 `ng_qty >= 0`。

## 快照与历史数据的排序设计规范

### 班次快照排序（GetShiftSnapshots）
- ORDER BY 固定为：`snapshot_date DESC, (SELECT sort_order FROM sys_shifts WHERE id = shift_id) DESC, device_id ASC`
- 同一天内的班次顺序由 `sys_shifts.sort_order` 决定，值越大排越前（夜班=2 > 中班=1 > 早班=0）。
- **不要**改回 `shift_id ASC`——shift_id 是插入顺序，和实际排班时序无关。

### sort_order 字段规范
- `sys_shifts.sort_order`：存在，用于班次内排序，已有数据（0=早班, 1=中班, 2=夜班）。
- `sys_teams`：**无** sort_order 字段，班组排序以 id 升序为准，不要在前端擅自 reverse。
- 如后续需要为 sys_teams 加排序，必须先在 DB 模型（`models/mes_models.go` 的 `SysTeam`）和建表 SQL 中同步增加字段，不允许前端靠 reverse() 变通。
