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

## 改动后至少检查

1. desktop/main.go 启动链路是否仍能初始化数据库、通道、任务调度与网关。
2. desktop/app.go 中的方法名是否与前端调用保持一致。
3. 若涉及 SQL、配置或热重载，至少保留一条可人工复查的验证路径。

## 快照与历史数据的排序设计规范

### 班次快照排序（GetShiftSnapshots）
- ORDER BY 固定为：`snapshot_date DESC, (SELECT sort_order FROM sys_shifts WHERE id = shift_id) DESC, device_id ASC`
- 同一天内的班次顺序由 `sys_shifts.sort_order` 决定，值越大排越前（夜班=2 > 中班=1 > 早班=0）。
- **不要**改回 `shift_id ASC`——shift_id 是插入顺序，和实际排班时序无关。

### sort_order 字段规范
- `sys_shifts.sort_order`：存在，用于班次内排序，已有数据（0=早班, 1=中班, 2=夜班）。
- `sys_teams`：**无** sort_order 字段，班组排序以 id 升序为准，不要在前端擅自 reverse。
- 如后续需要为 sys_teams 加排序，必须先在 DB 模型（`models/mes_models.go` 的 `SysTeam`）和建表 SQL 中同步增加字段，不允许前端靠 reverse() 变通。
