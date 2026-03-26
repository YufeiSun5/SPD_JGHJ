# ============================================
# Wails 应用打包脚本 (离线可用)
# ============================================

Write-Host "🚀 开始打包 IIoT 监控客户端..." -ForegroundColor Cyan
Write-Host ""

# 检查是否在 desktop 目录
$currentPath = Get-Location
if ($currentPath.Path -notlike "*desktop") {
    Write-Host "❌ 请在 desktop 目录下运行此脚本" -ForegroundColor Red
    Write-Host "   运行命令: cd desktop ; .\build.ps1" -ForegroundColor Yellow
    exit 1
}

# 检查 Wails 是否安装
Write-Host "📋 检查 Wails 环境..." -ForegroundColor Yellow
$wailsVersion = wails version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 未检测到 Wails，请先安装 Wails CLI" -ForegroundColor Red
    Write-Host "   安装命令: go install github.com/wailsapp/wails/v2/cmd/wails@latest" -ForegroundColor Yellow
    exit 1
}
Write-Host "✅ $wailsVersion" -ForegroundColor Green
Write-Host ""

# 清理旧的构建
Write-Host "🧹 清理旧的构建文件..." -ForegroundColor Yellow
if (Test-Path "build\bin") {
    Remove-Item -Path "build\bin\*.exe" -Force -ErrorAction SilentlyContinue
    Write-Host "✅ 已清理旧的可执行文件" -ForegroundColor Green
}
Write-Host ""

# 开始构建
Write-Host "🔨 开始构建应用 (生产模式)..." -ForegroundColor Cyan
Write-Host "   这可能需要几分钟时间..." -ForegroundColor Gray
Write-Host ""

wails build -clean

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "✅ 打包成功！" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "📦 输出文件位置:" -ForegroundColor Cyan
    $exePath = Join-Path $currentPath "build\bin\iot-monitor.exe"
    Write-Host "   $exePath" -ForegroundColor White
    Write-Host ""
    
    # 检查文件大小
    if (Test-Path $exePath) {
        $fileSize = (Get-Item $exePath).Length / 1MB
        Write-Host "📊 文件大小: $([math]::Round($fileSize, 2)) MB" -ForegroundColor Cyan
        Write-Host ""
    }
    
    Write-Host "💡 部署说明:" -ForegroundColor Yellow
    Write-Host "   1. 将 iot-monitor.exe 复制到目标机器" -ForegroundColor White
    Write-Host "   2. 确保目标机器已安装:" -ForegroundColor White
    Write-Host "      - Microsoft Edge WebView2 Runtime" -ForegroundColor Gray
    Write-Host "      - 下载地址: https://go.microsoft.com/fwlink/p/?LinkId=2124703" -ForegroundColor Gray
    Write-Host "   3. 将配置文件 config.env 放在同级目录" -ForegroundColor White
    Write-Host "   4. 双击 iot-monitor.exe 运行" -ForegroundColor White
    Write-Host ""
    Write-Host "✨ 应用已打包为离线可用版本，无需网络即可运行界面！" -ForegroundColor Green
    Write-Host ""
    
} else {
    Write-Host ""
    Write-Host "❌ 打包失败！" -ForegroundColor Red
    Write-Host "请检查错误信息并修复后重试" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}



























