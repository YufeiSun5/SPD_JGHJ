# services/knowledge_service.py
import os
import uuid
import asyncio
from typing import List, Dict, Optional
import encoding_fix  # 添加导入

try:
    from sentence_transformers import SentenceTransformer
    SENTENCE_TRANSFORMERS_AVAILABLE = True
except ImportError:
    SENTENCE_TRANSFORMERS_AVAILABLE = False
    encoding_fix.safe_print("警告: sentence-transformers 未安装")

try:
    import chromadb
    from chromadb.config import Settings
    CHROMADB_AVAILABLE = True
except ImportError:
    CHROMADB_AVAILABLE = False
    encoding_fix.safe_print("警告: chromadb 未安装")

try:
    from llama_cpp import Llama
    LLAMA_CPP_AVAILABLE = True
except ImportError:
    LLAMA_CPP_AVAILABLE = False
    encoding_fix.safe_print("警告: llama-cpp-python 未安装")
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, delete
import config
from models.knowledge import Knowledge, AsyncSQLiteSession
try:
    import encoding_fix
except ImportError:
    # 如果encoding_fix不存在，创建一个简单的替代
    class EncodingFix:
        @staticmethod
        def safe_print(message):
            try:
                print(message)
            except UnicodeEncodeError:
                print(message.encode('utf-8', errors='ignore').decode('utf-8'))
    encoding_fix = EncodingFix()

# 导入函数调用相关模块
try:
    from services.ai_function_caller import AIFunctionCaller
    AI_FUNCTION_CALLING_AVAILABLE = True
except ImportError:
    AI_FUNCTION_CALLING_AVAILABLE = False
    encoding_fix.safe_print("警告: ai_function_caller 模块导入失败")

# 保持旧的导入以便兼容
try:
    from services.function_caller import FunctionCaller
    FUNCTION_CALLING_AVAILABLE = True
except ImportError:
    FUNCTION_CALLING_AVAILABLE = False

class KnowledgeService:
    """知识库服务类，集成向量搜索和LLM生成"""
    
    def __init__(self):
        self.sentence_model = None
        self.chroma_client = None
        self.chroma_collection = None
        self.feedback_collection = None  # 反馈问题向量集合
        self.llm_model = None
        self.initialized = False
        self.function_caller = None
        self.ai_function_caller = None
        self._init_lock = asyncio.Lock()  # 添加初始化锁
        self._query_queue = asyncio.Queue()  # 查询排队队列
        self._is_processing = False  # 是否正在处理查询
        self._current_query_id = None  # 当前处理的查询ID
        self._queue_processor_task = None  # 队列处理任务
    
    async def initialize(self):
        """异步初始化所有模型和数据库连接"""
        # 使用锁防止并发初始化
        async with self._init_lock:
            if self.initialized:
                return
            
            encoding_fix.safe_print("🚀 [初始化] 开始初始化知识库服务...")
            
            # 检查依赖是否可用
            missing_deps = []
            if not SENTENCE_TRANSFORMERS_AVAILABLE:
                missing_deps.append("sentence-transformers")
            if not CHROMADB_AVAILABLE:
                missing_deps.append("chromadb")
            if not LLAMA_CPP_AVAILABLE:
                missing_deps.append("llama-cpp-python")
            
            if missing_deps:
                encoding_fix.safe_print(f"[错误] 缺少依赖包: {', '.join(missing_deps)}")
                encoding_fix.safe_print("请运行: pip install sentence-transformers chromadb llama-cpp-python")
                self.initialized = False
                return
            
            # 在线程池中初始化计算密集型模型
            loop = asyncio.get_event_loop()
            await loop.run_in_executor(None, self._init_models)
            
            # 初始化新的AI Function Calling
            if AI_FUNCTION_CALLING_AVAILABLE:
                self.ai_function_caller = AIFunctionCaller(self.llm_model)
                encoding_fix.safe_print("✅ AI Function Calling 初始化完成 (AI驱动的函数选择和参数提取)")
            
            # 保持旧的函数调用器作为备用
            if FUNCTION_CALLING_AVAILABLE:
                # 根据配置决定是否启用AI参数提取
                llm_for_extraction = self.llm_model if config.ENABLE_AI_PARAMETER_EXTRACTION else None
                self.function_caller = FunctionCaller(llm_model=llm_for_extraction)
                
                if config.ENABLE_AI_PARAMETER_EXTRACTION:
                    encoding_fix.safe_print("✅ 传统函数调用器初始化完成 (启用AI参数提取)")
                else:
                    encoding_fix.safe_print("✅ 传统函数调用器初始化完成 (使用规则参数提取)")
            
            self.initialized = True
            encoding_fix.safe_print("✅ [初始化] 知识库服务初始化完成！")
            
            # 检查并同步 SQLite 和 ChromaDB
            await self._sync_databases()
            
            # 清理重复的反馈记录
            await self._cleanup_duplicate_feedbacks()
            
            # 启动排队处理任务
            self._queue_processor_task = asyncio.create_task(self._process_query_queue())
    
    def _init_models(self):
        """在线程池中初始化模型（避免阻塞事件循环）"""
        encoding_fix.safe_print("正在初始化 Sentence Transformer...")
        self.sentence_model = SentenceTransformer(config.SENTENCE_TRANSFORMER_MODEL)
        
        encoding_fix.safe_print("正在初始化 ChromaDB...")
        # 确保ChromaDB目录存在
        os.makedirs(config.CHROMA_DB_PATH, exist_ok=True)
        
        # 使用更稳定的ChromaDB配置
        try:
            self.chroma_client = chromadb.PersistentClient(
                path=config.CHROMA_DB_PATH,
                settings=Settings(
                    anonymized_telemetry=False,
                    allow_reset=True,
                    is_persistent=True
                )
            )
        except Exception as e:
            encoding_fix.safe_print(f"创建PersistentClient失败，尝试使用默认客户端: {e}")
            # 如果PersistentClient失败，尝试使用默认客户端
            self.chroma_client = chromadb.Client(
                Settings(
                    anonymized_telemetry=False,
                    chroma_db_impl="duckdb+parquet",
                    persist_directory=config.CHROMA_DB_PATH
                )
            )
        
        # 获取或创建知识库集合
        try:
            self.chroma_collection = self.chroma_client.get_or_create_collection(
                name="knowledge_base",
                metadata={"hnsw:space": "cosine"}
            )
        except Exception as e:
            encoding_fix.safe_print(f"创建集合失败，重试: {e}")
            # 如果集合创建失败，尝试重置并重新创建
            try:
                self.chroma_client.reset()
                self.chroma_collection = self.chroma_client.get_or_create_collection(
                    name="knowledge_base",
                    metadata={"hnsw:space": "cosine"}
                )
            except Exception as e2:
                encoding_fix.safe_print(f"重置后仍然失败: {e2}")
                raise e2
        
        # 创建反馈问题向量集合
        try:
            self.feedback_collection = self.chroma_client.get_or_create_collection(
                name="feedback_questions",
                metadata={"hnsw:space": "cosine"}
            )
            encoding_fix.safe_print("✅ 反馈问题向量集合初始化完成")
        except Exception as e:
            encoding_fix.safe_print(f"⚠️ 反馈集合创建失败: {e}")
            self.feedback_collection = None
        
        encoding_fix.safe_print("正在初始化 DeepSeek LLM...")
        # 检查模型文件是否存在
        if not os.path.exists(config.DEEPSEEK_MODEL_PATH):
            raise FileNotFoundError(f"模型文件不存在: {config.DEEPSEEK_MODEL_PATH}")
        
        # 设置环境变量确保UTF-8编码
        os.environ['PYTHONIOENCODING'] = 'utf-8'
        os.environ['PYTHONUTF8'] = '1'
        
        self.llm_model = Llama(
            model_path=config.DEEPSEEK_MODEL_PATH,
            n_ctx=config.LLM_CONTEXT_LENGTH,
            n_threads=getattr(config, 'LLM_THREADS', 8),  # 使用配置的线程数
            n_batch=getattr(config, 'LLM_BATCH_SIZE', 512),  # 批处理大小
            verbose=False,
            # 确保模型使用UTF-8编码
            encoding='utf-8'
        )
        
        encoding_fix.safe_print("所有模型初始化完成!")
    
    async def add_knowledge(self, content: str, source: str = None) -> str:
        """添加知识到知识库"""
        encoding_fix.safe_print(f"\n📚 [知识添加开始] 内容: {content[:100]}{'...' if len(content) > 100 else ''}")
        encoding_fix.safe_print(f"   📂 来源: {source or '未指定'}")
        
        if not self.initialized:
            await self.initialize()
        
        # 生成唯一ID
        knowledge_id = str(uuid.uuid4())
        encoding_fix.safe_print(f"   🆔 生成知识ID: {knowledge_id}")
        
        try:
            # 1. 在线程池中生成向量
            encoding_fix.safe_print("📊 [向量化] 正在生成知识向量嵌入...")
            loop = asyncio.get_event_loop()
            embedding = await loop.run_in_executor(
                None, 
                self.sentence_model.encode, 
                content
            )
            encoding_fix.safe_print(f"✅ [向量化] 完成，向量维度: {embedding.shape}")
            
            # 2. 存储到ChromaDB
            encoding_fix.safe_print("[数据库] [向量存储] 保存到ChromaDB向量数据库...")
            await loop.run_in_executor(
                None,
                lambda: self.chroma_collection.add(
                    embeddings=[embedding.tolist()],
                    ids=[knowledge_id],
                    metadatas=[{"content": content, "source": source or ""}]
                )
            )
            encoding_fix.safe_print("✅ [向量存储] ChromaDB存储完成")
            
            # 3. 存储元数据到SQLite
            encoding_fix.safe_print("💾 [元数据存储] 保存到SQLite数据库...")
            async with AsyncSQLiteSession() as session:
                knowledge = Knowledge(
                    id=knowledge_id,
                    content=content,
                    source=source
                )
                session.add(knowledge)
                await session.commit()
            encoding_fix.safe_print("✅ [元数据存储] SQLite存储完成")
            encoding_fix.safe_print(f"🎉 [知识添加完成] 知识已成功添加到知识库: {knowledge_id}")
            
            return knowledge_id
            
        except Exception as e:
            encoding_fix.safe_print(f"[错误] [知识添加错误] 添加过程中出现错误: {e}")
            encoding_fix.safe_print(f"   🔧 错误详情: {type(e).__name__}")
            # 如果出错，尝试清理已添加的数据
            encoding_fix.safe_print("🧹 [错误清理] 尝试清理已添加的数据...")
            try:
                loop = asyncio.get_event_loop()
                await loop.run_in_executor(
                    None,
                    lambda: self.chroma_collection.delete(ids=[knowledge_id])
                )
                encoding_fix.safe_print("✅ [错误清理] 清理完成")
            except Exception as cleanup_error:
                encoding_fix.safe_print(f"[警告] [错误清理] 清理失败: {cleanup_error}")
            raise e
    
    async def delete_knowledge(self, knowledge_id: str) -> bool:
        """删除指定ID的知识"""
        if not self.initialized:
            await self.initialize()
        
        try:
            # 1. 从ChromaDB删除
            loop = asyncio.get_event_loop()
            await loop.run_in_executor(
                None,
                lambda: self.chroma_collection.delete(ids=[knowledge_id])
            )
            
            # 2. 从SQLite删除
            async with AsyncSQLiteSession() as session:
                stmt = delete(Knowledge).where(Knowledge.id == knowledge_id)
                result = await session.execute(stmt)
                await session.commit()
                
                return result.rowcount > 0
                
        except Exception as e:
            encoding_fix.safe_print(f"删除知识时出错: {e}")
            return False
    
    async def delete_knowledge_by_source(self, source: str) -> int:
        """删除指定来源的所有知识（用于替换式导入）
        
        Args:
            source: 知识来源
            
        Returns:
            删除的知识数量
        """
        if not self.initialized:
            await self.initialize()
        
        try:
            encoding_fix.safe_print(f"🔍 [删除知识] 开始查询来源为 '{source}' 的知识...")
            
            # 1. 从SQLite查询该来源的所有知识ID（使用同一个会话）
            async with AsyncSQLiteSession() as session:
                from sqlalchemy import select
                
                # 先查询所有知识，调试用
                all_stmt = select(Knowledge.id, Knowledge.source)
                all_result = await session.execute(all_stmt)
                all_knowledge = all_result.fetchall()
                encoding_fix.safe_print(f"📊 [调试] 数据库中共有 {len(all_knowledge)} 条知识")
                for kid, ksource in all_knowledge[:5]:  # 只显示前5条
                    encoding_fix.safe_print(f"   - ID: {kid[:8]}..., Source: '{ksource}'")
                
                # 查询指定来源的知识
                stmt = select(Knowledge.id).where(Knowledge.source == source)
                result = await session.execute(stmt)
                knowledge_ids = [row[0] for row in result.fetchall()]
                
                encoding_fix.safe_print(f"🔍 [查询结果] 找到 {len(knowledge_ids)} 条来源为 '{source}' 的知识")
                
                if not knowledge_ids:
                    encoding_fix.safe_print(f"📭 [删除知识] 来源 '{source}' 没有找到任何知识，跳过删除")
                    return 0
                
                encoding_fix.safe_print(f"🗑️ [删除知识] 准备删除 {len(knowledge_ids)} 条知识...")
                for kid in knowledge_ids:
                    encoding_fix.safe_print(f"   - 将删除: {kid}")
                
                # 2. 从ChromaDB批量删除
                encoding_fix.safe_print(f"🗑️ [ChromaDB] 开始删除向量数据...")
                loop = asyncio.get_event_loop()
                await loop.run_in_executor(
                    None,
                    lambda: self.chroma_collection.delete(ids=knowledge_ids)
                )
                encoding_fix.safe_print(f"✅ [ChromaDB] 删除完成")
                
                # 3. 从SQLite删除（在同一个会话中）
                encoding_fix.safe_print(f"🗑️ [SQLite] 开始删除元数据...")
                delete_stmt = delete(Knowledge).where(Knowledge.source == source)
                delete_result = await session.execute(delete_stmt)
                await session.commit()
                deleted_count = delete_result.rowcount
                
                encoding_fix.safe_print(f"✅ [SQLite] 删除完成，共删除 {deleted_count} 条记录")
            
            encoding_fix.safe_print(f"🎉 [删除知识] 成功删除来源 '{source}' 的 {deleted_count} 条知识")
            return deleted_count
                
        except Exception as e:
            encoding_fix.safe_print(f"❌ [删除知识] 按来源删除知识时出错: {e}")
            import traceback
            encoding_fix.safe_print(f"   错误堆栈: {traceback.format_exc()}")
            return 0
    
    async def query_knowledge(self, question: str) -> Dict:
        """查询知识库并生成回答，支持function calling"""
        # 安全处理用户输入的问题，避免编码错误
        safe_question = encoding_fix.safe_encode_for_model(question)
        encoding_fix.safe_print(f"\n🔍 [AI处理开始] 收到用户问题: {safe_question}")
        
        if not self.initialized:
            await self.initialize()
        
        # 🚀 首先尝试新的AI Function Calling
        if AI_FUNCTION_CALLING_AVAILABLE and self.ai_function_caller:
            try:
                encoding_fix.safe_print("🤖 [AI Function Calling] 让AI分析问题并决定是否调用函数...")
                ai_function_result = await self.ai_function_caller.analyze_and_call(safe_question)
                
                if ai_function_result.get("needs_function_call") and ai_function_result.get("function_called"):
                    # AI成功选择并执行了函数
                    function_result = ai_function_result.get("result")
                    function_name = ai_function_result.get("function_name")
                    arguments = ai_function_result.get("arguments")
                    reasoning = ai_function_result.get("reasoning", "")
                    
                    encoding_fix.safe_print(f"✅ [AI Function Calling] AI成功执行函数: {function_name}")
                    encoding_fix.safe_print(f"📝 [AI推理] {reasoning}")
                    
                    # 格式化数据库查询结果
                    formatted_result = self.ai_function_caller.format_database_result(function_result)
                    
                    # 使用LLM生成基于数据库结果的回答
                    context = f"数据库查询结果：\n{formatted_result}"
                    prompt = self._build_function_result_prompt(safe_question, context)
                    
                    response = await asyncio.get_event_loop().run_in_executor(
                        None,
                        self._generate_response,
                        prompt
                    )
                    
                    return {
                        'answer': response,
                        'relevant_docs': [],
                        'context_used': True,
                        'function_called': True,
                        'function_result': function_result,
                        'database_data': formatted_result,
                        'ai_reasoning': reasoning,
                        'function_name': function_name,
                        'function_arguments': arguments
                    }
                
                elif ai_function_result.get("needs_function_call") and not ai_function_result.get("function_called"):
                    # AI认为需要函数调用但执行失败
                    error_msg = ai_function_result.get("error", "函数执行失败")
                    reasoning = ai_function_result.get("reasoning", "")
                    
                    encoding_fix.safe_print(f"❌ [AI Function Calling] 函数执行失败: {error_msg}")
                    
                    # 生成错误回答
                    prompt = f"""用户问题: {safe_question}
                    
AI分析: {reasoning}
尝试调用数据库函数但失败了，错误: {error_msg}

请向用户说明无法查询到相关数据，并建议他们检查查询条件或稍后重试。"""
                    
                    response = await asyncio.get_event_loop().run_in_executor(
                        None,
                        self._generate_response,
                        prompt
                    )
                    
                    return {
                        'answer': response,
                        'relevant_docs': [],
                        'context_used': False,
                        'function_called': False,
                        'error': error_msg,
                        'ai_reasoning': reasoning
                    }
                else:
                    # AI认为不需要函数调用，继续知识库查询
                    reasoning = ai_function_result.get("reasoning", "")
                    encoding_fix.safe_print(f"💭 [AI Function Calling] AI判断不需要函数调用: {reasoning}")
                    
            except Exception as e:
                encoding_fix.safe_print(f"[警告] AI Function Calling失败，回退到传统方式: {e}")
                
                # 🔄 回退到传统Function Calling
                if FUNCTION_CALLING_AVAILABLE and self.function_caller:
                    try:
                        function_call_result = await self.function_caller.analyze_and_call(safe_question)
                        
                        if function_call_result.get("needs_function_call") and function_call_result.get("function_called"):
                            function_result = function_call_result.get("result")
                            encoding_fix.safe_print(f"✅ [传统Function Calling] 成功执行数据库查询")
                            
                            formatted_result = self.function_caller.format_database_result(function_result)
                            context = f"数据库查询结果：\n{formatted_result}"
                            prompt = self._build_function_result_prompt(safe_question, context)
                            
                            response = await asyncio.get_event_loop().run_in_executor(
                                None,
                                self._generate_response,
                                prompt
                            )
                            
                            return {
                                'answer': response,
                                'relevant_docs': [],
                                'context_used': True,
                                'function_called': True,
                                'function_result': function_result,
                                'database_data': formatted_result,
                                'fallback_mode': True
                            }
                            
                    except Exception as fallback_error:
                        encoding_fix.safe_print(f"[警告] 传统Function Calling也失败，回退到知识库查询: {fallback_error}")
        
        try:
            # 1. 将问题转换为向量
            encoding_fix.safe_print("📊 [向量化] 正在将问题转换为向量嵌入...")
            loop = asyncio.get_event_loop()
            question_embedding = await loop.run_in_executor(
                None,
                self.sentence_model.encode,
                safe_question
            )
            encoding_fix.safe_print(f"✅ [向量化] 完成，向量维度: {question_embedding.shape}")
            
            # 2. 在ChromaDB中搜索相似文档
            import time
            
            # 检查知识库大小并给出性能提示
            collection_count = self.chroma_collection.count()
            encoding_fix.safe_print(f"📊 [知识库状态] 当前知识库大小: {collection_count} 条")
            
            if collection_count > config.PERFORMANCE_WARNING_SIZE:
                encoding_fix.safe_print(f"[警告]  [性能提醒] 知识库较大({collection_count} > {config.PERFORMANCE_WARNING_SIZE})，检索可能需要更多时间")
            
            encoding_fix.safe_print(f"🔎 [向量检索] 在知识库中搜索相似文档，检索数量: {config.TOP_K_RESULTS}")
            
            # 记录检索开始时间
            search_start_time = time.time()
            
            search_results = await loop.run_in_executor(
                None,
                lambda: self.chroma_collection.query(
                    query_embeddings=[question_embedding.tolist()],
                    n_results=config.TOP_K_RESULTS,
                    include=["metadatas", "distances"]
                )
            )
            
            # 计算检索耗时
            search_time = time.time() - search_start_time
            encoding_fix.safe_print(f"🔍 [向量检索] 完成，耗时: {search_time:.2f}秒，找到 {len(search_results['ids'][0]) if search_results['ids'] and len(search_results['ids']) > 0 else 0} 个候选文档")
            
            # 性能分析
            if search_time > 2.0:
                encoding_fix.safe_print(f"🐌 [性能警告] 检索耗时较长({search_time:.2f}s)，建议考虑优化")
            elif search_time > 1.0:
                encoding_fix.safe_print(f"[计时]  [性能提示] 检索耗时适中({search_time:.2f}s)")
            else:
                encoding_fix.safe_print(f"⚡ [性能优秀] 检索速度很快({search_time:.2f}s)")
            
            # 3. 获取相关文档的详细信息
            encoding_fix.safe_print("📋 [相似度过滤] 分析文档相似度...")
            relevant_docs = []
            if search_results['ids'] and len(search_results['ids'][0]) > 0:
                doc_ids = search_results['ids'][0]
                distances = search_results['distances'][0]
                metadatas = search_results['metadatas'][0]
                
                # 根据相似度阈值过滤
                for i, (doc_id, distance, metadata) in enumerate(zip(doc_ids, distances, metadatas)):
                    similarity = 1 - distance  # 将距离转换为相似度
                    encoding_fix.safe_print(f"   [文档] 文档 {i+1}: 相似度 {similarity:.3f} ({'[通过]' if similarity >= config.SIMILARITY_THRESHOLD else '[过滤]'})")
                    if similarity >= config.SIMILARITY_THRESHOLD:
                        relevant_docs.append({
                            'id': doc_id,
                            'content': metadata['content'],
                            'source': metadata['source'],
                            'similarity': similarity
                        })
                        encoding_fix.safe_print(f"      📝 内容预览: {metadata['content'][:50]}...")
            
            encoding_fix.safe_print(f"✅ [相似度过滤] 完成，有效文档数: {len(relevant_docs)}")
            
            # 4. 构建上下文和prompt
            if relevant_docs:
                context = "\n\n".join([doc['content'] for doc in relevant_docs])
                encoding_fix.safe_print("📝 [上下文构建] 基于相关文档构建上下文")
                encoding_fix.safe_print(f"   📊 上下文长度: {len(context)} 字符")
                prompt = self._build_prompt(safe_question, context)
            else:
                encoding_fix.safe_print("📝 [上下文构建] 无相关文档，使用通用模式")
                prompt = self._build_prompt(safe_question, "")
            
            encoding_fix.safe_print(f"[机器人] [LLM调用] 调用大语言模型生成回答...")
            encoding_fix.safe_print(f"   [设置] 模型参数: max_tokens={config.LLM_MAX_TOKENS}, temperature={config.LLM_TEMPERATURE}")
            
            # 5. 使用LLM生成回答
            response = await loop.run_in_executor(
                None,
                self._generate_response,
                prompt
            )
            
            encoding_fix.safe_print(f"🧠 [LLM调用] 回答生成完成，长度: {len(response)} 字符")
            encoding_fix.safe_print(f"🎯 [AI处理完成] 最终回答: {response[:100]}{'...' if len(response) > 100 else ''}")
            
            return {
                'answer': response,
                'relevant_docs': relevant_docs,
                'context_used': len(relevant_docs) > 0
            }
            
        except Exception as e:
            encoding_fix.safe_print(f"[错误] [AI处理错误] 查询知识库时出错: {e}")
            encoding_fix.safe_print(f"   🔧 错误详情: {type(e).__name__}")
            return {
                'answer': '抱歉，查询过程中出现错误，请稍后重试。',
                'relevant_docs': [],
                'context_used': False,
                'error': str(e)
            }
    
    def _build_prompt(self, question: str, context: str) -> str:
        """构建用于LLM的prompt"""
        # 系统身份设定
        identity = """你是经验丰富的工业设备维修老师傅，说话简洁直接但有人情味，像在车间面对面指导徒弟。"""
        
        if context:
            return f"""{identity}

参考信息：
{context}

用户问题：{question}

回答风格要求：
1. 不要重复用户的问题，直接说"这个情况..."或"遇到这个..."就行
2. 不要用"问题、解决方案、预防措施、操作建议"这种标题格式
3. 像老师傅说话一样：先简单说怎么回事，然后重点说怎么处理，最后提醒怎么预防
4. 用简短的句子，说人话，但要专业、实际、接地气
5. 只说跟设备维修相关的，不要扯什么"喝酒、说话"这种无关的东西
6. 控制在120字左右，简洁有效
7. 回答完就停止，不要继续输出其他错误码信息

回答："""
        else:
            return f"""{identity}

用户问题：{question}

注意：知识库中暂未找到相关知识，以下回答基于大连盛云科技智能助理本身的知识（截止到2024年3月的训练数据）。

回答风格：像老师傅说话，简洁专业但不冷冰冰。不确定就直说，别瞎编。

回答："""
    
    def _build_function_result_prompt(self, question: str, database_result: str) -> str:
        """构建基于数据库查询结果的prompt（优化版：更简洁）"""
        return f"""你是数据分析助手。基于以下数据回答用户问题。

用户问题：{question}

查询结果：
{database_result}

要求：
1. 直接回答核心解決方案，提供关键数据
2. 简洁分析趋势或异常（如有）
3. 给出1-2条实用建议
4. 控制在150字以内

回答："""
    
    def _build_parameter_request_prompt(self, question: str, function_name: str, missing_params: List[str]) -> str:
        """构建请求补充参数的prompt"""
        identity = """你是盛云知识问答小助手，专门为用户提供智能问答服务。"""
        
        param_descriptions = {
            'date': '具体日期（格式：YYYY-MM-DD，如：2024-01-15）',
            'start_date': '开始日期（格式：YYYY-MM-DD）',
            'end_date': '结束日期（格式：YYYY-MM-DD）',
            'model_name': '机型名称（如：iPhone15、SamsungS23等）'
        }
        
        missing_desc = []
        for param in missing_params:
            desc = param_descriptions.get(param, param)
            missing_desc.append(f"• {desc}")
        
        return f"""{identity}

我理解您想要查询数据库获取相关信息，但需要您提供一些具体的参数才能完成查询。

您的问题：{question}

需要补充的信息：
{chr(10).join(missing_desc)}

请提供这些信息后，我就可以为您查询准确的数据了。

回答："""
    
    def _generate_response(self, prompt: str) -> str:
        """使用LLM生成回答"""
        try:
            # 安全处理prompt中的特殊字符
            safe_prompt = self._safe_encode_text(prompt)
            
            response = self.llm_model(
                safe_prompt,
                max_tokens=config.LLM_MAX_TOKENS,
                temperature=config.LLM_TEMPERATURE,
                stop=["</s>", "\n用户问题：", "\n用户问题:", "用户问题：", "用户问题:", "\n\n用户", "\n知识库信息：", "\n回答：", "\n【错误码", "\n错误码", "【错误码"],
                echo=False
            )
            
            answer = response['choices'][0]['text'].strip()
            return answer if answer else "抱歉，我无法生成合适的回答。"
            
        except Exception as e:
            encoding_fix.safe_print(f"LLM生成回答时出错: {e}")
            return "抱歉，生成回答时出现错误。"
    
    def _safe_encode_text(self, text: str) -> str:
        """安全处理文本中的特殊字符，避免编码错误"""
        return encoding_fix.safe_encode_for_model(text)
    
    async def query_knowledge_stream(self, question: str) -> Dict:
        """查询知识库并生成流式回答，支持function calling"""
        # 安全处理用户输入的问题，避免编码错误
        safe_question = encoding_fix.safe_encode_for_model(question)
        encoding_fix.safe_print(f"\n🌊 [流式AI处理开始] 收到用户问题: {safe_question}")
        
        if not self.initialized:
            await self.initialize()
        
        # 检查是否启用纯 RAG 模式
        if config.PURE_RAG_MODE:
            encoding_fix.safe_print(f"📚 [纯 RAG 模式] 跳过函数调用，直接使用向量检索")
            # 跳过函数调用逻辑，直接进行 RAG 检索
            pass
        # 首先尝试新的AI Function Calling（仅在非纯RAG模式下）
        elif AI_FUNCTION_CALLING_AVAILABLE and self.ai_function_caller:
            try:
                print(f"🤖 [AI Function Calling] 开始分析用户问题...")  # 强制打印
                ai_function_result = await self.ai_function_caller.analyze_and_call(safe_question)
                
                # 检查是否尝试调用了不存在的函数
                if ai_function_result.get("needs_function_call") and not ai_function_result.get("function_called"):
                    error_msg = ai_function_result.get("error", "")
                    if "不存在" in error_msg or "available_functions" in ai_function_result:
                        # AI 选择了不存在的函数，给出提示并回退到 RAG
                        encoding_fix.safe_print(f"⚠️ [AI Function Calling] {error_msg}")
                        encoding_fix.safe_print(f"📚 [回退] 该查询超出可用函数范围，使用知识库回答")
                        # 不返回，继续执行下面的 RAG 流程
                    
                elif ai_function_result.get("needs_function_call") and ai_function_result.get("function_called"):
                    # 成功调用了数据库查询函数
                    function_result = ai_function_result.get("result")
                    print(f"✅ [AI Function Calling] 成功执行数据库查询")  # 强制打印
                    
                    # 格式化数据库查询结果
                    print(f"🔍 [调试] function_result内容: {function_result}")  # 调试信息
                    formatted_result = self.ai_function_caller.format_database_result(function_result)
                    print(f"🔍 [调试] formatted_result: {formatted_result}")  # 调试信息
                    
                    # 使用LLM生成基于数据库结果的流式回答
                    context = f"数据库查询结果：\n{formatted_result}"
                    prompt = self._build_function_result_prompt(safe_question, context)
                    
                    # 创建流式生成器
                    async def generate_function_result_stream():
                        try:
                            safe_prompt = self._safe_encode_text(prompt)
                            token_count = 0
                            
                            for chunk in self.llm_model(
                                safe_prompt,
                                max_tokens=config.LLM_MAX_TOKENS,
                                temperature=config.LLM_TEMPERATURE,
                                stop=["</s>", "\n用户问题：", "\n用户问题:", "用户问题：", "用户问题:", "\n\n用户", "\n知识库信息：", "\n回答：", "\n【错误码", "\n错误码", "【错误码"],
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
                            
                            encoding_fix.safe_print(f"✅ [流式Function Result] 生成完成，总计 {token_count} 个token")
                        except Exception as e:
                            encoding_fix.safe_print(f"[错误] [流式Function Result错误] {e}")
                            yield f"抱歉，生成回答时出现错误: {str(e)}"
                    
                    return {
                        'relevant_docs': [],
                        'answer_stream': generate_function_result_stream(),
                        'function_called': True,
                        'function_result': function_result,
                        'database_data': formatted_result
                    }
                
                elif ai_function_result.get("query_type") == "incomplete_database_query":
                    # 需要数据库查询但参数不完整
                    missing_params = ai_function_result.get("missing_parameters", [])
                    function_name = ai_function_result.get("function_name", "")
                    
                    # 生成参数请求的流式回答
                    async def generate_parameter_request_stream():
                        prompt = self._build_parameter_request_prompt(safe_question, function_name, missing_params)
                        try:
                            safe_prompt = self._safe_encode_text(prompt)
                            for chunk in self.llm_model(
                                safe_prompt,
                                max_tokens=config.LLM_MAX_TOKENS,
                                temperature=config.LLM_TEMPERATURE,
                                stop=["</s>", "\n用户问题：", "\n用户问题:", "用户问题：", "用户问题:", "\n\n用户", "\n知识库信息：", "\n回答：", "\n【错误码", "\n错误码", "【错误码"],
                                stream=True
                            ):
                                if 'choices' in chunk and len(chunk['choices']) > 0:
                                    delta = chunk['choices'][0].get('delta', {})
                                    if 'content' in delta:
                                        yield delta['content']
                                    elif 'text' in chunk['choices'][0]:
                                        yield chunk['choices'][0]['text']
                        except Exception as e:
                            yield f"抱歉，生成回答时出现错误: {str(e)}"
                    
                    return {
                        'relevant_docs': [],
                        'answer_stream': generate_parameter_request_stream(),
                        'function_called': False,
                        'needs_more_info': True,
                        'missing_parameters': missing_params
                    }
                    
            except Exception as e:
                print(f"❌ [AI Function Calling] 失败: {e}")  # 强制打印
                encoding_fix.safe_print(f"[警告] AI Function calling失败，尝试传统Function calling: {e}")
        
        # 如果新的AI Function Calling不可用，尝试旧的function_caller作为备用
        elif FUNCTION_CALLING_AVAILABLE and self.function_caller:
            try:
                print(f"🔄 [传统Function Calling] 开始分析用户问题...")  # 强制打印
                function_call_result = await self.function_caller.analyze_and_call(safe_question)
                
                if function_call_result.get("needs_function_call") and function_call_result.get("function_called"):
                    # 成功调用了数据库查询函数
                    function_result = function_call_result.get("result")
                    print(f"✅ [传统Function Calling] 成功执行数据库查询")  # 强制打印
                    
                    # 格式化数据库查询结果
                    formatted_result = self.function_caller.format_database_result(function_result)
                    
                    # 使用LLM生成基于数据库结果的流式回答
                    context = f"数据库查询结果：\n{formatted_result}"
                    prompt = self._build_function_result_prompt(safe_question, context)
                    
                    # 创建流式生成器
                    async def generate_function_result_stream():
                        try:
                            safe_prompt = self._safe_encode_text(prompt)
                            token_count = 0
                            
                            for chunk in self.llm_model(
                                safe_prompt,
                                max_tokens=config.LLM_MAX_TOKENS,
                                temperature=config.LLM_TEMPERATURE,
                                stop=["</s>", "\n用户问题：", "\n用户问题:", "用户问题：", "用户问题:", "\n\n用户", "\n知识库信息：", "\n回答：", "\n【错误码", "\n错误码", "【错误码"],
                                stream=True
                            ):
                                if 'choices' in chunk and len(chunk['choices']) > 0:
                                    delta = chunk['choices'][0].get('delta', {})
                                    if 'content' in delta:
                                        token_count += 1
                                        if token_count % 10 == 0:
                                            encoding_fix.safe_print(f"   🔄 已生成 {token_count} 个token...")
                                        yield delta['content']
                                    elif 'text' in chunk['choices'][0]:
                                        token_count += 1
                                        if token_count % 10 == 0:
                                            encoding_fix.safe_print(f"   🔄 已生成 {token_count} 个token...")
                                        yield chunk['choices'][0]['text']
                            
                            encoding_fix.safe_print(f"✅ [流式LLM] 生成完成，总计 {token_count} 个token")
                        except Exception as e:
                            yield f"抱歉，生成回答时出现错误: {str(e)}"
                    
                    return {
                        'relevant_docs': [],
                        'context_used': True,
                        'stream_generator': generate_function_result_stream(),
                        'function_called': True,
                        'function_result': function_result,
                        'database_data': formatted_result
                    }
                
                elif function_call_result.get("query_type") == "incomplete_database_query":
                    # 需要数据库查询但参数不完整
                    missing_params = function_call_result.get("missing_parameters", [])
                    function_name = function_call_result.get("function_name", "")
                    
                    # 生成参数请求的流式回答
                    async def generate_parameter_request_stream():
                        prompt = self._build_parameter_request_prompt(safe_question, function_name, missing_params)
                        try:
                            safe_prompt = self._safe_encode_text(prompt)
                            for chunk in self.llm_model(
                                safe_prompt,
                                max_tokens=config.LLM_MAX_TOKENS,
                                temperature=config.LLM_TEMPERATURE,
                                stop=["</s>", "\n用户问题：", "\n用户问题:", "用户问题：", "用户问题:", "\n\n用户", "\n知识库信息：", "\n回答：", "\n【错误码", "\n错误码", "【错误码"],
                                stream=True
                            ):
                                if 'choices' in chunk and len(chunk['choices']) > 0:
                                    delta = chunk['choices'][0].get('delta', {})
                                    if 'content' in delta:
                                        yield delta['content']
                                    elif 'text' in chunk['choices'][0]:
                                        yield chunk['choices'][0]['text']
                        except Exception as e:
                            yield f"抱歉，生成回答时出现错误: {str(e)}"
                    
                    return {
                        'relevant_docs': [],
                        'context_used': False,
                        'stream_generator': generate_parameter_request_stream(),
                        'function_called': False,
                        'needs_more_info': True,
                        'missing_parameters': missing_params
                    }
                    
            except Exception as e:
                print(f"❌ [传统Function Calling] 也失败了: {e}")  # 强制打印
                encoding_fix.safe_print(f"[警告] 传统Function calling也失败，回退到知识库查询: {e}")
        
        try:
            # 1. 将问题转换为向量
            loop = asyncio.get_event_loop()
            question_embedding = await loop.run_in_executor(
                None,
                self.sentence_model.encode,
                safe_question
            )
            
            # 2. 在ChromaDB中搜索相似文档
            encoding_fix.safe_print(f"🔎 [向量检索] 在知识库中搜索相似文档，检索数量: {config.TOP_K_RESULTS}")
            search_results = await loop.run_in_executor(
                None,
                lambda: self.chroma_collection.query(
                    query_embeddings=[question_embedding.tolist()],
                    n_results=config.TOP_K_RESULTS,
                    include=["metadatas", "distances"]  # 只需要 metadatas 和 distances
                )
            )
            
            encoding_fix.safe_print(f"🔍 [向量检索] 完成，找到 {len(search_results['ids'][0]) if search_results['ids'] and len(search_results['ids']) > 0 else 0} 个候选文档")
            
            # 3. 获取相关文档的详细信息
            encoding_fix.safe_print("📋 [相似度过滤] 分析文档相似度...")
            relevant_docs = []
            if search_results['ids'] and len(search_results['ids'][0]) > 0:
                doc_ids = search_results['ids'][0]
                distances = search_results['distances'][0]
                metadatas = search_results['metadatas'][0]
                
                # 根据相似度阈值过滤
                for i, (doc_id, distance, metadata) in enumerate(zip(doc_ids, distances, metadatas)):
                    similarity = 1 - distance  # 将距离转换为相似度
                    encoding_fix.safe_print(f"   [文档] 文档 {i+1}: 相似度 {similarity:.3f} ({'✅通过' if similarity >= config.SIMILARITY_THRESHOLD else '❌过滤'})")
                    if similarity >= config.SIMILARITY_THRESHOLD:
                        # 从 metadata 中获取内容,因为我们存储时就是放在 metadata 里的
                        content = metadata.get('content', '')
                        if content:  # 确保内容不为空
                            relevant_docs.append({
                                'id': doc_id,
                                'content': content,
                                'source': metadata.get('source', '未知来源'),
                                'similarity': similarity
                            })
                            encoding_fix.safe_print(f"      📝 内容预览: {content[:50]}...")
                        else:
                            encoding_fix.safe_print(f"      ⚠️ 警告: 文档内容为空")
            
            encoding_fix.safe_print(f"✅ [相似度过滤] 完成，有效文档数: {len(relevant_docs)}")
            
            # 4. 构建上下文和prompt  
            if relevant_docs:
                context = "\n\n".join([doc['content'] for doc in relevant_docs])
                prompt = self._build_prompt(safe_question, context)
            else:
                prompt = self._build_prompt(safe_question, "")
            
            # 5. 创建流式生成器
            encoding_fix.safe_print(f"🌊 [流式LLM] 开始流式生成回答...")
            async def generate_answer_stream():
                try:
                    token_count = 0
                    # 安全处理prompt中的特殊字符
                    safe_prompt = self._safe_encode_text(prompt)
                    
                    # 使用LLM流式生成回答
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
                                if token_count % 10 == 0:  # 每10个token输出一次日志
                                    encoding_fix.safe_print(f"   🔄 已生成 {token_count} 个token...")
                                yield delta['content']
                            elif 'text' in chunk['choices'][0]:
                                token_count += 1
                                if token_count % 10 == 0:
                                    encoding_fix.safe_print(f"   🔄 已生成 {token_count} 个token...")
                                yield chunk['choices'][0]['text']
                    encoding_fix.safe_print(f"✅ [流式LLM] 生成完成，总计 {token_count} 个token")
                except GeneratorExit:
                    # 生成器被关闭（客户端断开）
                    encoding_fix.safe_print(f"🛑 [流式LLM] 客户端断开连接，已生成 {token_count} 个token，停止生成")
                    return
                except Exception as e:
                    encoding_fix.safe_print(f"[错误] [流式生成错误] {e}")
                    yield f"抱歉，生成回答时出现错误: {str(e)}"
            
            return {
                'relevant_docs': relevant_docs,
                'answer_stream': generate_answer_stream()
            }
            
        except Exception as e:
            encoding_fix.safe_print(f"查询知识库时出错: {e}")
            
            async def error_stream():
                yield f"抱歉，查询过程中出现错误: {str(e)}"
            
            return {
                'relevant_docs': [],
                'answer_stream': error_stream()
            }
    
    async def get_knowledge_list(self, limit: int = 100, offset: int = 0) -> List[Dict]:
        """获取知识库列表"""
        async with AsyncSQLiteSession() as session:
            stmt = select(Knowledge).limit(limit).offset(offset).order_by(Knowledge.created_at.desc())
            result = await session.execute(stmt)
            knowledge_list = result.scalars().all()
            
            return [
                {
                    'id': k.id,
                    'content': k.content[:200] + '...' if len(k.content) > 200 else k.content,
                    'source': k.source,
                    'created_at': k.created_at.isoformat() if k.created_at else None
                }
                for k in knowledge_list
            ]
    
    async def get_performance_stats(self) -> Dict:
        """获取知识库性能统计信息"""
        if not self.initialized:
            await self.initialize()
        
        try:
            # 获取知识库大小
            collection_count = self.chroma_collection.count()
            
            # 计算性能等级
            if collection_count < 1000:
                performance_level = "优秀"
                performance_color = "green"
                estimated_response_time = "0.5-1秒"
            elif collection_count < 5000:
                performance_level = "良好"
                performance_color = "blue"
                estimated_response_time = "1-2秒"
            elif collection_count < 10000:
                performance_level = "一般"
                performance_color = "orange"
                estimated_response_time = "2-4秒"
            else:
                performance_level = "较慢"
                performance_color = "red"
                estimated_response_time = "4-8秒"
            
            # 获取数据库文件大小（如果可能）
            try:
                import os
                db_size = os.path.getsize("./chroma_db") if os.path.exists("./chroma_db") else 0
                db_size_mb = db_size / (1024 * 1024)
            except:
                db_size_mb = 0
            
            return {
                "knowledge_count": collection_count,
                "performance_level": performance_level,
                "performance_color": performance_color,
                "estimated_response_time": estimated_response_time,
                "database_size_mb": round(db_size_mb, 2),
                "recommendations": self._get_performance_recommendations(collection_count)
            }
            
        except Exception as e:
            encoding_fix.safe_print(f"获取性能统计时出错: {e}")
            return {
                "knowledge_count": 0,
                "performance_level": "未知",
                "performance_color": "gray",
                "estimated_response_time": "未知",
                "database_size_mb": 0,
                "recommendations": []
            }
    
    def _get_performance_recommendations(self, collection_count: int) -> List[str]:
        """根据知识库大小给出性能建议"""
        recommendations = []
        
        if collection_count > 10000:
            recommendations.extend([
                "建议定期清理不相关的知识条目",
                "考虑按主题分类建立多个知识库",
                "可以提高相似度阈值以减少检索范围"
            ])
        elif collection_count > 5000:
            recommendations.extend([
                "性能仍在可接受范围，建议监控响应时间",
                "可以考虑优化知识条目的长度和质量"
            ])
        elif collection_count > 1000:
            recommendations.append("性能良好，可以继续添加知识")
        else:
            recommendations.append("知识库较小，响应速度很快")
        
        return recommendations
    
    async def _sync_databases(self):
        """检查并同步 SQLite 和 ChromaDB 数据"""
        try:
            encoding_fix.safe_print("\n🔄 [数据同步] 检查数据库一致性...")
            
            # 1. 获取 SQLite 中的知识数量
            async with AsyncSQLiteSession() as session:
                result = await session.execute(select(Knowledge))
                sqlite_items = result.scalars().all()
                sqlite_count = len(sqlite_items)
            
            # 2. 获取 ChromaDB 中的文档数量
            loop = asyncio.get_event_loop()
            chroma_count = await loop.run_in_executor(
                None,
                lambda: self.chroma_collection.count()
            )
            
            encoding_fix.safe_print(f"   📊 SQLite 知识数: {sqlite_count}")
            encoding_fix.safe_print(f"   📊 ChromaDB 文档数: {chroma_count}")
            
            # 3. 如果数量不一致，进行同步
            if sqlite_count != chroma_count:
                encoding_fix.safe_print(f"⚠️ [数据不一致] 检测到数据不同步！")
                encoding_fix.safe_print(f"   SQLite: {sqlite_count} 条")
                encoding_fix.safe_print(f"   ChromaDB: {chroma_count} 条")
                encoding_fix.safe_print(f"🔄 [自动修复] 开始从 SQLite 同步到 ChromaDB...")
                
                # 获取 ChromaDB 中已存在的 ID
                existing_ids = set()
                if chroma_count > 0:
                    try:
                        chroma_data = await loop.run_in_executor(
                            None,
                            lambda: self.chroma_collection.get()
                        )
                        existing_ids = set(chroma_data.get('ids', []))
                        encoding_fix.safe_print(f"   📋 ChromaDB 中已有 {len(existing_ids)} 个ID")
                    except Exception as e:
                        encoding_fix.safe_print(f"   ⚠️ 获取现有ID失败: {e}")
                
                # 同步缺失的数据
                synced_count = 0
                skipped_count = 0
                error_count = 0
                
                for item in sqlite_items:
                    try:
                        # 如果 ChromaDB 中已存在，跳过
                        if item.id in existing_ids:
                            skipped_count += 1
                            continue
                        
                        # 生成向量
                        embedding = await loop.run_in_executor(
                            None,
                            self.sentence_model.encode,
                            item.content
                        )
                        
                        # 添加到 ChromaDB
                        await loop.run_in_executor(
                            None,
                            lambda item_id=item.id, emb=embedding, content=item.content, src=item.source: 
                                self.chroma_collection.add(
                                    embeddings=[emb.tolist()],
                                    ids=[item_id],
                                    metadatas=[{"content": content, "source": src or ""}]
                                )
                        )
                        
                        synced_count += 1
                        
                        if synced_count % 5 == 0:
                            encoding_fix.safe_print(f"   ✅ 已同步 {synced_count}/{sqlite_count - len(existing_ids)} 条...")
                        
                    except Exception as e:
                        error_count += 1
                        encoding_fix.safe_print(f"   ❌ 同步失败 [{item.id}]: {e}")
                
                # 输出同步结果
                encoding_fix.safe_print(f"\n📊 [同步完成]")
                encoding_fix.safe_print(f"   ✅ 新增: {synced_count} 条")
                encoding_fix.safe_print(f"   ⏭️ 跳过: {skipped_count} 条 (已存在)")
                encoding_fix.safe_print(f"   ❌ 失败: {error_count} 条")
                
                # 验证同步后的数量
                final_count = await loop.run_in_executor(
                    None,
                    lambda: self.chroma_collection.count()
                )
                encoding_fix.safe_print(f"   💾 ChromaDB 最终文档数: {final_count}")
                
                if final_count == sqlite_count:
                    encoding_fix.safe_print(f"✅ [数据同步] 数据库已同步！")
                else:
                    encoding_fix.safe_print(f"⚠️ [数据同步] 同步后仍有差异，请检查日志")
            else:
                encoding_fix.safe_print(f"✅ [数据同步] 数据库一致，无需同步")
            
            # 同步反馈数据
            await self._sync_feedback_databases()
                
        except Exception as e:
            encoding_fix.safe_print(f"❌ [数据同步] 同步过程出错: {e}")
            encoding_fix.safe_print(f"   提示: 如果问题持续，可以手动运行 restore_knowledge.py")

    async def _sync_feedback_databases(self):
        """同步反馈数据库"""
        try:
            encoding_fix.safe_print("\n🔄 [反馈同步] 检查反馈数据一致性...")
            
            # 导入模型
            from models.knowledge import KnowledgeFeedback
            
            # 1. 获取 SQLite 中的反馈数量
            async with AsyncSQLiteSession() as session:
                result = await session.execute(select(KnowledgeFeedback))
                sqlite_feedbacks = result.scalars().all()
                sqlite_count = len(sqlite_feedbacks)
            
            # 2. 获取 ChromaDB 中的反馈问题数量
            loop = asyncio.get_event_loop()
            if self.feedback_collection:
                chroma_count = await loop.run_in_executor(
                    None,
                    lambda: self.feedback_collection.count()
                )
            else:
                chroma_count = 0
            
            encoding_fix.safe_print(f"   📊 SQLite 反馈数: {sqlite_count}")
            encoding_fix.safe_print(f"   📊 ChromaDB 反馈问题数: {chroma_count}")
            
            # 3. 如果数量不一致，进行同步
            if sqlite_count != chroma_count and self.feedback_collection:
                encoding_fix.safe_print(f"⚠️ [反馈不一致] 开始同步...")
                
                # 获取已存在的ID
                existing_ids = set()
                if chroma_count > 0:
                    chroma_data = await loop.run_in_executor(
                        None,
                        lambda: self.feedback_collection.get()
                    )
                    existing_ids = set(chroma_data.get('ids', []))
                
                # 同步缺失的反馈
                synced = 0
                for feedback in sqlite_feedbacks:
                    if feedback.id not in existing_ids:
                        # 生成问题向量
                        question_embedding = await loop.run_in_executor(
                            None,
                            self.sentence_model.encode,
                            feedback.question_text
                        )
                        
                        # 添加到 ChromaDB
                        await loop.run_in_executor(
                            None,
                            lambda fid=feedback.id, emb=question_embedding, q=feedback.question_text, lc=feedback.like_count: 
                                self.feedback_collection.add(
                                    embeddings=[emb.tolist()],
                                    ids=[fid],
                                    metadatas=[{
                                        "question": q,
                                        "like_count": lc
                                    }]
                                )
                        )
                        synced += 1
                
                encoding_fix.safe_print(f"✅ [反馈同步] 完成，同步了 {synced} 条反馈")
            else:
                encoding_fix.safe_print(f"✅ [反馈同步] 数据一致，无需同步")
                
        except Exception as e:
            encoding_fix.safe_print(f"❌ [反馈同步] 失败: {e}")
    
    async def _cleanup_duplicate_feedbacks(self):
        """清理重复的反馈记录（同一个问题有多条记录）"""
        try:
            encoding_fix.safe_print("\n🧹 [重复清理] 检查并清理重复的反馈记录...")
            
            from models.knowledge import KnowledgeFeedback
            from sqlalchemy import func as sql_func
            
            async with AsyncSQLiteSession() as session:
                # 找出所有重复的 question_hash
                stmt = select(
                    KnowledgeFeedback.question_hash,
                    sql_func.count(KnowledgeFeedback.id).label('count')
                ).group_by(
                    KnowledgeFeedback.question_hash
                ).having(
                    sql_func.count(KnowledgeFeedback.id) > 1
                )
                
                result = await session.execute(stmt)
                duplicates = result.all()
                
                if not duplicates:
                    encoding_fix.safe_print("✅ [重复清理] 没有发现重复记录")
                    return
                
                encoding_fix.safe_print(f"⚠️ [重复清理] 发现 {len(duplicates)} 组重复记录")
                
                total_deleted = 0
                for question_hash, count in duplicates:
                    encoding_fix.safe_print(f"   📋 处理重复组: {question_hash[:8]}... (共{count}条)")
                    
                    # 获取该问题的所有记录，按点赞数和创建时间排序
                    stmt = select(KnowledgeFeedback).where(
                        KnowledgeFeedback.question_hash == question_hash
                    ).order_by(
                        KnowledgeFeedback.like_count.desc(),
                        KnowledgeFeedback.created_at.desc()
                    )
                    result = await session.execute(stmt)
                    records = result.scalars().all()
                    
                    # 保留点赞数最高的（如果点赞数相同则保留最新的）
                    keep_record = records[0]
                    delete_records = records[1:]
                    
                    encoding_fix.safe_print(f"      ✅ 保留: ID={keep_record.id[:8]}..., 点赞={keep_record.like_count}")
                    
                    # 删除其他记录
                    for record in delete_records:
                        encoding_fix.safe_print(f"      🗑️ 删除: ID={record.id[:8]}..., 点赞={record.like_count}")
                        
                        # 从 SQLite 删除
                        await session.delete(record)
                        
                        # 从 ChromaDB 删除
                        if self.feedback_collection:
                            try:
                                loop = asyncio.get_event_loop()
                                await loop.run_in_executor(
                                    None,
                                    lambda rid=record.id: self.feedback_collection.delete(ids=[rid])
                                )
                            except Exception as e:
                                encoding_fix.safe_print(f"         ⚠️ ChromaDB删除失败: {e}")
                        
                        total_deleted += 1
                
                await session.commit()
                encoding_fix.safe_print(f"✅ [重复清理] 完成，共删除 {total_deleted} 条重复记录")
            
            # 清理 ChromaDB 中的孤立记录（SQLite中不存在的）
            await self._cleanup_orphan_vectors()
                
        except Exception as e:
            encoding_fix.safe_print(f"❌ [重复清理] 失败: {e}")
    
    async def _cleanup_orphan_vectors(self):
        """清理 ChromaDB 中的孤立向量（SQLite 中不存在的记录）"""
        try:
            if not self.feedback_collection:
                return
            
            encoding_fix.safe_print("\n🧹 [向量清理] 检查并清理孤立的向量记录...")
            
            from models.knowledge import KnowledgeFeedback
            
            # 获取 SQLite 中所有有效的反馈 ID
            async with AsyncSQLiteSession() as session:
                stmt = select(KnowledgeFeedback.id)
                result = await session.execute(stmt)
                valid_ids = set(row[0] for row in result.all())
            
            encoding_fix.safe_print(f"   📊 SQLite 中有效反馈数: {len(valid_ids)}")
            
            # 获取 ChromaDB 中所有的反馈 ID
            loop = asyncio.get_event_loop()
            chroma_data = await loop.run_in_executor(
                None,
                lambda: self.feedback_collection.get()
            )
            chroma_ids = set(chroma_data.get('ids', []))
            
            encoding_fix.safe_print(f"   📊 ChromaDB 中向量数: {len(chroma_ids)}")
            
            # 找出孤立的向量（在 ChromaDB 中但不在 SQLite 中）
            orphan_ids = chroma_ids - valid_ids
            
            if orphan_ids:
                encoding_fix.safe_print(f"⚠️ [向量清理] 发现 {len(orphan_ids)} 个孤立向量，开始清理...")
                
                # 批量删除孤立向量
                orphan_ids_list = list(orphan_ids)
                batch_size = 100
                deleted_count = 0
                
                for i in range(0, len(orphan_ids_list), batch_size):
                    batch = orphan_ids_list[i:i+batch_size]
                    try:
                        await loop.run_in_executor(
                            None,
                            lambda b=batch: self.feedback_collection.delete(ids=b)
                        )
                        deleted_count += len(batch)
                        encoding_fix.safe_print(f"   🗑️ 已删除 {deleted_count}/{len(orphan_ids)} 个孤立向量...")
                    except Exception as e:
                        encoding_fix.safe_print(f"   ⚠️ 批量删除失败: {e}")
                
                encoding_fix.safe_print(f"✅ [向量清理] 完成，共删除 {deleted_count} 个孤立向量")
            else:
                encoding_fix.safe_print(f"✅ [向量清理] 没有发现孤立向量")
                
        except Exception as e:
            encoding_fix.safe_print(f"❌ [向量清理] 失败: {e}")
    
    async def _process_query_queue(self):
        """处理查询排队队列"""
        encoding_fix.safe_print("🚀 [排队系统] 查询排队处理器已启动")
        
        while True:
            try:
                # 从队列中获取查询任务
                query_task = await self._query_queue.get()
                
                if query_task is None:  # 停止信号
                    break
                
                query_id = query_task['id']
                question = query_task['question']
                result_future = query_task['future']
                
                encoding_fix.safe_print(f"📝 [排队系统] 开始处理查询 {query_id}")
                self._is_processing = True
                self._current_query_id = query_id
                
                try:
                    # 执行实际的查询
                    result = await self._do_query_knowledge_stream(question)
                    result_future.set_result(result)
                except Exception as e:
                    result_future.set_exception(e)
                finally:
                    self._is_processing = False
                    self._current_query_id = None
                    self._query_queue.task_done()
                    encoding_fix.safe_print(f"✅ [排队系统] 查询 {query_id} 处理完成")
                    
            except Exception as e:
                encoding_fix.safe_print(f"❌ [排队系统] 处理出错: {e}")
    
    async def query_knowledge_stream_with_queue(self, question: str) -> Dict:
        """带排队功能的流式查询（方案2：所有请求都排队）"""
        import uuid
        
        query_id = str(uuid.uuid4())[:8]
        queue_size = self._query_queue.qsize()
        
        # 所有请求都加入队列（包括第一个）
        result_future = asyncio.Future()
        await self._query_queue.put({
            'id': query_id,
            'question': question,
            'future': result_future
        })
        
        if queue_size > 0:
            # 有其他请求在排队，显示排队信息
            encoding_fix.safe_print(f"⏳ [排队系统] 查询 {query_id} 加入队列，位置: {queue_size + 1}")
            
            async def queue_waiting_stream():
                yield f"⏳ 当前有 {queue_size} 个查询正在处理，您的查询排在第 {queue_size + 1} 位...\n"
                yield "⏱️ 预计等待时间: 10-30秒\n\n"
                
                # 等待结果
                try:
                    result = await result_future
                    # 将实际结果的流传递出去
                    if 'answer_stream' in result:
                        async for chunk in result['answer_stream']:
                            yield chunk
                except Exception as e:
                    yield f"\n❌ 查询处理失败: {str(e)}"
            
            return {
                'relevant_docs': [],
                'answer_stream': queue_waiting_stream(),
                'queued': True,
                'queue_position': queue_size + 1
            }
        else:
            # 队列为空，立即处理（但仍然通过队列）
            encoding_fix.safe_print(f"🚀 [排队系统] 查询 {query_id} 立即处理")
            result = await result_future
            return result
    
    async def _do_query_knowledge_stream(self, question: str) -> Dict:
        """实际执行查询的内部方法（带反馈优化）"""
        safe_question = encoding_fix.safe_encode_for_model(question)
        encoding_fix.safe_print(f"\n🌊 [流式AI处理] 收到问题: {safe_question[:50]}...")
        
        if not self.initialized:
            await self.initialize()
        
        # 🎯 步骤1：检查是否有相似的点赞反馈
        feedback_match = await self._find_similar_feedback(safe_question)
        
        if feedback_match:
            encoding_fix.safe_print(f"✨ [点赞匹配] 找到相似反馈（相似度: {feedback_match['similarity']:.2%}，点赞数: {feedback_match['like_count']}）")
            
            # 更新使用统计
            await self._update_feedback_usage(feedback_match['id'])
            
            # 使用点赞反馈生成回答
            return await self._generate_answer_from_feedback(safe_question, feedback_match)
        
        # 如果没有匹配的反馈，使用正常的RAG流程
        encoding_fix.safe_print(f"📚 [常规检索] 未找到匹配的点赞反馈，使用常规RAG")
        return await self.query_knowledge_stream(safe_question)
    
    async def _find_similar_feedback(
        self, 
        question: str,
        similarity_threshold: float = 0.90  # 相似度阈值（提高到90%，确保精确匹配）
    ) -> Optional[Dict]:
        """查找相似的点赞反馈（优先返回相似度最高的，点赞数作为次要因素）
        
        注意：由于问题格式已简化为"错误码 X：错误信息"，需要更高的相似度阈值来确保精确匹配
        """
        
        if not self.feedback_collection:
            return None
        
        try:
            # 生成问题向量
            loop = asyncio.get_event_loop()
            question_embedding = await loop.run_in_executor(
                None,
                self.sentence_model.encode,
                question
            )
            
            # 在反馈集合中搜索（多返回几个候选）
            search_results = await loop.run_in_executor(
                None,
                lambda: self.feedback_collection.query(
                    query_embeddings=[question_embedding.tolist()],
                    n_results=5,  # 获取前5个相似的
                    include=["metadatas", "distances"]
                )
            )
            
            if search_results['ids'] and len(search_results['ids'][0]) > 0:
                # 收集所有满足相似度阈值的反馈
                candidates = []
                
                for idx in range(len(search_results['ids'][0])):
                    distance = search_results['distances'][0][idx]
                    similarity = 1 - distance
                    
                    if similarity >= similarity_threshold:
                        feedback_id = search_results['ids'][0][idx]
                        metadata = search_results['metadatas'][0][idx]
                        
                        # 从SQLite获取完整反馈信息
                        from models.knowledge import KnowledgeFeedback
                        async with AsyncSQLiteSession() as session:
                            stmt = select(KnowledgeFeedback).where(
                                KnowledgeFeedback.id == feedback_id
                            )
                            result = await session.execute(stmt)
                            feedback = result.scalar_one_or_none()
                            
                            if feedback:
                                import json
                                # 调试日志：检查答案内容
                                encoding_fix.safe_print(f"🔍 [反馈调试] ID: {feedback.id}, 点赞: {feedback.like_count}, 答案预览: {feedback.answer_text[:50]}...")
                                candidates.append({
                                    'id': feedback.id,
                                    'answer': feedback.answer_text,
                                    'doc_ids': json.loads(feedback.relevant_doc_ids),
                                    'like_count': feedback.like_count,
                                    'similarity': similarity
                                })
                
                # 如果有候选，按相似度优先，点赞数次之
                if candidates:
                    # 先按相似度降序，再按点赞数降序（相似度权重更高）
                    candidates.sort(key=lambda x: (x['similarity'], x['like_count']), reverse=True)
                    best_match = candidates[0]
                    
                    encoding_fix.safe_print(f"🏆 [反馈匹配] 找到{len(candidates)}个候选，选择最佳（相似度:{best_match['similarity']:.2%}, 点赞:{best_match['like_count']}）")
                    return best_match
            
            return None
            
        except Exception as e:
            encoding_fix.safe_print(f"⚠️ [反馈查询] 失败: {e}")
            return None
    
    async def _update_feedback_usage(self, feedback_id: str):
        """更新反馈使用统计"""
        try:
            from models.knowledge import KnowledgeFeedback
            from sqlalchemy import update
            async with AsyncSQLiteSession() as session:
                from sqlalchemy.sql import func
                stmt = update(KnowledgeFeedback).where(
                    KnowledgeFeedback.id == feedback_id
                ).values(
                    use_count=KnowledgeFeedback.use_count + 1,
                    last_used_at=func.now()
                )
                await session.execute(stmt)
                await session.commit()
        except Exception as e:
            encoding_fix.safe_print(f"⚠️ [使用统计] 更新失败: {e}")
    
    async def _generate_answer_from_feedback(
        self, 
        question: str, 
        feedback_match: Dict
    ) -> Dict:
        """基于点赞反馈生成回答（直接返回缓存的答案）"""
        
        encoding_fix.safe_print(f"💡 [反馈复用] 直接使用点赞回答（点赞数: {feedback_match['like_count']}）")
        
        # 从ChromaDB获取反馈中的文档详情
        doc_ids = feedback_match['doc_ids']
        relevant_docs = []
        
        if doc_ids:
            try:
                loop = asyncio.get_event_loop()
                docs_data = await loop.run_in_executor(
                    None,
                    lambda: self.chroma_collection.get(
                        ids=doc_ids,
                        include=["metadatas"]
                    )
                )
                
                if docs_data and docs_data.get('metadatas'):
                    for i, metadata in enumerate(docs_data['metadatas']):
                        relevant_docs.append({
                            'id': doc_ids[i] if i < len(doc_ids) else '',
                            'content': metadata.get('content', ''),
                            'source': metadata.get('source', ''),
                            'similarity': 0.95  # 来自点赞反馈，标记高相似度
                        })
            except Exception as e:
                encoding_fix.safe_print(f"⚠️ [文档获取] 失败: {e}")
        
        # 直接返回缓存的答案，不需要重新生成
        cached_answer = feedback_match['answer']
        
        # 创建一个简单的流式生成器，立即返回缓存的答案
        async def generate_cached_answer_stream():
            try:
                # 添加精选回答提示（不显示点赞数，右上角已有显示）
                prefix = "⭐ 以下是精选高赞回答\n\n"
                
                # 先输出提示前缀
                for char in prefix:
                    yield char
                    await asyncio.sleep(0.01)
                
                # 再输出缓存的答案
                chunk_size = 10  # 每次输出10个字符
                for i in range(0, len(cached_answer), chunk_size):
                    chunk = cached_answer[i:i+chunk_size]
                    yield chunk
                    await asyncio.sleep(0.02)  # 轻微延迟，让用户感知流式输出
                
                encoding_fix.safe_print(f"✅ [反馈复用] 缓存答案已返回，长度: {len(cached_answer)} 字符")
            except Exception as e:
                encoding_fix.safe_print(f"❌ [反馈复用] 返回失败: {e}")
                yield f"抱歉，返回缓存答案时出现错误: {str(e)}"
        
        return {
            'relevant_docs': relevant_docs,
            'answer_stream': generate_cached_answer_stream(),
            'from_feedback': True,
            'feedback_like_count': feedback_match['like_count'],
            'feedback_similarity': feedback_match['similarity'],
            'cached_answer': True,  # 标记这是缓存的答案
            'feedback_id': feedback_match['id']  # 返回反馈ID，用于减少点赞
        }
    
    async def record_feedback(
        self, 
        question: str, 
        answer: str, 
        relevant_docs: List[Dict]
    ) -> str:
        """记录或替换用户点赞反馈（完全替换模式：每个问题只保留一个最优答案）
        
        如果问题已存在：替换答案内容，重置点赞数为1，同时更新SQLite和ChromaDB
        如果问题不存在：创建新记录，同时存储到SQLite和ChromaDB
        """
        import hashlib
        import json
        
        if not self.initialized:
            await self.initialize()
        
        try:
            # 生成问题特征哈希（用于去重）
            # 对问题进行简单的标准化处理
            normalized_question = question.strip().lower()
            question_hash = hashlib.md5(normalized_question.encode()).hexdigest()
            
            # 检查是否已存在（查找所有相同问题的记录）
            from models.knowledge import KnowledgeFeedback
            from sqlalchemy.sql import func
            async with AsyncSQLiteSession() as session:
                stmt = select(KnowledgeFeedback).where(
                    KnowledgeFeedback.question_hash == question_hash
                ).order_by(KnowledgeFeedback.created_at.desc())  # 按创建时间倒序
                result = await session.execute(stmt)
                existing_records = result.scalars().all()
                
                if existing_records:
                    # 如果存在多条记录，只保留第一条，删除其他的
                    if len(existing_records) > 1:
                        encoding_fix.safe_print(f"⚠️ [重复记录] 发现 {len(existing_records)} 条重复记录，清理中...")
                        for duplicate in existing_records[1:]:
                            # 从 SQLite 删除
                            await session.delete(duplicate)
                            encoding_fix.safe_print(f"   🗑️ 删除重复记录: {duplicate.id}")
                            
                            # 从 ChromaDB 删除
                            if self.feedback_collection:
                                try:
                                    loop = asyncio.get_event_loop()
                                    await loop.run_in_executor(
                                        None,
                                        lambda did=duplicate.id: self.feedback_collection.delete(ids=[did])
                                    )
                                except Exception as e:
                                    encoding_fix.safe_print(f"   ⚠️ ChromaDB删除失败: {e}")
                        
                        await session.commit()
                        encoding_fix.safe_print(f"✅ [重复清理] 已清理 {len(existing_records) - 1} 条重复记录")
                    
                    existing = existing_records[0]
                    encoding_fix.safe_print(f"🔄 [反馈更新] 发现已存在的反馈，ID: {existing.id}")
                    
                    # 清理答案中的精选提示前缀（避免重复累加）
                    clean_answer = answer
                    prefix_pattern = "⭐ 以下是精选高赞回答\n\n"
                    while clean_answer.startswith(prefix_pattern):
                        clean_answer = clean_answer[len(prefix_pattern):]
                    
                    # 清理已存在答案的前缀
                    clean_existing_answer = existing.answer_text
                    while clean_existing_answer.startswith(prefix_pattern):
                        clean_existing_answer = clean_existing_answer[len(prefix_pattern):]
                    
                    # 判断是否为相同答案（去除前缀后比较）
                    encoding_fix.safe_print(f"📊 [答案比较] 清理后的新答案长度: {len(clean_answer.strip())}")
                    encoding_fix.safe_print(f"📊 [答案比较] 清理后的旧答案长度: {len(clean_existing_answer.strip())}")
                    encoding_fix.safe_print(f"📊 [答案比较] 新答案前80字: {clean_answer.strip()[:80]}")
                    encoding_fix.safe_print(f"📊 [答案比较] 旧答案前80字: {clean_existing_answer.strip()[:80]}")
                    
                    if clean_answer.strip() == clean_existing_answer.strip():
                        # 相同答案，增加点赞数（工人的成就感）
                        encoding_fix.safe_print(f"✅ [答案判定] 答案相同，累加点赞")
                        existing.like_count += 1
                        existing.updated_at = func.now()
                        await session.commit()
                        
                        # 更新 ChromaDB 元数据中的点赞数
                        if self.feedback_collection:
                            try:
                                loop = asyncio.get_event_loop()
                                await loop.run_in_executor(
                                    None,
                                    lambda: self.feedback_collection.update(
                                        ids=[existing.id],
                                        metadatas=[{
                                            "question": question,
                                            "like_count": existing.like_count
                                        }]
                                    )
                                )
                                encoding_fix.safe_print(f"✅ [ChromaDB更新] 已更新点赞数")
                            except Exception as e:
                                encoding_fix.safe_print(f"⚠️ [ChromaDB更新] 失败: {e}")
                        
                        encoding_fix.safe_print(f"👍 [点赞累加] 相同答案，点赞数: {existing.like_count - 1} -> {existing.like_count}")
                        return existing.id
                    else:
                        # 不同答案，替换模式
                        encoding_fix.safe_print(f"🔄 [答案替换] 检测到新答案，替换旧答案")
                        encoding_fix.safe_print(f"   📝 旧答案预览: {clean_existing_answer[:80]}...")
                        encoding_fix.safe_print(f"   📝 新答案预览: {clean_answer[:80]}...")
                        
                        # 提取文档ID
                        doc_ids = [doc.get('id') for doc in relevant_docs if doc.get('id')]
                        
                        # 更新答案内容（使用清理后的答案）
                        existing.answer_text = clean_answer
                        existing.relevant_doc_ids = json.dumps(doc_ids)
                        existing.like_count = 1  # 重置为1
                        existing.updated_at = func.now()
                        await session.commit()
                        
                        encoding_fix.safe_print(f"✅ [SQLite更新] 答案已在数据库中更新")
                        
                        # 同时更新 ChromaDB
                        if self.feedback_collection:
                            try:
                                loop = asyncio.get_event_loop()
                                # 重新生成问题向量
                                question_embedding = await loop.run_in_executor(
                                    None,
                                    self.sentence_model.encode,
                                    question
                                )
                                
                                # 更新 ChromaDB 中的向量和元数据
                                await loop.run_in_executor(
                                    None,
                                    lambda: self.feedback_collection.update(
                                        ids=[existing.id],
                                        embeddings=[question_embedding.tolist()],
                                        metadatas=[{
                                            "question": question,
                                            "like_count": 1
                                        }]
                                    )
                                )
                                encoding_fix.safe_print(f"✅ [ChromaDB更新] 已同步更新向量和元数据")
                            except Exception as e:
                                encoding_fix.safe_print(f"⚠️ [ChromaDB更新] 失败: {e}")
                        
                        encoding_fix.safe_print(f"✅ [答案替换] 已用新答案替换旧答案，ID: {existing.id}")
                        encoding_fix.safe_print(f"   🔢 点赞数已重置为: 1")
                        return existing.id
                else:
                    # 新建反馈记录
                    feedback_id = str(uuid.uuid4())
                    encoding_fix.safe_print(f"🆕 [反馈记录] 创建新反馈，ID: {feedback_id}")
                    
                    # 清理答案中的精选提示前缀（首次点赞不应该有前缀，但以防万一）
                    clean_answer = answer
                    prefix_pattern = "⭐ 以下是精选高赞回答\n\n"
                    while clean_answer.startswith(prefix_pattern):
                        clean_answer = clean_answer[len(prefix_pattern):]
                    
                    # 提取文档ID
                    doc_ids = [doc.get('id') for doc in relevant_docs if doc.get('id')]
                    
                    feedback = KnowledgeFeedback(
                        id=feedback_id,
                        question_text=question,
                        question_hash=question_hash,
                        answer_text=clean_answer,  # 使用清理后的答案
                        relevant_doc_ids=json.dumps(doc_ids),
                        like_count=1
                    )
                    session.add(feedback)
                    await session.commit()
                    
                    # 同时添加到 ChromaDB
                    if self.feedback_collection:
                        loop = asyncio.get_event_loop()
                        question_embedding = await loop.run_in_executor(
                            None,
                            self.sentence_model.encode,
                            question
                        )
                        
                        await loop.run_in_executor(
                            None,
                            lambda: self.feedback_collection.add(
                                embeddings=[question_embedding.tolist()],
                                ids=[feedback_id],
                                metadatas=[{
                                    "question": question,
                                    "like_count": 1
                                }]
                            )
                        )
                    
                    encoding_fix.safe_print(f"✅ [反馈记录] 新建反馈记录: {feedback_id}")
                    return feedback_id
                    
        except Exception as e:
            encoding_fix.safe_print(f"❌ [反馈记录] 失败: {e}")
            raise e

# 全局知识库服务实例
knowledge_service = KnowledgeService()
