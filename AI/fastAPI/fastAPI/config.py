# config.py
import os

# FastAPI 配置
DATABASE_URL = "mysql+asyncmy://root:root@localhost:3306/opc_3d_test"
HOST = "0.0.0.0"  # 绑定IP（或用 0.0.0.0 听全部）
PORT = 8006             # 监听端口

# SQLite 配置（用于知识库元数据）
SQLITE_DATABASE_URL = "sqlite+aiosqlite:///./knowledge.db"

# 模型配置
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
DEEPSEEK_MODEL_PATH = os.path.join(BASE_DIR, "models", "ggml-model-Q4_0.gguf")

# Sentence Transformer 配置
# 使用本地缓存的模型路径,避免联网下载
# 指向 snapshots 下的具体版本目录
SENTENCE_TRANSFORMER_MODEL = os.path.join(
    BASE_DIR, "models", "embedding_cache", 
    "models--BAAI--bge-small-zh-v1.5", "snapshots", 
    "7999e1d3359715c523056ef9478215996d62a620"
)

# ChromaDB 配置
CHROMA_DB_PATH = "./chroma_db"

# LLM 配置
LLM_CONTEXT_LENGTH = 4096
LLM_MAX_TOKENS = 200  # 减少生成长度，提高速度（从 256 降到 150）
LLM_TEMPERATURE = 0.7

# LLM 性能优化
LLM_THREADS = 8  # CPU 线程数（根据您的 CPU 核心数调整）
LLM_BATCH_SIZE = 512  # 批处理大小

# RAG 配置
TOP_K_RESULTS = 3  # 检索最相似的文档数量
SIMILARITY_THRESHOLD = 0.5  # 相似度阈值（降低以包含更多相关文档）

# 性能配置
MAX_COLLECTION_SIZE = 10000  # 建议的最大知识库条目数
PERFORMANCE_WARNING_SIZE = 5000  # 性能警告阈值
BATCH_PROCESSING_SIZE = 100  # 批处理大小

# ==================== 模式控制 ====================
# RAG 模式开关
# True: 纯 RAG 模式，直接查询向量数据库回答，不进行意图识别和函数调用
# False: 智能模式，先识别用户意图，根据需要选择 RAG 或函数调用
PURE_RAG_MODE = True  

# Function Calling 配置（仅在 PURE_RAG_MODE = False 时生效）
ENABLE_FUNCTION_CALLING = True  # 是否启用函数调用功能
ENABLE_AI_PARAMETER_EXTRACTION = True  # 是否启用AI参数提取
AI_EXTRACTION_TEMPERATURE = 0.1  # AI参数提取的温度（低温度确保稳定性）
AI_EXTRACTION_MAX_TOKENS = 200  # AI参数提取的最大token数
DEBUG_FUNCTION_CALLING = True  # 是否显示详细的函数调用调试信息