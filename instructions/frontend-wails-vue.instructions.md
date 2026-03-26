# [前端规则] Vue 3 + Wails Desktop

## 长期事实

1. 当前前端使用 Vue 3 + Vite + Naive UI。
2. 路由采用 Hash 模式，以适配 Wails 桌面壳路径环境。
3. 页面主要通过 window.go.main.App 或 wailsjs 调用后端绑定方法。

## 修改前端时的默认规则

1. 优先复用现有 Wails 绑定能力，不要为了图省事先补 fetch/axios API 层。
2. 保持桌面语义，注意窗口控制、全屏、标题栏等 Wails 场景，不要按纯浏览器站点假设写交互。
3. 新增页面或复杂逻辑时，优先和现有页面风格、Naive UI 用法、目录结构保持一致。
4. 如果改动涉及业务核心页面或关键状态流，补充中英日三语说明。

## 改动后至少验证

1. wails dev 启动正常。
2. 页面能正确调用绑定方法。
3. 控制台没有 window.go 或 runtime 未定义类错误。
