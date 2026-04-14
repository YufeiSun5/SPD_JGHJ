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
