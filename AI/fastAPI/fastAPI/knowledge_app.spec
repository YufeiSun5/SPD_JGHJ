# -*- mode: python ; coding: utf-8 -*-

import sys
import os
from pathlib import Path

# 获取项目根目录
project_root = Path(os.path.abspath('.'))

a = Analysis(
    ['startup.py'],  # 主启动脚本
    pathex=[str(project_root)],
    binaries=[],
    datas=[
        # 包含模型文件
        ('models/*.gguf', 'models'),
        # 包含HTML文件
        ('*.html', '.'),
        # 包含数据库文件（如果存在）
        ('*.db', '.'),
        # 包含ChromaDB数据（如果存在）
        ('chroma_db', 'chroma_db'),
        # 包含配置文件
        ('config.py', '.'),
        ('main.py', '.'),
        ('db.py', '.'),
        ('init_database.py', '.'),
        # 包含路由和服务
        ('routers', 'routers'),
        ('services', 'services'),
        ('models', 'models'),
    ],
    hiddenimports=[
        'uvicorn',
        'uvicorn.lifespan',
        'uvicorn.lifespan.on',
        'uvicorn.loops',
        'uvicorn.loops.auto',
        'uvicorn.protocols',
        'uvicorn.protocols.http',
        'uvicorn.protocols.http.auto',
        'uvicorn.protocols.websockets',
        'uvicorn.protocols.websockets.auto',
        'fastapi',
        'pydantic',
        'starlette',
        'sentence_transformers',
        'transformers',
        'torch',
        'numpy',
        'sklearn',
        'chromadb',
        'sqlite3',
        'aiosqlite',
        'sqlalchemy',
        'llama_cpp',
        'asyncmy',
        'asyncio',
        'concurrent.futures',
        'logging',
        'json',
        'datetime',
        'typing',
        'email.mime',
        'email.mime.text',
        'email.mime.multipart',
        'asyncmy.cursors',
        'asyncmy.connection',
        'uvloop',
        'httptools',
        'websockets',
        'watchfiles',
        'python-multipart',
        'jinja2',
        'aiofiles',
    ],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    win_no_prefer_redirects=False,
    win_private_assemblies=False,
    cipher=None,
    noarchive=False,
)

pyz = PYZ(a.pure, a.zipped_data, cipher=None)

exe = EXE(
    pyz,
    a.scripts,
    [],
    exclude_binaries=True,
    name='知识库助手',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    console=True,  # 显示控制台窗口以查看日志
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)

coll = COLLECT(
    exe,
    a.binaries,
    a.zipfiles,
    a.datas,
    strip=False,
    upx=True,
    upx_exclude=[],
    name='知识库助手',
)
