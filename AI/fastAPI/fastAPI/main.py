# main.py
# 首先导入编码修复模块，确保全局编码设置
import encoding_fix

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
from contextlib import asynccontextmanager
# from routers import query_sql  # SQL 接口已禁用
from routers import knowledge
from routers import function_management
import os
from pathlib import Path
# 暂时注释掉SQLite数据库初始化
# from models.knowledge import init_sqlite_db
import config

@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用启动和关闭时的生命周期管理"""
    # 启动时初始化数据库（暂时禁用）
    print("基础API服务启动完成!")
    # await init_sqlite_db()
    
    # 在应用启动时预加载知识库服务（模型初始化）
    from services.knowledge_service import knowledge_service
    print("🚀 [启动] 开始预加载知识库服务...")
    await knowledge_service.initialize()
    print("✅ [启动] 知识库服务预加载完成！")
    
    yield
    
    # 关闭时的清理工作（如果需要的话）
    print("应用正在关闭...")

app = FastAPI(
    title="智能知识库 API",
    description="基于RAG技术的智能知识库系统，支持知识管理和智能问答",
    version="1.0.0",
    lifespan=lifespan
)

# 添加 CORS 中间件
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # 允许所有来源跨域（开发用）
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# 注册接口
app.include_router(knowledge.router, prefix="/api/knowledge", tags=["知识库"])

# 根据配置决定是否注册函数管理路由
if not config.PURE_RAG_MODE and config.ENABLE_FUNCTION_CALLING:
    app.include_router(function_management.router, prefix="/api/functions", tags=["函数管理"])
    print(f"✅ 智能模式已启用：支持 RAG + 函数调用")
else:
    print(f"✅ 纯 RAG 模式已启用：仅支持向量检索问答")

# 挂载静态文件（如果有静态资源文件夹的话）
static_dir = Path("static")
if static_dir.exists():
    app.mount("/static", StaticFiles(directory="static"), name="static")

# 添加HTML页面路由
@app.get("/knowledge_chat.html", summary="知识库聊天界面")
async def knowledge_chat():
    """返回知识库聊天界面HTML文件"""
    html_file = Path("knowledge_chat.html")
    if html_file.exists():
        return FileResponse(html_file, media_type="text/html")
    else:
        return {"error": "Chat interface not found"}

@app.get("/示例测试页面.html", summary="示例测试页面")
async def sample_page():
    """返回示例测试页面HTML文件"""
    html_file = Path("示例测试页面.html")
    if html_file.exists():
        return FileResponse(html_file, media_type="text/html")
    else:
        return {"error": "Sample page not found"}

@app.get("/", summary="API根路径")
async def root():
    """API根路径，返回服务基本信息"""
    # 根据配置动态生成功能列表和端点
    features = ["知识库管理（RAG）", "智能问答"]
    endpoints = {
        "docs": "/docs",
        "redoc": "/redoc",
        "knowledge": "/api/knowledge",
        "chat_interface": "/knowledge_chat.html"
    }
    
    mode = "纯 RAG 模式" if config.PURE_RAG_MODE else "智能模式（RAG + 函数调用）"
    
    if not config.PURE_RAG_MODE and config.ENABLE_FUNCTION_CALLING:
        features.extend(["函数调用（Function Calling）", "函数注册和管理", "意图识别"])
        endpoints["functions"] = "/api/functions"
    
    return {
        "message": "智能知识库 API",
        "version": "1.0.0",
        "mode": mode,
        "description": "基于RAG技术的智能知识库系统",
        "features": features,
        "endpoints": endpoints,
        "config": {
            "pure_rag_mode": config.PURE_RAG_MODE,
            "function_calling_enabled": config.ENABLE_FUNCTION_CALLING if not config.PURE_RAG_MODE else False
        }
    }

# 添加一个函数供 uvicorn 加载 host/port
def get_config():
    return config.HOST, config.PORT
