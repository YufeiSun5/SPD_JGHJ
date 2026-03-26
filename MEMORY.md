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
