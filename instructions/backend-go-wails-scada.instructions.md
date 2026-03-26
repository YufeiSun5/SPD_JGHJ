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

## 改动后至少检查

1. desktop/main.go 启动链路是否仍能初始化数据库、通道、任务调度与网关。
2. desktop/app.go 中的方法名是否与前端调用保持一致。
3. 若涉及 SQL、配置或热重载，至少保留一条可人工复查的验证路径。
