# models/knowledge.py
from sqlalchemy import Column, String, Text, DateTime, Integer, Float, func
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
import config

Base = declarative_base()

class Knowledge(Base):
    """知识库元数据表"""
    __tablename__ = "knowledge"
    
    id = Column(String(36), primary_key=True, comment="知识库条目唯一ID (UUID)")
    content = Column(Text, nullable=False, comment="原始文本内容")
    source = Column(String(255), nullable=True, comment="知识来源")
    created_at = Column(DateTime, default=func.now(), comment="创建时间")
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now(), comment="更新时间")

class KnowledgeFeedback(Base):
    """知识反馈表 - 记录用户点赞的问答对"""
    __tablename__ = "knowledge_feedback"
    
    id = Column(String(36), primary_key=True, comment="反馈ID")
    question_text = Column(Text, nullable=False, comment="完整问题文本")
    question_hash = Column(String(64), nullable=False, unique=True, index=True, comment="问题特征哈希")
    answer_text = Column(Text, nullable=False, comment="AI回答内容")
    relevant_doc_ids = Column(Text, nullable=False, comment="相关文档ID列表（JSON数组）")
    like_count = Column(Integer, default=1, comment="点赞次数")
    use_count = Column(Integer, default=0, comment="被使用次数")
    success_rate = Column(Float, default=1.0, comment="成功率")
    created_at = Column(DateTime, default=func.now(), comment="创建时间")
    last_used_at = Column(DateTime, default=func.now(), comment="最后使用时间")
    updated_at = Column(DateTime, default=func.now(), onupdate=func.now(), comment="更新时间")

# 创建SQLite数据库异步引擎
sqlite_engine = create_async_engine(
    config.SQLITE_DATABASE_URL,
    echo=False,
    future=True,
)

# 创建session工厂
AsyncSQLiteSession = sessionmaker(
    sqlite_engine,
    class_=AsyncSession,
    expire_on_commit=False
)

# 获取SQLite session
async def get_sqlite_session() -> AsyncSession:
    async with AsyncSQLiteSession() as session:
        yield session

# 初始化数据库表
async def init_sqlite_db():
    """初始化SQLite数据库表"""
    async with sqlite_engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)





























