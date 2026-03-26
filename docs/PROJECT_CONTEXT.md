# 项目上下文摘要

## 项目类型

当前项目是桌面 SCADA/IIoT 系统，主技术栈为：

1. Go
2. Wails v2
3. Vue 3 + Vite + Naive UI
4. MQTT
5. MySQL

## 与旧项目的关键区别

1. 不是 Spring 三服务项目。
2. 不是 Taro 小程序项目。
3. 不是 BSSID/空间共振产品。
4. 主调用链不是 HTTP API，而是 Wails 直绑 Go 方法。

## 当前关键文件

1. desktop/main.go
   - 桌面启动与后端核心初始化入口。
2. desktop/app.go
   - Wails 绑定的主要业务方法集合。
3. desktop/frontend/src
   - Vue 页面与组件。
4. gateway/
   - MQTT 网关管理。
5. workers/
   - 逻辑、存储、报警、任务、事件处理等异步工作流。
6. docs/项目架构总览.md
   - 当前并发模型和数据流设计说明。

## 当前默认原则

1. MQTT 是现场实时数据主入口。
2. 数据和配置以数据库持久化，实时值以内存与通道解耦处理。
3. 不要随意把慢 IO 塞回热路径。
4. 不要为了“通用前后端架构”而硬补 REST 主链路。
5. 根目录暴露的业务说明文档默认收口在 docs/，数据库脚本默认收口在 sql/。
