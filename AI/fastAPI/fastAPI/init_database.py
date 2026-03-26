#!/usr/bin/env python3
# init_database.py
"""
数据库初始化脚本
用于设置SQLite数据库和ChromaDB向量数据库
"""

import asyncio
import os
import sys
from models.knowledge import init_sqlite_db
import config

async def init_databases():
    """初始化所有数据库"""
    print("=" * 50)
    print("开始初始化数据库...")
    print("=" * 50)
    
    try:
        # 1. 初始化SQLite数据库
        print("1. 正在初始化SQLite数据库...")
        await init_sqlite_db()
        print("   ✅ SQLite数据库初始化完成")
        
        # 2. 确保ChromaDB目录存在
        print("2. 正在设置ChromaDB...")
        os.makedirs(config.CHROMA_DB_PATH, exist_ok=True)
        print(f"   ✅ ChromaDB目录已创建: {config.CHROMA_DB_PATH}")
        
        # 3. 检查模型文件
        print("3. 正在检查模型文件...")
        if os.path.exists(config.DEEPSEEK_MODEL_PATH):
            print(f"   ✅ DeepSeek模型文件存在: {config.DEEPSEEK_MODEL_PATH}")
        else:
            print(f"   ⚠️  警告: DeepSeek模型文件不存在: {config.DEEPSEEK_MODEL_PATH}")
            print(f"      请确保模型文件已正确放置在指定位置")
        
        print("\n" + "=" * 50)
        print("数据库初始化完成! 🎉")
        print("=" * 50)
        print("\n接下来你可以:")
        print("1. 安装依赖: pip install -r requirements.txt")
        print("2. 启动服务: python -m uvicorn main:app --reload")
        print("3. 访问API文档: http://localhost:8004/docs")
        print("\nAPI接口预览:")
        print("- 添加知识: POST /api/knowledge/add")
        print("- 删除知识: DELETE /api/knowledge/delete/{knowledge_id}")
        print("- 智能问答: POST /api/knowledge/query")
        print("- 知识列表: GET /api/knowledge/list")
        print("- 健康检查: GET /api/knowledge/health")
        
    except Exception as e:
        print(f"❌ 数据库初始化失败: {e}")
        sys.exit(1)

if __name__ == "__main__":
    asyncio.run(init_databases())





























