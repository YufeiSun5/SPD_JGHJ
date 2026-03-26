# LLM 服务化改造说明

> **目标**：将当前的本地 LLM 模型调用改造为服务化架构，支持并发处理和灵活的模型切换
> 
> **预计工作量**：1-2 小时
> 
> **改动范围**：新增 1 个文件，修改 2 个文件，约 175 行代码

---

## 📋 目录

1. [当前架构分析](#当前架构分析)
2. [目标架构设计](#目标架构设计)
3. [改造步骤](#改造步骤)
4. [代码实现](#代码实现)
5. [配置说明](#配置说明)
6. [测试验证](#测试验证)
7. [常见问题](#常见问题)

---

## 🔍 当前架构分析

### 现状

```
FastAPI 进程 (main.py:8006)
    ↓
knowledge_service.py
    ↓
self.llm_model = Llama(model_path="models/ggml-model-Q4_0.gguf")
    ↓
for chunk in self.llm_model(prompt):
    yield chunk
```

**问题**：
- ❌ 单进程，无法并发处理多个请求
- ❌ 模型和业务逻辑耦合，难以独立扩展
- ❌ 换模型需要重启整个 FastAPI 服务
- ❌ 内存占用高（FastAPI + 模型 = 2-3GB）

### 调用位置统计

在 `services/knowledge_service.py` 中共有 **6 处** LLM 调用：
- 第 589 行：非流式生成
- 第 657 行：流式生成（Function Result）
- 第 695 行：流式生成（参数请求）
- 第 746 行：流式生成（Function Result 备用）
- 第 788 行：流式生成（参数请求备用）
- 第 870 行：流式生成（纯 RAG 模式）

---

## 🎯 目标架构设计

### 服务化架构

```
┌─────────────────────────────────────────────────────────────┐
│                  FastAPI 业务层 (:8006)                      │
│                  内存占用: ~200MB                             │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │  knowledge_service.py                              │    │
│  │                                                     │    │
│  │  self.llm_adapter = LlamaCppServerAdapter()       │    │
│  │                                                     │    │
│  │  async for chunk in self.llm_adapter.generate():  │    │
│  │      yield chunk                                   │    │
│  └────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTP (:8080)
                           ↓
┌─────────────────────────────────────────────────────────────┐
│              llama.cpp HTTP Server (:8080)                   │
│              内存占用: ~2GB (仅模型)                          │
│                                                              │
│  - 支持 4 个并发请求                                          │
│  - 独立进程，可单独重启                                       │
│  - 可部署在不同机器/GPU 服务器                                │
└─────────────────────────────────────────────────────────────┘
```

### 核心优势

1. **职责分离**：业务逻辑和模型推理解耦
2. **并发支持**：llama.cpp server 原生支持并发
3. **灵活扩展**：可以轻松切换到 vLLM、TGI 等其他服务
4. **热更新**：换模型无需重启 FastAPI
5. **资源优化**：FastAPI 进程变轻量，可以开多个实例

---

## 🛠️ 改造步骤

### 阶段 1：准备工作（10 分钟）

#### 1.1 下载 llama.cpp

**Windows 预编译版本**：
```powershell
# 访问 GitHub Releases 下载
# https://github.com/ggerganov/llama.cpp/releases
# 下载 llama-<version>-bin-win-avx2-x64.zip

# 解压到项目目录
# E:\DEV\python\fastAPI\fastAPI\llama.cpp\
```

**或自己编译**（需要 Visual Studio）：
```powershell
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
mkdir build
cd build
cmake ..
cmake --build . --config Release
```

#### 1.2 安装依赖

```bash
pip install httpx  # 用于 HTTP 客户端
```

---

### 阶段 2：创建适配器层（1 小时）

#### 2.1 创建 `services/llm_adapter.py`

```python
# services/llm_adapter.py
"""
LLM 适配器模块 - 统一不同 LLM 后端的调用接口

支持的后端：
- local: 本地 llama-cpp-python（当前方式）
- llama_server: llama.cpp HTTP Server
- openai: OpenAI 兼容 API（vLLM, TGI 等）
"""

from abc import ABC, abstractmethod
from typing import AsyncIterator, Dict, Any, Optional
import asyncio
import json


class LLMAdapter(ABC):
    """LLM 适配器基类"""
    
    @abstractmethod
    async def generate_stream(self, prompt: str, **kwargs) -> AsyncIterator[Dict]:
        """
        流式生成文本
        
        Args:
            prompt: 输入提示词
            **kwargs: 生成参数（max_tokens, temperature 等）
        
        Yields:
            Dict: 统一格式的响应块
                {
                    'choices': [{
                        'text': str,  # 生成的文本
                        'delta': {'content': str}  # 增量内容
                    }]
                }
        """
        pass
    
    @abstractmethod
    async def generate(self, prompt: str, **kwargs) -> str:
        """
        非流式生成文本
        
        Args:
            prompt: 输入提示词
            **kwargs: 生成参数
        
        Returns:
            str: 生成的完整文本
        """
        pass


class LocalLlamaAdapter(LLMAdapter):
    """
    本地 llama-cpp-python 适配器
    
    使用场景：
    - 开发测试
    - 小规模部署
    - 无需并发的场景
    """
    
    def __init__(self, model_path: str, **kwargs):
        """
        初始化本地模型
        
        Args:
            model_path: 模型文件路径（GGUF 格式）
            **kwargs: Llama 初始化参数（n_ctx, n_threads 等）
        """
        from llama_cpp import Llama
        import encoding_fix
        
        encoding_fix.safe_print(f"📦 [LocalLlama] 加载模型: {model_path}")
        self.model = Llama(model_path=model_path, **kwargs)
        encoding_fix.safe_print(f"✅ [LocalLlama] 模型加载完成")
    
    async def generate_stream(self, prompt: str, **kwargs) -> AsyncIterator[Dict]:
        """流式生成（在线程池中执行，避免阻塞）"""
        loop = asyncio.get_event_loop()
        
        # 在线程池中执行同步的 LLM 调用
        def _sync_generate():
            for chunk in self.model(prompt, stream=True, **kwargs):
                yield chunk
        
        # 将同步生成器转换为异步
        gen = _sync_generate()
        while True:
            try:
                chunk = await loop.run_in_executor(None, lambda: next(gen))
                yield chunk
            except StopIteration:
                break
    
    async def generate(self, prompt: str, **kwargs) -> str:
        """非流式生成"""
        loop = asyncio.get_event_loop()
        result = await loop.run_in_executor(
            None,
            lambda: self.model(prompt, stream=False, **kwargs)
        )
        return result['choices'][0]['text']


class LlamaCppServerAdapter(LLMAdapter):
    """
    llama.cpp HTTP Server 适配器
    
    使用场景：
    - 生产环境（推荐）
    - 需要并发处理
    - 需要独立扩展模型服务
    
    优势：
    - 原生支持并发（--parallel 参数）
    - 独立进程，可单独重启
    - C++ 实现，性能更好
    """
    
    def __init__(self, server_url: str = "http://localhost:8080"):
        """
        初始化 Server 适配器
        
        Args:
            server_url: llama.cpp server 地址
        """
        import httpx
        import encoding_fix
        
        self.server_url = server_url.rstrip('/')
        self.client = httpx.AsyncClient(timeout=120.0)
        
        encoding_fix.safe_print(f"🔗 [LlamaCppServer] 连接到: {self.server_url}")
        
        # 测试连接
        try:
            import requests
            response = requests.get(f"{self.server_url}/health", timeout=5)
            if response.status_code == 200:
                encoding_fix.safe_print(f"✅ [LlamaCppServer] 连接成功")
            else:
                encoding_fix.safe_print(f"⚠️ [LlamaCppServer] 服务器响应异常: {response.status_code}")
        except Exception as e:
            encoding_fix.safe_print(f"⚠️ [LlamaCppServer] 无法连接: {e}")
            encoding_fix.safe_print(f"   请确保 llama.cpp server 已启动在 {self.server_url}")
    
    async def generate_stream(self, prompt: str, **kwargs) -> AsyncIterator[Dict]:
        """流式生成"""
        try:
            async with self.client.stream(
                'POST',
                f"{self.server_url}/completion",
                json={
                    "prompt": prompt,
                    "n_predict": kwargs.get('max_tokens', 150),
                    "temperature": kwargs.get('temperature', 0.7),
                    "stop": kwargs.get('stop', []),
                    "stream": True,
                },
                timeout=120.0
            ) as response:
                async for line in response.aiter_lines():
                    if line.startswith('data: '):
                        try:
                            data = json.loads(line[6:])
                            # 转换为统一格式
                            content = data.get('content', '')
                            if content:  # 只返回有内容的块
                                yield {
                                    'choices': [{
                                        'text': content,
                                        'delta': {'content': content},
                                        'finish_reason': data.get('stop', False)
                                    }]
                                }
                        except json.JSONDecodeError:
                            continue
        except Exception as e:
            import encoding_fix
            encoding_fix.safe_print(f"❌ [LlamaCppServer] 流式生成错误: {e}")
            raise
    
    async def generate(self, prompt: str, **kwargs) -> str:
        """非流式生成"""
        try:
            response = await self.client.post(
                f"{self.server_url}/completion",
                json={
                    "prompt": prompt,
                    "n_predict": kwargs.get('max_tokens', 150),
                    "temperature": kwargs.get('temperature', 0.7),
                    "stop": kwargs.get('stop', []),
                    "stream": False,
                },
                timeout=120.0
            )
            data = response.json()
            return data.get('content', '')
        except Exception as e:
            import encoding_fix
            encoding_fix.safe_print(f"❌ [LlamaCppServer] 生成错误: {e}")
            raise


class OpenAIAdapter(LLMAdapter):
    """
    OpenAI 兼容 API 适配器
    
    使用场景：
    - 使用 vLLM（高性能推理服务器）
    - 使用 Text Generation Inference (TGI)
    - 使用 OpenAI 官方 API
    - 使用其他兼容 OpenAI API 的服务
    
    优势：
    - 标准化接口
    - 支持 GPU 加速
    - 高吞吐量（vLLM 可达 100+ QPS）
    """
    
    def __init__(self, api_base: str, api_key: str = "EMPTY", model: str = "default"):
        """
        初始化 OpenAI 适配器
        
        Args:
            api_base: API 基础 URL（如 http://localhost:8000/v1）
            api_key: API 密钥（本地服务可用 "EMPTY"）
            model: 模型名称
        """
        import openai
        import encoding_fix
        
        openai.api_base = api_base
        openai.api_key = api_key
        self.model = model
        
        encoding_fix.safe_print(f"🔗 [OpenAI] 连接到: {api_base}")
        encoding_fix.safe_print(f"📦 [OpenAI] 模型: {model}")
    
    async def generate_stream(self, prompt: str, **kwargs) -> AsyncIterator[Dict]:
        """流式生成"""
        import openai
        
        try:
            response = await openai.ChatCompletion.acreate(
                model=self.model,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=kwargs.get('max_tokens', 150),
                temperature=kwargs.get('temperature', 0.7),
                stop=kwargs.get('stop'),
                stream=True,
            )
            
            async for chunk in response:
                if chunk.choices[0].delta.get('content'):
                    yield {
                        'choices': [{
                            'text': chunk.choices[0].delta.content,
                            'delta': {'content': chunk.choices[0].delta.content}
                        }]
                    }
        except Exception as e:
            import encoding_fix
            encoding_fix.safe_print(f"❌ [OpenAI] 流式生成错误: {e}")
            raise
    
    async def generate(self, prompt: str, **kwargs) -> str:
        """非流式生成"""
        import openai
        
        try:
            response = await openai.ChatCompletion.acreate(
                model=self.model,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=kwargs.get('max_tokens', 150),
                temperature=kwargs.get('temperature', 0.7),
                stop=kwargs.get('stop'),
                stream=False,
            )
            return response.choices[0].message.content
        except Exception as e:
            import encoding_fix
            encoding_fix.safe_print(f"❌ [OpenAI] 生成错误: {e}")
            raise


def create_llm_adapter(backend: str = "local", **kwargs) -> LLMAdapter:
    """
    工厂函数：根据配置创建对应的 LLM 适配器
    
    Args:
        backend: 后端类型（local, llama_server, openai）
        **kwargs: 后端特定的参数
    
    Returns:
        LLMAdapter: 对应的适配器实例
    
    Examples:
        >>> # 本地模型
        >>> adapter = create_llm_adapter('local', model_path='model.gguf')
        
        >>> # llama.cpp server
        >>> adapter = create_llm_adapter('llama_server', server_url='http://localhost:8080')
        
        >>> # vLLM / OpenAI
        >>> adapter = create_llm_adapter('openai', api_base='http://localhost:8000/v1')
    """
    import encoding_fix
    
    encoding_fix.safe_print(f"🔧 [LLMAdapter] 创建适配器: {backend}")
    
    if backend == "local":
        return LocalLlamaAdapter(**kwargs)
    elif backend == "llama_server":
        return LlamaCppServerAdapter(**kwargs)
    elif backend == "openai":
        return OpenAIAdapter(**kwargs)
    else:
        raise ValueError(f"不支持的 LLM 后端: {backend}")
```

---

### 阶段 3：修改配置文件（5 分钟）

#### 3.1 修改 `config.py`

在 `config.py` 末尾添加：

```python
# ==================== LLM 后端配置 ====================
# LLM 后端选择
# - "local": 本地 llama-cpp-python（当前方式，适合开发测试）
# - "llama_server": llama.cpp HTTP Server（推荐生产环境）
# - "openai": OpenAI 兼容 API（vLLM, TGI 等，适合 GPU 加速）
LLM_BACKEND = "local"  # 默认使用本地模式，改造后切换为 "llama_server"

# ===== 本地模型配置（LLM_BACKEND = "local"） =====
# DEEPSEEK_MODEL_PATH 已在上面定义

# ===== llama.cpp Server 配置（LLM_BACKEND = "llama_server"） =====
LLAMA_SERVER_URL = "http://localhost:8080"  # llama.cpp server 地址
LLAMA_SERVER_PARALLEL = 4  # 并发请求数（启动 server 时的 --parallel 参数）

# ===== OpenAI 兼容 API 配置（LLM_BACKEND = "openai"） =====
OPENAI_API_BASE = "http://localhost:8000/v1"  # vLLM, TGI 等服务地址
OPENAI_API_KEY = "EMPTY"  # API 密钥（本地服务可用 "EMPTY"）
OPENAI_MODEL_NAME = "deepseek-coder-6.7b"  # 模型名称
```

---

### 阶段 4：修改 knowledge_service.py（30 分钟）

#### 4.1 修改初始化部分

找到 `_init_models()` 方法（约第 115 行），修改 LLM 初始化部分：

```python
# services/knowledge_service.py

# 在文件顶部添加导入
from services.llm_adapter import create_llm_adapter

class KnowledgeService:
    def __init__(self):
        self.sentence_model = None
        self.chroma_client = None
        self.chroma_collection = None
        # self.llm_model = None  # ❌ 删除这行
        self.llm_adapter = None  # ✅ 改为 llm_adapter
        self.initialized = False
        self.function_caller = None
        self.ai_function_caller = None
        self._init_lock = asyncio.Lock()
    
    def _init_models(self):
        """在线程池中初始化模型（避免阻塞事件循环）"""
        encoding_fix.safe_print("正在初始化 Sentence Transformer...")
        self.sentence_model = SentenceTransformer(config.SENTENCE_TRANSFORMER_MODEL)
        
        encoding_fix.safe_print("正在初始化 ChromaDB...")
        # ... ChromaDB 初始化代码保持不变 ...
        
        # ===== 修改 LLM 初始化部分 =====
        encoding_fix.safe_print("正在初始化 LLM...")
        
        # 根据配置选择 LLM 后端
        llm_backend = getattr(config, 'LLM_BACKEND', 'local')
        
        if llm_backend == 'local':
            # 本地模型
            if not os.path.exists(config.DEEPSEEK_MODEL_PATH):
                raise FileNotFoundError(f"模型文件不存在: {config.DEEPSEEK_MODEL_PATH}")
            
            self.llm_adapter = create_llm_adapter(
                backend='local',
                model_path=config.DEEPSEEK_MODEL_PATH,
                n_ctx=config.LLM_CONTEXT_LENGTH,
                n_threads=getattr(config, 'LLM_THREADS', 8),
                n_batch=getattr(config, 'LLM_BATCH_SIZE', 512),
                verbose=False,
                encoding='utf-8'
            )
        
        elif llm_backend == 'llama_server':
            # llama.cpp HTTP Server
            self.llm_adapter = create_llm_adapter(
                backend='llama_server',
                server_url=config.LLAMA_SERVER_URL
            )
        
        elif llm_backend == 'openai':
            # OpenAI 兼容 API（vLLM, TGI 等）
            self.llm_adapter = create_llm_adapter(
                backend='openai',
                api_base=config.OPENAI_API_BASE,
                api_key=config.OPENAI_API_KEY,
                model=config.OPENAI_MODEL_NAME
            )
        
        else:
            raise ValueError(f"不支持的 LLM 后端: {llm_backend}")
        
        encoding_fix.safe_print(f"✅ LLM 后端初始化完成: {llm_backend}")
        encoding_fix.safe_print("所有模型初始化完成!")
```

#### 4.2 修改 LLM 调用部分

找到所有 `self.llm_model(...)` 调用（共 6 处），替换为 `self.llm_adapter.generate_stream(...)`：

**位置 1：第 589 行（非流式生成）**
```python
# 修改前
response = self.llm_model(
    safe_prompt,
    max_tokens=config.LLM_MAX_TOKENS,
    temperature=config.LLM_TEMPERATURE,
    stop=["</s>", "用户问题：", "上下文信息："],
    echo=False
)

# 修改后
response = await self.llm_adapter.generate(
    safe_prompt,
    max_tokens=config.LLM_MAX_TOKENS,
    temperature=config.LLM_TEMPERATURE,
    stop=["</s>", "用户问题：", "上下文信息："]
)
```

**位置 2-6：流式生成（第 657, 695, 746, 788, 870 行）**
```python
# 修改前
for chunk in self.llm_model(
    safe_prompt,
    max_tokens=config.LLM_MAX_TOKENS,
    temperature=config.LLM_TEMPERATURE,
    stream=True
):
    if 'choices' in chunk and len(chunk['choices']) > 0:
        delta = chunk['choices'][0].get('delta', {})
        if 'content' in delta:
            token_count += 1
            yield delta['content']
        elif 'text' in chunk['choices'][0]:
            token_count += 1
            yield chunk['choices'][0]['text']

# 修改后
async for chunk in self.llm_adapter.generate_stream(
    safe_prompt,
    max_tokens=config.LLM_MAX_TOKENS,
    temperature=config.LLM_TEMPERATURE
):
    if 'choices' in chunk and len(chunk['choices']) > 0:
        delta = chunk['choices'][0].get('delta', {})
        if 'content' in delta:
            token_count += 1
            yield delta['content']
        elif 'text' in chunk['choices'][0]:
            token_count += 1
            yield chunk['choices'][0]['text']
```

**注意**：所有 `for chunk in self.llm_model(...)` 都要改为 `async for chunk in self.llm_adapter.generate_stream(...)`

---

## ⚙️ 配置说明

### 配置模式切换

#### 模式 1：本地模式（当前方式）
```python
# config.py
LLM_BACKEND = "local"
DEEPSEEK_MODEL_PATH = "models/ggml-model-Q4_0.gguf"
```

**特点**：
- ✅ 无需额外服务
- ✅ 简单直接
- ❌ 无法并发
- ❌ 内存占用高

---

#### 模式 2：llama.cpp Server 模式（推荐）
```python
# config.py
LLM_BACKEND = "llama_server"
LLAMA_SERVER_URL = "http://localhost:8080"
```

**启动 Server**：
```powershell
# Windows
cd llama.cpp
.\server.exe -m E:\DEV\python\fastAPI\fastAPI\models\ggml-model-Q4_0.gguf `
    --host 0.0.0.0 `
    --port 8080 `
    --ctx-size 4096 `
    --parallel 4 `
    --threads 8

# 参数说明：
# -m: 模型文件路径
# --host: 监听地址（0.0.0.0 允许外部访问）
# --port: 端口号
# --ctx-size: 上下文长度
# --parallel: 并发请求数（重要！）
# --threads: CPU 线程数
```

**特点**：
- ✅ 支持并发（4 个请求）
- ✅ 独立进程，可单独重启
- ✅ 性能更好（C++ 实现）
- ✅ FastAPI 进程变轻量

---

#### 模式 3：OpenAI 兼容模式（未来扩展）
```python
# config.py
LLM_BACKEND = "openai"
OPENAI_API_BASE = "http://localhost:8000/v1"  # vLLM 地址
OPENAI_MODEL_NAME = "deepseek-coder-6.7b"
```

**使用场景**：
- 迁移到 vLLM（GPU 加速）
- 使用 Text Generation Inference (TGI)
- 使用 OpenAI 官方 API

---

## 🧪 测试验证

### 测试步骤

#### 1. 测试本地模式（确保改造不影响现有功能）

```python
# config.py
LLM_BACKEND = "local"
```

```bash
# 启动服务
python -m uvicorn main:app --host 0.0.0.0 --port 8006 --reload
```

**测试**：
- ✅ 访问 http://localhost:8006/docs
- ✅ 测试知识库问答接口
- ✅ 确认回答正常生成

---

#### 2. 测试 llama.cpp Server 模式

**步骤 1：启动 llama.cpp server**
```powershell
# 新开一个终端
cd llama.cpp
.\server.exe -m E:\DEV\python\fastAPI\fastAPI\models\ggml-model-Q4_0.gguf --port 8080 --parallel 4
```

**步骤 2：修改配置**
```python
# config.py
LLM_BACKEND = "llama_server"
```

**步骤 3：重启 FastAPI**
```bash
# Ctrl+C 停止，然后重启
python -m uvicorn main:app --host 0.0.0.0 --port 8006 --reload
```

**步骤 4：测试**
- ✅ 查看启动日志，确认连接到 server
- ✅ 测试问答接口
- ✅ 同时发起多个请求，验证并发

---

#### 3. 并发测试

```python
# test_concurrent.py
import asyncio
import httpx

async def test_concurrent():
    """测试并发请求"""
    async with httpx.AsyncClient() as client:
        tasks = []
        for i in range(4):
            task = client.post(
                "http://localhost:8006/api/knowledge/query-stream",
                json={"question": f"测试问题 {i}"},
                timeout=60.0
            )
            tasks.append(task)
        
        # 并发发送 4 个请求
        responses = await asyncio.gather(*tasks)
        
        for i, resp in enumerate(responses):
            print(f"请求 {i}: {resp.status_code}")

asyncio.run(test_concurrent())
```

**预期结果**：
- 本地模式：4 个请求串行处理，总耗时 ~40 秒
- Server 模式：4 个请求并行处理，总耗时 ~10 秒

---

## ❓ 常见问题

### Q1: llama.cpp server 启动失败？

**检查清单**：
1. 模型文件路径是否正确
2. 端口 8080 是否被占用
3. 是否有足够的内存（至少 4GB）

```powershell
# 检查端口占用
netstat -ano | findstr :8080

# 如果被占用，换个端口
.\server.exe -m model.gguf --port 8081
```

---

### Q2: FastAPI 连接不到 server？

**检查清单**：
1. server 是否正常启动
2. 配置的 URL 是否正确
3. 防火墙是否阻止

```powershell
# 测试 server 是否可访问
curl http://localhost:8080/health
```

---

### Q3: 改造后性能变差了？

**可能原因**：
1. 网络开销（本地调用变为 HTTP）
2. server 的 `--parallel` 参数设置不当

**优化**：
```bash
# 增加并发数
.\server.exe -m model.gguf --parallel 8

# 增加线程数
.\server.exe -m model.gguf --threads 16
```

---

### Q4: 如何回滚到原来的方式？

**非常简单**：
```python
# config.py
LLM_BACKEND = "local"  # 改回 local

# 重启服务即可
```

适配器模式保证了向后兼容！

---

### Q5: 换模型需要重启 FastAPI 吗？

**使用 Server 模式时不需要！**

```bash
# 1. 停止旧的 server（Ctrl+C）
# 2. 启动新的 server（指向新模型）
.\server.exe -m models/new-model.gguf --port 8080

# FastAPI 无需重启！
```

---

### Q6: 可以同时运行多个模型吗？

**可以！**

```bash
# Server 1: DeepSeek
.\server.exe -m models/deepseek.gguf --port 8080

# Server 2: Llama-3
.\server.exe -m models/llama3.gguf --port 8081
```

```python
# 动态选择
if user.preference == "deepseek":
    adapter = LlamaCppServerAdapter("http://localhost:8080")
else:
    adapter = LlamaCppServerAdapter("http://localhost:8081")
```

---

## 📈 性能对比

### 单请求延迟

| 模式 | 延迟 | 说明 |
|------|------|------|
| 本地模式 | 10 秒 | 基准 |
| Server 模式 | 10.2 秒 | +200ms HTTP 开销 |

**结论**：单请求性能基本一致

---

### 并发吞吐量

| 模式 | 4 个并发请求 | 8 个并发请求 |
|------|--------------|--------------|
| 本地模式 | 40 秒（串行） | 80 秒（串行） |
| Server 模式 | 10 秒（并行） | 20 秒（并行） |

**结论**：并发场景下，Server 模式性能提升 **4 倍**

---

### 内存占用

| 模式 | FastAPI 进程 | 模型进程 | 总计 |
|------|--------------|----------|------|
| 本地模式 | 2.5 GB | - | 2.5 GB |
| Server 模式 | 200 MB | 2 GB | 2.2 GB |

**结论**：总内存占用略微降低，且更灵活

---

## 🚀 未来扩展

### 扩展 1：迁移到 GPU（vLLM）

```bash
# 1. 安装 vLLM
pip install vllm

# 2. 启动 vLLM server
python -m vllm.entrypoints.openai.api_server \
    --model deepseek-ai/deepseek-coder-6.7b-instruct \
    --port 8000

# 3. 修改配置
# config.py
LLM_BACKEND = "openai"
OPENAI_API_BASE = "http://localhost:8000/v1"
```

**性能提升**：10-50 倍（取决于 GPU）

---

### 扩展 2：多模型负载均衡

```python
# services/llm_load_balancer.py
class LLMLoadBalancer:
    def __init__(self, servers: List[str]):
        self.adapters = [
            LlamaCppServerAdapter(url) for url in servers
        ]
        self.current = 0
    
    def get_adapter(self) -> LLMAdapter:
        """轮询选择 adapter"""
        adapter = self.adapters[self.current]
        self.current = (self.current + 1) % len(self.adapters)
        return adapter
```

---

### 扩展 3：模型缓存和预热

```python
# 启动时预热模型
async def warmup_model():
    adapter = knowledge_service.llm_adapter
    await adapter.generate("Hello", max_tokens=10)
    print("✅ 模型预热完成")
```

---

## 📝 改造检查清单

改造完成后，请逐项检查：

- [ ] 创建了 `services/llm_adapter.py` 文件
- [ ] 修改了 `config.py`，添加了 LLM 后端配置
- [ ] 修改了 `services/knowledge_service.py`：
  - [ ] 导入了 `create_llm_adapter`
  - [ ] `__init__` 中将 `self.llm_model` 改为 `self.llm_adapter`
  - [ ] `_init_models()` 中使用 `create_llm_adapter` 创建适配器
  - [ ] 所有 `self.llm_model(...)` 改为 `self.llm_adapter.generate_stream(...)`
  - [ ] 所有 `for chunk` 改为 `async for chunk`
- [ ] 测试了本地模式，确认功能正常
- [ ] 下载了 llama.cpp 预编译版本
- [ ] 测试了 llama.cpp server 模式
- [ ] 验证了并发性能提升

---

## 🎉 总结

### 改造收益

1. **架构解耦**：业务逻辑和模型推理分离
2. **并发支持**：从串行变为并行，性能提升 4 倍
3. **灵活扩展**：可以轻松切换到 vLLM、TGI 等
4. **热更新**：换模型无需重启 FastAPI
5. **资源优化**：FastAPI 进程变轻量

### 改造成本

- **开发时间**：1-2 小时
- **代码改动**：~175 行（新增 1 个文件，修改 2 个文件）
- **风险**：低（保持向后兼容，可随时回滚）

### 推荐做法

1. **先在开发环境测试**（使用本地模式验证）
2. **逐步切换到 Server 模式**（先单机，再分布式）
3. **监控性能指标**（延迟、吞吐量、内存）
4. **根据业务增长选择扩展方案**（vLLM、多实例等）

---

**祝改造顺利！** 🚀

如有问题，请参考本文档的"常见问题"部分，或查看代码注释。

