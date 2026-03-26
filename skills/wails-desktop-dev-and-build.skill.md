# [技能] Wails 桌面开发与构建流程

## 适用场景

当需要启动桌面开发环境、联调前端页面、或构建桌面安装产物时，使用本流程。

## 开发模式

```powershell
Set-Location c:/DEV_C/GO/SPD_JGHJ/desktop
wails dev
```

## 构建模式

```powershell
Set-Location c:/DEV_C/GO/SPD_JGHJ/desktop
wails build
```

## 执行前检查

1. desktop/frontend 依赖已安装。
2. MySQL 配置可用。
3. 如需要 AI 联调，本机 FastAPI 服务已启动在 127.0.0.1:8006。

## 常见排查

1. 窗口起不来：先查 WebView2 Runtime。
2. 页面能开但数据全空：先查数据库与网关配置。
3. 前端方法报错：先核对 window.go.main.App 方法名与 desktop/app.go 是否一致。
