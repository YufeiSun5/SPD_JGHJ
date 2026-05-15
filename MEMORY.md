# [开发进度记忆库]
# - 项目类型: Go + Wails + Vue 3 + Naive UI 的桌面 SCADA/IIoT 系统，不是传统前后端 HTTP API 分离项目。
# - 当前调用模型: 前端主要通过 window.go.main.App 和 wailsjs 直接调用 Go 绑定方法；桌面启动时会一并初始化后端内核。
# - 当前主入口: 现场实时数据主链路是 MQTT，多网关配置以数据库表为准；config.env 中 MQTT 配置主要是兼容旧代码。
# - 当前数据库基线: config.env 默认 localhost:3306 / spd_jghj，运行时依赖 MySQL 驱动；代码里仍保留部分通用/历史数据库命名。
# - 当前前端结构: desktop/frontend 使用 Vue 3 + Vite + Naive UI，Hash 路由；核心页面覆盖驾驶舱、生产、质量、报警、人员、历史、设备、任务、系统设置、AI 助手。
# - 当前后端结构: desktop/main.go 启动时初始化数据库、TagManager、Worker 池、TaskScheduler、GatewayManager、ConfigReloader；核心业务绑定集中在 desktop/app.go。
# - 当前 AI 状态: AI 问答是旁路能力，desktop/ai_client.go 通过本机 http://127.0.0.1:8006 调用 FastAPI 服务，不属于主 SCADA 控制链路。
# - 当前长期约束: 不要先入为主补 HTTP API；不要破坏 MQTT -> 内存 -> Worker/Task/Event 的异步并发模型；业务核心位置要求中英日三语说明。
# - 当前 AI 文档结构: 根目录 docs、instructions、skills 已按当前 SCADA 项目重建，供编辑器/AI 读取；旧的 Spring/Taro 版本已移入作废目录。
# - 当前目录整理: 根目录暴露的业务 Markdown 已收口到 docs，根目录暴露的 SQL 脚本已收口到 sql；AGENTS.md、MEMORY.md、README.md 继续保留在根目录。
# - 当前文档状态: README 仍带有旧的 Gin API 项目表述，后续如继续整理文档，应优先按现有 Wails 桌面架构修正。
#
# ── Wails 绑定层薄壳化第一轮（2026-05-15）────────────────────────────────────
# - 已创建 backend/service 后端 service 包，第一轮将一批原本写在 desktop/app.go 的业务逻辑抽出：
#   ① runtime_service.go：实时点位 DTO、GetRealtimeData、GetAllTags、GetSystemMonitor；
#   ② order_service.go：工单创建/更新/删除、StartProductionSmart 状态流转与开工编排；
#   ③ history_variable_service.go：历史查询、变量配置 CRUD、批量更新/删除、配置版本热重载触发；
#   ④ device_status_service.go：设备状态汇总、24h/筛选历史、班次人员拼接、状态统计；
#   ⑤ alarm_task_service.go：报警列表/确认/统计、变量筛选项、任务 CRUD/手动触发、网关列表。
# - desktop/app.go 中对应 Wails 方法已改为薄壳委托，前端方法名与签名保持不变；新增类型通过 type alias 继续兼容 Wails 绑定。
# - 本轮未抽离 OEE/逻辑日班次窗口、班次配置保存、快照重建、AI stream 等高耦合路径；后续若继续薄壳化，应优先抽 Shift/OEE service，
#   但必须保持 last_nonzero_val 防回跳、逻辑日 calendar_day_offset、快照排序等既有口径。
# - 验证：go build ./...、go test ./... 通过。
#
# ── Wails 班次绑定薄壳化（2026-05-15）──────────────────────────────────────
# - 新增 backend/service/shift_service.go，承接原 desktop/app_shifts.go 中的班次重业务：
#   ① ShiftBreak / ShiftConfig / ShiftScheduleConfig / LogicalDayShift DTO；
#   ② 时间安排组/班次/休息段 upsert、级联删除、孤立脏数据清理；
#   ③ 设备 schedule_id 绑定、设备 cycle_time 读写；
#   ④ 当前班次、逻辑日班次、默认班次窗口计算。
# - desktop/app_shifts.go 已改为 Wails 薄壳：只保留类型别名和 App 方法委托，前端方法名保持不变。
# - desktop/app_snapshot.go 中原 modelToShiftConfig 调用改为 service.ModelToShiftConfig，避免快照继续依赖 Wails 文件里的转换 helper。
# - 保持既有逻辑日规则：sort_order 最小活动班次作为逻辑日边界；开始时间早于边界的班次使用 CalendarDayOffset=1。
# - 验证：go test ./...、go build ./...、desktop/frontend npm run build、desktop wails build -skipbindings 通过。
#   完整 wails build 的绑定生成阶段因已有桌面实例占用 127.0.0.1:47391 被单实例保护拦截，非本次代码编译错误。
#
# ── Wails 非核心薄壳化第二轮（2026-05-15）──────────────────────────────────
# - 新增并迁移一批明显不该留在 Wails 壳里的低/中风险业务：
#   ① error_code_service.go：错误码同步与查询；
#   ② statistics_service.go：基础统计透传、逻辑日小时产量、月度产量/质量；
#   ③ staff_session_service.go：设备列表、员工/班组 CRUD、人员调动、设备上下岗、session 查询；
#   ④ system_config_service.go：用户目录 user_config.json 的系统配置、每日工时、休息时间读写；
#   ⑤ energy_service.go：日电能差值计算、设备实时功率/日电能汇总。
# - desktop/app.go 对应 Wails 方法改为委托 service，方法名和前端调用保持不变；Startup、main 初始化、Wails runtime 事件桥仍留在 desktop。
# - OEE 与班次快照仍暂留 desktop：这两块互相依赖逻辑日窗口、设备 CT、last_nonzero_val 和快照排序，后续应作为独立高风险步骤迁移。
# - 验证：go test ./...、go build ./...、desktop/frontend npm run build、desktop wails build -skipbindings 通过。
#   完整 wails build 仍因已有桌面实例占用 127.0.0.1:47391 被单实例保护拦截，非本次代码编译错误。
#
# ── OEE 设备级 CT 口径统一（2026-04-18）────────────────────────────────────
# - 甲方要求两台设备可设置不同 CT；OEE 计算链路已统一按设备读取 sys_devices.cycle_time。
# - Wails 绑定 GetHourlyOEE 已改为无参接口：前端不再传 DeviceOEEConfig，后端统一从 DB 读取设备 schedule_id + cycle_time，
#   按 schedule_id 分组计算 OEE；设备 CT 为空时才回退系统默认 CT（user_config.production_coefficient）。
# - 旧的单点节拍绑定 GetProductionCoefficient / SetProductionCoefficient / GetProductionCoefficientFromEnv 已移除；
#   系统默认 CT 只通过 GetSystemConfig / SetSystemConfig 管理，作为未配置设备 CT 的兜底值。
# - DebugOEEByShift 改为与驾驶舱一致：使用设备级 CT、只展示已到达班次、当前班次按 NOW() 截断计划时间，不再 full-day 强行展开。
# - OEEDebug.vue 顶部汇总改为当前/最近已到达班次，且保留两台设备各自合计行和生效 CT，避免旧逻辑只取第一条合计行漏掉设备。
# - Quality.vue 工单性能稼动率、详情理论时间、CSV 理论时间已按工单 target_device_id 的设备 CT 计算；无设备 CT 时用系统默认 CT。
# - sql/spd_jghj.sql 的 sys_devices 基线补齐 schedule_id 和 cycle_time，避免新环境只导 SQL 时缺列。
#
# ── 采集层产量防抖与启动快照配置（2026-04-16）──────────────────────────────
# - sys_variables 新增可选配置列：suspicious_value（NULL=关闭采集防抖）、debounce_threshold（默认5）、
#   startup_snapshot_enable（NULL=兼容旧首帧行为，1=开启，0=关闭）。
# - 应用启动时 database.EnsureVariableAcquisitionConfigColumns() 会幂等补齐这些列；老表读写路径仍做列存在性判断，
#   避免 Unknown column 影响基础点位创建/更新。
# - models.Tag.ApplyNumericSample() 在 LogicWorker 更新内存前执行观察窗清洗：
#   LVV > threshold 且收到 suspicious_value 时先缓冲不入库、不触发任务、不更新 CurrentValue；
#   下一个有效值 >= LVV 时丢弃可疑值，下一个有效值 < LVV 时补发一个可疑值作为真实复位基准。
#   观察窗不超时，长期停在可疑值时不会落库，直到下一次非可疑值到来。
# - 低计数值保护：LVV <= threshold 时可疑值直接放行，保留 1,0,1 / 5,0,3 这类真实小数值复位。
# - 启动首帧可疑值修正：开启防抖的点位会在加载时预热 sys_data_history 最近非零值；
#   若首帧就是 suspicious_value 且历史 LVV > threshold，先挂起不执行 InitOnly 快照。
#   下一帧 >= 历史 LVV 时丢弃首帧可疑值并把下一帧作为启动快照；下一帧 < 历史 LVV 时补发可疑值再处理新值。
# - DeviceConfig.vue 已增加可疑抖动值、起滤阈值、启动快照三项配置；页面继续通过 Wails App 方法读写。
# - 验证：go test ./...、go build ./...、PowerShell 7 环境下 npm run build 已通过。
#
# ── 工单数量修复口径确认（2026-04-16）──────────────────────────────────────
# - 工单产量口径已确认：产量累计点位增量 = actual_qty（总产量），不是 ok_qty。
# - NG 数量来自 NG+1 / NG-1 脉冲净值；现场确认现有 NG 与 NG-1 统计正确时，历史修复 SQL 不应覆盖 ng_qty。
# - OK 数量公式固定为 ok_qty = actual_qty - ng_qty；修复顺序应先按累计产量点位和工单/运行时间窗口重算 actual_qty，
#   再按公式修 ok_qty，避免把总产量当 OK 或重复扣 NG。
# - 历史修复 SQL 必须默认 START TRANSACTION + 预览差异 + ROLLBACK；确认无误后再把最后一行改为 COMMIT。
#   默认先处理已完工工单；生产中工单因结束时间使用 NOW()，不建议混在第一轮批量修复里。
#
# ── 前端 UI 规范（已落地，2026-04-13）──────────────────────────────────────
# - 弹窗组件统一样式: ShiftModal / ActiveDevicesModal / SessionHistoryModal / StaffModal / TeamModal
#   五个组件的 CSS 已统一为 rgba(30,40,60,0.95) + blur(20px) + border-radius:12px，头部无渐变，
#   关闭按钮无方块背景，与 TaskManagement.vue 的 dialog 样式一致。
# - 禁止原生对话框: 项目内禁止使用 alert() / confirm() / prompt()。
#   ① 操作结果提示 → window.$message.success/error/warning（由 MessageSetup.vue 挂载到全局）
#   ② 二次确认操作 → src/components/ConfirmDialog.vue（支持details/warnings，参考 Production.vue 用法）
#   以上规范已写入 instructions/frontend-wails-vue.instructions.md。
# - Staff.vue 已完成全面替换: 所有 alert() → window.$message，所有 confirm() → ConfirmDialog 组件；
#   结束班次确认框显示设备/班组/人员/已工作时长详情，与生产计划页面风格一致。
#
# ── 班次快照与班次配置体系（已落地，2026-04-13）──────────────────────────────
# - 自动换班快照机制:
#   desktop/app_snapshot.go — StartSnapshotTicker（定时器）→ checkAndGenerateSnapshots →
#   generateOneSnapshot（对每个 schedule × shift × device 聚合快照写入 pro_shift_snapshots）。
#   对外暴露 Wails 绑定：GetShiftSnapshots（筛选查询）、RegenerateShiftSnapshot（手动重建单条）、
#   BatchRegenerateShiftSnapshots（批量回填，支持 dates×shiftIDs×deviceIDs，空列表=全部）。
#
# - 快照产量口径统一（2026-04-13 落地）:
#   generateOneSnapshot 的产量统计已替换为 database.GetShiftWindowProduction，
#   与 GetHourlyProductionAccurate / GetMonthlyProductionAccurate 共享同一套三档
#   last_nonzero_val 防复位回跳规则，解决设备复位后回跳被重计的问题（如 73→0→73）。
#   NG 加减按钮本轮未改（保留 val=1 脉冲计数，不做边沿去重）。
#   历史快照可通过 BatchRegenerateShiftSnapshots 按指定日期批量回填。
#
# - 班次配置数据模型: models/mes_models.go 中 SysShift / SysShiftSchedule；
#   desktop/app_shifts.go 中封装了 ShiftConfig / ShiftScheduleConfig 视图结构，
#   对外绑定：GetShiftSchedules、GetShifts、SaveShiftSchedules、GetCurrentShift、
#   GetShiftsForLogicalDay、GetDefaultShiftWindow、SetDeviceSchedule、GetDeviceCycleTime、SetDeviceCycleTime。
#
# - 系统设置页面（SystemSettings.vue）:
#   包含「工作安排」二层配置（时间安排 → 班次 → 休息段）+ 设备关联到时间安排；
#   逻辑日工时汇总区域（只读，由班次配置自动计算）；各班次时长/休息分钟显示。
#
# - 班次生产追溯页面（ShiftReport.vue）:
#   展示 pro_shift_snapshots 数据；支持时间范围/设备/班组/人员多维筛选（筛选项变化自动查询）；
#   班次色彩徽章、人员排名（金银铜铁土），CSV 导出；详情弹窗。
#
# - OEE 调试页面（OEEDebug.vue）:
#   新增调试页，展示今日 OEE 汇总、班次快照列表、逐小时明细，供排查 OEE 计算问题。
#
# - 驾驶舱（Cockpit.vue）调整:
#   OEE 卡片双栏展示两台设备；OEE 趋势图 x 轴由 GetShiftsForLogicalDay 接口提供逻辑日班次构建；
#   今日小时产量、月度产量等数据已接入。
#
# - 快照相关文档: docs/refactor_snapshot_staff/ 目录下有完整改造方案总览和分阶段说明文档。
#
# ── 数据正确性与 OEE 修复（2026-04-14）──────────────────────────────────────
# - MySQL JSON 空串修复:
#   desktop/app_snapshot.go — generateOneSnapshot 中 snap 初始化时，
#   StaffSnapshot 和 SessionsSnapshot 默认值从空字符串改为 "[]"，
#   修复 MySQL Error 3140（JSON 列拒绝空字符串）。
#
# - 质量率数据源统一:
#   desktop/app.go — GetAllShiftsQualitySummary() 完全重写，
#   从 GetShiftQualityByRun（pro_production_runs 表）改为逐设备调用 database.GetShiftWindowProduction（sys_data_history 表），
#   与快照产量口径一致。新增 "math" import 用于 Round。
#
# - OEE 封顶移除:
#   desktop/app_snapshot.go — computeOEE() 移除了 AvailPct > 100 和 PerfPct > 100 的截断逻辑，
#   三项 OEE 指标不再硬性封顶到 100%，允许如实反映原始数据。
#
# - OEE 防回跳统一（核心修复）:
#   database/statistics_db.go 中两个函数补齐了 last_nonzero_val 三档 CASE 防回跳逻辑：
#   ① GetHourlyOEE (~L880): ProductionRaw CTE 新增 last_nonzero_val 子查询列，
#      ProductionStats CTE 从 2-branch CASE 改为 3-branch（val=prev → 0, val=0且prev≠0 → last_nonzero, else → val-prev）。
#   ② GetHourlyOEEWithSQL (~L1185): 同上修复。
#   至此，所有产量计数路径（GetHourlyProductionAccurate、GetMonthlyProductionAccurate、
#   GetShiftWindowProduction、GetHourlyOEE、GetHourlyOEEWithSQL）共享同一套防回跳规则。
#   不需要修复的函数：GetHourlyProduction（pro_production_runs 表）、GetHourlyProductionPulse（脉冲计数）、
#   质量相关函数（pro_orders/pro_production_runs）。
#
# - sys_device_status 数据清理指导:
#   提供了完整 SQL 脚本供用户在 MySQL 客户端执行：
#   ① 诊断 run_sec > plan_sec 的异常快照记录
#   ② 清理 sys_device_status 中 end_time IS NULL 的悬挂记录
#   ③ 删除重叠/无效状态区间
#   ④ 删除异常快照后通过 BatchRegenerateShiftSnapshots 回填
#
# ── 快照逻辑日临界修复（2026-04-15）──────────────────────────────────────────
# - 问题: checkAndGenerateSnapshots 直接用迭代的日历日作为 snapshot_date，
#   不知道"逻辑日临界时刻"概念。三班(0:00-7:40)在日历次日运行，但应属于前一个逻辑日。
#   结果：启动后立即为当天凌晨的三班生成了 snapshot_date="今天" 的错误快照。
# - 修复: desktop/app_snapshot.go — checkAndGenerateSnapshots：
#   ① 检查窗口从2天扩为3天（day-2, day-1, today），防止逻辑日+日历偏移后漏检。
#   ② 每个 sched 循环内，找 sort_order 最小的活跃班次的 start_min 作为 logicalBoundaryMin。
#   ③ 对每个班次：若 shiftStartMin < logicalBoundaryMin → calendarDayOffset=1（日历次日运行）。
#   ④ calendarBase = logicalDate + calendarDayOffset（用于 resolveShiftDatetime 生成正确时间戳）。
#   ⑤ dateStr 始终用 logicalDate（逻辑日），不再用 calendarBase。
# - 已知约束: sort_order 是逻辑日语义的唯一依据，sort_order 最小 = 逻辑日第一班 = 临界时刻来源。
# - 验证: go build ./... 通过；可用 BatchRegenerateShiftSnapshots 清除旧错误记录后回填。
#
# ── 条件任务日志爆炸修复（2026-04-15/16）── sys_task_execution_logs 暴增 227 万行 ──────────────
# - 问题: task_type=3 条件任务（如设备离线/上线检测）配置了 interval_sec=5，
#   每5秒评估一次条件；条件为真时无边沿检测，每个周期都触发一次写入。
#   设备持续离线时，离线检测任务每5秒写一条日志 → 227万行 → 磁盘100% → MySQL I/O卡死 →
#   应用启动时 InitDatabase 挂起，Wails 窗口永远无法打开（进程内存仅0.5MB）。
# - 止血操作: 在数据库 UPDATE sys_tasks SET is_enabled=0 WHERE task_id IN (22,23,28,29);
#   再 DELETE + OPTIMIZE TABLE sys_task_execution_logs 清理历史日志。
# - 根治修复: workers/task_scheduler.go：
#   ① TaskScheduler 结构体新增 conditionLastResult map[int64]bool 字段。
#   ② InitTaskScheduler 初始化该 map。
#   ③ checkScheduledTasks 内条件任务改为边沿触发：
#      记录 prevResult = conditionLastResult[taskID]；
#      只有 result==true && prevResult==false 时才调用 triggerTask；
#      每次都更新 conditionLastResult[taskID] = result。
#   效果：条件持续为真时只触发一次（false→true跳变），不再重复写日志。
# - 附加修复: database/database.go — MySQL DSN 加入 timeout=5s&readTimeout=30s&writeTimeout=30s，
#   防止 MySQL 不可达时 TCP 握手卡死启动（原来无超时，最长卡 2-3 分钟）。
# - 恢复操作: 打包部署后执行 UPDATE sys_tasks SET is_enabled=1 WHERE task_id IN (22,23,28,29);
# - 验证: go build ./... 通过（无输出）。
#
# ── OEE 逻辑日临界修复（2026-04-15）── GetShiftsForLogicalDay / buildShiftOEEWindow / buildScheduleOEEWindow
# - 问题: GetShiftsForLogicalDay 使用硬编码 const logicalBoundaryHour=5 定位逻辑日锚点，
#   且计算 shiftStart/shiftEnd 时直接用 logicalBase 而不考虑日历偏移，
#   导致三班(0:00-7:40) 的 hasArrived/isCurrent 判断错误，
#   buildShiftOEEWindow/buildScheduleOEEWindow 的时间窗口 SQL 也指向错误日期。
#   影响范围: Cockpit.vue 实时 OEE、OEEDebug.vue 调试数据均错误。
# - 修复:
#   ① desktop/app_shifts.go — LogicalDayShift 结构体新增 CalendarDayOffset int 字段。
#   ② GetShiftsForLogicalDay: 移除 logicalBoundaryHour=5 硬编码；改用 active[0]（sort_order最小）
#      的 startMin 作为 logicalBoundaryMin；对每个班次若 startMin < logicalBoundaryMin → offset=1；
#      以 calendarBase=logicalBase+offset 计算时间戳和设置 CalendarDayOffset。
#   ③ desktop/app.go — buildShiftOEEWindow: 新增 hourOffset=CalendarDayOffset*24；
#      startHourSQL/endHourSQL/HourStart/WorkStart/WorkEnd 全部加 hourOffset（保留原 crossMidnight 加24逻辑）；
#      break 的 startH/endH 基础值改为 b.StartHour+hourOffset / b.EndHour+hourOffset。
#   ④ desktop/app.go — buildScheduleOEEWindow: workStartMin/lastEndMin/lastStartMin/sStartMin/sEndRaw
#      均加 +CalendarDayOffset*24*60，确保三班分钟数落在正确区间。
# - 三班验证示例: 三班(0:00-7:40, CalendarDayOffset=1)
#   → WorkStart="24:00", WorkEnd="31:40", HourStart=24, HourEnd=32
#   → ADDTIME('2026-04-14','24:00:00')=2026-04-15 00:00:00 ✓
# - 验证: go build ./... 通过（无输出）。
#
# ── ShiftReport 班次排序修复（2026-04-14）────────────────────────────────────
# - 问题: GetShiftSnapshots ORDER BY 使用 shift_id ASC，导致同一天内早/中/夜班顺序固定按 id 升序，
#   与实际排班时间不符（如夜班 sort_order=2 应排最前，不是最后）。
# - 修复: desktop/app_snapshot.go — GetShiftSnapshots ORDER BY 改为：
#   snapshot_date DESC, (SELECT sort_order FROM sys_shifts WHERE id = shift_id) DESC, device_id ASC
#   以 sys_shifts.sort_order DESC 为同一天内的班次排序依据，sort_order 值越大排越前。
# - 数据对应: sort_order=0 早班 / sort_order=1 中班 / sort_order=2 夜班 → 夜班在最上方。
# - 错误修正: 同次 code session 中曾误对前端班组下拉做 .reverse()，已回滚为后端原始顺序。
#   sys_teams 表无 sort_order 字段，班组排序以后端查询顺序（id 升序）为准。
#
# ── 驾驶舱布局优化（2026-04-14）──────────────────────────────────────────────
# - Cockpit.vue KPI 卡片布局重构:
#   .kpi-cards: 新增 height:14vh, align-items:stretch
#   .kpi-card: 改为 display:flex, flex-direction:column, padding-bottom:2.8vh
#   .kpi-shift-history: margin-top:auto（历史班次信息贴底）, position:relative, z-index:2
#   .kpi-sparkline: 高度从 30% 降至 24%
#   .kpi-dual-label: 从 margin-bottom:0.5vh 改为 margin-top:2px（标签移到数值下方）
#   .kpi-device-quality: flex-direction 从 column 改为 row + flex-wrap:wrap + gap:0 0.8vw（设备行横排）
#   .kpi-shift-row: line-height 从 1.6 降至 1.35
#   模板变更：OEE/Performance 双栏卡片中 kpi-dual-label 移到 kpi-dual-value 之后（先显示数值，再显示标签）
