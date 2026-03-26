#!/usr/bin/env python3
# start_knowledge_api.py
"""
知识库API启动脚本
自动初始化数据库并启动FastAPI服务
"""

import asyncio
import subprocess
import sys
import os
from models.knowledge import init_sqlite_db
import config

async def check_and_init():
    """检查并初始化数据库"""
    try:
        print("正在检查数据库状态...")
        
        # 检查SQLite数据库文件是否存在
        db_file = config.SQLITE_DATABASE_URL.replace("sqlite+aiosqlite:///", "")
        if not os.path.exists(db_file):
            print("SQLite数据库不存在，正在初始化...")
            await init_sqlite_db()
            print("✅ SQLite数据库初始化完成")
        else:
            print("✅ SQLite数据库已存在")
        
        # 检查ChromaDB目录
        if not os.path.exists(config.CHROMA_DB_PATH):
            print("ChromaDB目录不存在，正在创建...")
            os.makedirs(config.CHROMA_DB_PATH, exist_ok=True)
            print("✅ ChromaDB目录创建完成")
        else:
            print("✅ ChromaDB目录已存在")
        
        # 检查模型文件
        if os.path.exists(config.DEEPSEEK_MODEL_PATH):
            print("✅ DeepSeek模型文件存在")
        else:
            print(f"⚠️  警告: DeepSeek模型文件不存在: {config.DEEPSEEK_MODEL_PATH}")
            print("请确保模型文件已正确放置")
        
        return True
        
    except Exception as e:
        print(f"❌ 初始化检查失败: {e}")
        return False

def start_server():
    """启动FastAPI服务器"""
    try:
        print(f"\n🚀 启动知识库API服务...")
        print(f"   地址: http://{config.HOST}:{config.PORT}")
        print(f"   API文档: http://{config.HOST}:{config.PORT}/docs")
        print(f"   ReDoc文档: http://{config.HOST}:{config.PORT}/redoc")
        print("\n按 Ctrl+C 停止服务\n")
        
        # 启动uvicorn服务器
        cmd = [
            sys.executable, "-m", "uvicorn", 
            "main:app",
            "--host", config.HOST,
            "--port", str(config.PORT),
            "--reload"
        ]
        
        subprocess.run(cmd)
        
    except KeyboardInterrupt:
        print("\n\n👋 服务已停止")
    except Exception as e:
        print(f"❌ 启动服务失败: {e}")

async def main():
    """主函数"""
    print("=" * 60)
    print("🧠 智能知识库 API 启动器")
    print("=" * 60)
    
    # 检查和初始化
    if await check_and_init():
        print("\n✅ 所有检查通过，准备启动服务...")
        # 启动服务器
        start_server()
    else:
        print("\n❌ 初始化失败，无法启动服务")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(main())





























