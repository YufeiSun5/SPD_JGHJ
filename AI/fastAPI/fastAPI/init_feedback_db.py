# init_feedback_db.py
# -*- coding: utf-8 -*-
"""
初始化反馈数据库表
运行此脚本以创建 knowledge_feedback 表
"""
import asyncio
import sys
import io

# 设置标准输出为 UTF-8
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

from models.knowledge import Base, sqlite_engine

async def init_feedback_table():
    """初始化反馈表"""
    print("开始初始化反馈数据库表...")
    
    async with sqlite_engine.begin() as conn:
        # 创建所有表（包括新的 knowledge_feedback 表）
        await conn.run_sync(Base.metadata.create_all)
    
    print("反馈数据库表初始化完成！")
    print("已创建表:")
    print("   - knowledge (知识库)")
    print("   - knowledge_feedback (点赞反馈)")

if __name__ == "__main__":
    asyncio.run(init_feedback_table())

