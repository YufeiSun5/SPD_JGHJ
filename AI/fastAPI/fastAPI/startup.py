#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
知识库应用启动脚本
"""

import os
import sys
import webbrowser
import threading
import time
from pathlib import Path
import signal
import uvicorn

# 导入编码修复模块
import encoding_fix

# 设置工作目录为exe所在目录
if getattr(sys, 'frozen', False):
    # 如果是打包后的exe
    app_dir = Path(sys.executable).parent
    internal_dir = app_dir / "_internal"
    
    # 保持在exe目录，但设置正确的Python路径
    os.chdir(app_dir)
    
    # 检查_internal目录是否存在，将其添加到路径
    if internal_dir.exists():
        sys.path.insert(0, str(internal_dir))
    
    # 确保当前目录也在路径中
    sys.path.insert(0, str(app_dir))
else:
    # 如果是开发环境
    app_dir = Path(__file__).parent
    os.chdir(app_dir)
    sys.path.insert(0, str(app_dir))

def open_browser():
    """延迟打开浏览器"""
    time.sleep(3)  # 等待服务器启动
    try:
        # 打开知识库聊天界面
        chat_url = "http://localhost:8000/knowledge_chat.html"
        webbrowser.open(chat_url)
        print(f"[浏览器] 浏览器已打开: {chat_url}")
    except Exception as e:
        print(f"[警告] 无法自动打开浏览器: {e}")
        print("请手动打开浏览器访问: http://localhost:8000/knowledge_chat.html")

def signal_handler(signum, frame):
    """信号处理器"""
    print("\n[再见] 正在关闭知识库助手...")
    sys.exit(0)

def main():
    """主函数"""
    print("=" * 60)
    print("[机器人] 盛云知识问答小助手")
    print("=" * 60)
    print(f"[文件夹] 工作目录: {app_dir}")
    print(f"[Python] Python版本: {sys.version}")
    
    # 检查必要文件
    required_files = ['main.py', 'config.py']
    missing_files = [f for f in required_files if not Path(f).exists()]
    
    if missing_files:
        print(f"[错误] 缺少必要文件: {missing_files}")
        input("按Enter键退出...")
        return
    
    # 设置信号处理
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    try:
        # 启动浏览器线程
        browser_thread = threading.Thread(target=open_browser, daemon=True)
        browser_thread.start()
        
        print("[启动] 正在启动知识库服务...")
        print("[网页] Web界面: http://localhost:8000/knowledge_chat.html")
        print("[文档] API文档: http://localhost:8000/docs")
        print("\n按 Ctrl+C 退出")
        print("-" * 60)
        
        # 启动FastAPI应用
        import config
        uvicorn.run(
            "main:app",
            host=config.HOST,
            port=config.PORT,
            reload=False,
            log_level="info"
        )
        
    except KeyboardInterrupt:
        print("\n[中断] 用户中断，正在退出...")
    except Exception as e:
        print(f"[错误] 启动失败: {e}")
        input("按Enter键退出...")

if __name__ == "__main__":
    main()
