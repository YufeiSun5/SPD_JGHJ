#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
简单的服务器启动脚本
避免 uvicorn --reload 的子进程问题
"""

import uvicorn
from main_rag_only import app

if __name__ == "__main__":
    print("正在启动 RAG 知识库服务...")
    print("服务地址: http://0.0.0.0:8004")
    print("API文档: http://localhost:8004/docs")
    print("按 Ctrl+C 停止服务")
    
    uvicorn.run(
        app,
        host="0.0.0.0",
        port=8004,
        reload=False,  # 关闭热重载避免子进程问题
        workers=1
    )




















