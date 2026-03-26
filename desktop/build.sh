#!/bin/bash

# ============================================
# Wails 应用打包脚本 (离线可用)
# ============================================

echo "🚀 开始打包 IIoT 监控客户端..."
echo ""

# 检查是否在 desktop 目录
if [[ ! $(pwd) =~ desktop$ ]]; then
    echo "❌ 请在 desktop 目录下运行此脚本"
    echo "   运行命令: cd desktop && ./build.sh"
    exit 1
fi

# 检查 Wails 是否安装
echo "📋 检查 Wails 环境..."
if ! command -v wails &> /dev/null; then
    echo "❌ 未检测到 Wails，请先安装 Wails CLI"
    echo "   安装命令: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi
wails version
echo ""

# 清理旧的构建
echo "🧹 清理旧的构建文件..."
if [ -d "build/bin" ]; then
    rm -f build/bin/*
    echo "✅ 已清理旧的可执行文件"
fi
echo ""

# 开始构建
echo "🔨 开始构建应用 (生产模式)..."
echo "   这可能需要几分钟时间..."
echo ""

wails build -clean

if [ $? -eq 0 ]; then
    echo ""
    echo "========================================"
    echo "✅ 打包成功！"
    echo "========================================"
    echo ""
    echo "📦 输出文件位置:"
    ls -lh build/bin/
    echo ""
    echo "💡 部署说明:"
    echo "   1. 将可执行文件复制到目标机器"
    echo "   2. 将配置文件 config.env 放在同级目录"
    echo "   3. 运行程序"
    echo ""
    echo "✨ 应用已打包为离线可用版本，无需网络即可运行界面！"
    echo ""
else
    echo ""
    echo "❌ 打包失败！"
    echo "请检查错误信息并修复后重试"
    echo ""
    exit 1
fi



























