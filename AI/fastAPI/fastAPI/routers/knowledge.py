# routers/knowledge.py
from fastapi import APIRouter, HTTPException, Depends
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field
from typing import Optional, List, Dict, AsyncGenerator
import json
import asyncio
import config

# 统一使用 knowledge_service，通过配置控制行为
from services.knowledge_service import knowledge_service

if config.PURE_RAG_MODE:
    print("📚 纯 RAG 模式已启用：直接查询向量数据库")
else:
    print("🧠 智能模式已启用：支持意图识别和函数调用")

router = APIRouter()

# Pydantic 模型定义
class AddKnowledgeRequest(BaseModel):
    content: str = Field(..., description="知识内容", min_length=1, max_length=10000)
    source: Optional[str] = Field(None, description="知识来源", max_length=255)

class AddKnowledgeResponse(BaseModel):
    success: bool
    message: str
    knowledge_id: Optional[str] = None

class DeleteKnowledgeResponse(BaseModel):
    success: bool
    message: str

class QueryRequest(BaseModel):
    question: str = Field(..., description="用户问题", min_length=1, max_length=1000)

class RelevantDoc(BaseModel):
    id: str
    content: str
    source: str
    similarity: float

class QueryResponse(BaseModel):
    answer: str
    relevant_docs: List[RelevantDoc]
    context_used: bool
    error: Optional[str] = None

class KnowledgeItem(BaseModel):
    id: str
    content: str
    source: Optional[str]
    created_at: Optional[str]

class KnowledgeListResponse(BaseModel):
    success: bool
    message: str
    data: List[KnowledgeItem]
    total: int

class FeedbackRequest(BaseModel):
    question: str = Field(..., description="用户问题")
    answer: str = Field(..., description="AI回答")
    relevant_docs: List[Dict] = Field(..., description="相关文档列表")

class FeedbackResponse(BaseModel):
    success: bool
    message: str
    feedback_id: Optional[str] = None

@router.post("/add", response_model=AddKnowledgeResponse, summary="添加知识到知识库")
async def add_knowledge(request: AddKnowledgeRequest):
    """
    添加知识到知识库
    
    - **content**: 知识内容（必填，1-10000字符）
    - **source**: 知识来源（可选，最多255字符）
    
    返回知识ID，用于后续的删除或引用
    """
    try:
        knowledge_id = await knowledge_service.add_knowledge(
            content=request.content,
            source=request.source
        )
        
        return AddKnowledgeResponse(
            success=True,
            message="知识添加成功",
            knowledge_id=knowledge_id
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"添加知识失败: {str(e)}"
        )

@router.delete("/delete/{knowledge_id}", response_model=DeleteKnowledgeResponse, summary="删除指定知识")
async def delete_knowledge(knowledge_id: str):
    """
    删除指定ID的知识
    
    - **knowledge_id**: 要删除的知识ID
    
    删除操作会同时清除SQLite中的元数据和ChromaDB中的向量数据
    """
    if not knowledge_id.strip():
        raise HTTPException(
            status_code=400,
            detail="知识ID不能为空"
        )
    
    try:
        success = await knowledge_service.delete_knowledge(knowledge_id)
        
        if success:
            return DeleteKnowledgeResponse(
                success=True,
                message="知识删除成功"
            )
        else:
            return DeleteKnowledgeResponse(
                success=False,
                message="知识不存在或已被删除"
            )
            
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"删除知识失败: {str(e)}"
        )

@router.delete("/delete_by_source/{source}", response_model=DeleteKnowledgeResponse, summary="删除指定来源的所有知识")
async def delete_knowledge_by_source(source: str):
    """
    删除指定来源的所有知识（用于替换式导入）
    
    - **source**: 知识来源
    
    删除操作会同时清除SQLite中的元数据和ChromaDB中的向量数据
    """
    if not source.strip():
        raise HTTPException(
            status_code=400,
            detail="知识来源不能为空"
        )
    
    try:
        deleted_count = await knowledge_service.delete_knowledge_by_source(source)
        
        return DeleteKnowledgeResponse(
            success=True,
            message=f"成功删除 {deleted_count} 条知识"
        )
            
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"删除知识失败: {str(e)}"
        )

@router.post("/query", response_model=QueryResponse, summary="智能问答")
async def query_knowledge(request: QueryRequest):
    """
    基于知识库进行智能问答
    
    - **question**: 用户问题（必填，1-1000字符）
    
    系统会：
    1. 将问题转换为向量
    2. 在知识库中搜索最相关的文档
    3. 使用DeepSeek模型基于相关文档生成回答
    
    返回生成的回答以及相关的知识文档
    """
    try:
        result = await knowledge_service.query_knowledge(request.question)
        
        relevant_docs = [
            RelevantDoc(
                id=doc['id'],
                content=doc['content'],
                source=doc['source'],
                similarity=doc['similarity']
            )
            for doc in result['relevant_docs']
        ]
        
        return QueryResponse(
            answer=result['answer'],
            relevant_docs=relevant_docs,
            context_used=result['context_used'],
            error=result.get('error')
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"查询失败: {str(e)}"
        )

@router.post("/query-stream", summary="智能问答（流式输出）")
async def query_knowledge_stream(request: QueryRequest):
    """
    基于知识库进行智能问答 - 流式输出版本
    
    - **question**: 用户问题（必填，1-1000字符）
    
    系统会：
    1. 将问题转换为向量
    2. 在知识库中搜索最相关的文档
    3. 使用模型基于相关文档生成流式回答
    
    返回Server-Sent Events格式的流式数据
    """
    # 用于跟踪客户端是否断开
    client_disconnected = False
    
    async def generate_stream() -> AsyncGenerator[str, None]:
        try:
            import sys
            import asyncio
            print(f"[Router] Received streaming query request: {request.question}", flush=True)
            sys.stdout.flush()
            
            # 创建后台任务来调用服务
            service_task = asyncio.create_task(
                knowledge_service.query_knowledge_stream(request.question)
            )
            
            # 定义循环的"安慰性"思考步骤（会持续循环直到服务返回）
            comfort_messages = [
                "🤔 收到您的问题，正在启动分析...",
                "📝 正在理解问题内容和意图...",
                "🔧 正在准备AI模型...",
                "🔄 正在将问题转换为向量表示...",
                "🔍 正在知识库中搜索相关信息...",
                "📚 正在分析知识库内容...",
                "🧩 正在匹配最相关的知识...",
                "⚙️ 正在加载语言模型...",
                "💭 正在整理思路...",
                "💡 正在准备生成回答...",
                "🎯 正在优化回答质量...",
                "✨ 马上就好...",
            ]
            
            message_index = 0
            # 在等待服务响应期间，缓慢持续发送思考步骤
            while not service_task.done():
                # 发送当前思考步骤
                thinking_data = {
                    "type": "thinking",
                    "data": comfort_messages[message_index % len(comfort_messages)]
                }
                try:
                    yield f"data: {json.dumps(thinking_data, ensure_ascii=True)}\n\n"
                except (GeneratorExit, StopAsyncIteration):
                    print("[Router] Client disconnected during thinking phase", flush=True)
                    service_task.cancel()
                    return
                
                message_index += 1
                
                # 等待 0.8 秒，让用户有时间看清每条消息
                try:
                    await asyncio.wait_for(asyncio.shield(service_task), timeout=0.8)
                    break
                except asyncio.TimeoutError:
                    continue
            
            # 获取服务结果
            result = await service_task
            
            # 不再输出剩余的思考步骤，直接进入回答阶段
            # 这样用户会看到思考步骤自然停止，然后开始看到回答
            print(f"[Router] Received service result with keys: {list(result.keys())}", flush=True)
            sys.stdout.flush()
            
            # 如果有数据库查询结果，先发送数据库结果
            if result.get("function_called") and result.get("function_result"):
                # 安全处理数据库结果中的文本
                formatted_data = result.get("database_data", "")
                if formatted_data:
                    formatted_data = formatted_data.encode('utf-8', errors='replace').decode('utf-8', errors='replace')
                
                database_result_data = {
                    "type": "database_result",
                    "data": {
                        "success": result["function_result"].get("success", False),
                        "function_name": result["function_result"].get("function_name"),
                        "parameters": result["function_result"].get("parameters"),
                        "count": result["function_result"].get("count", 0),
                        "formatted_data": formatted_data,
                        "error": result["function_result"].get("error")
                    }
                }
                yield f"data: {json.dumps(database_result_data, ensure_ascii=True)}\n\n"
            
            # 发送相关文档
            relevant_docs = result.get("relevant_docs", [])
            relevant_docs_data = {
                "type": "docs",
                "data": relevant_docs
            }
            print(f"[Router] Sending {len(relevant_docs)} relevant documents", flush=True)
            try:
                yield f"data: {json.dumps(relevant_docs_data, ensure_ascii=True)}\n\n"
            except (GeneratorExit, StopAsyncIteration):
                print("[Router] Client disconnected while sending docs", flush=True)
                return
            
            # 发送检索结果思考步骤（最后一个真实的思考步骤）
            if len(relevant_docs) > 0:
                thinking_data = {
                    "type": "thinking",
                    "data": f"✅ 找到了 {len(relevant_docs)} 个相关文档，开始生成回答..."
                }
            else:
                thinking_data = {
                    "type": "thinking", 
                    "data": "⚠️ 知识库中未找到直接相关的信息，将使用通用知识回答..."
                }
            try:
                yield f"data: {json.dumps(thinking_data, ensure_ascii=True)}\n\n"
            except (GeneratorExit, StopAsyncIteration):
                print("[Router] Client disconnected while sending final thinking", flush=True)
                return
            
            # 流式生成回答
            print("[Router] Starting to iterate answer_stream...", flush=True)
            answer_stream = result.get("answer_stream")
            if answer_stream:
                try:
                    async for chunk in answer_stream:
                        if chunk:
                            # 安全处理chunk文本
                            safe_chunk = chunk.encode('utf-8', errors='replace').decode('utf-8', errors='replace')
                            response_data = {
                                "type": "token",
                                "data": safe_chunk
                            }
                            
                            try:
                                yield f"data: {json.dumps(response_data, ensure_ascii=True)}\n\n"
                                # 强制刷新输出缓冲区并添加微小延迟确保数据发送
                                import sys
                                sys.stdout.flush()
                                await asyncio.sleep(0)  # 让出控制权，确保数据被发送
                            except (GeneratorExit, StopAsyncIteration, ConnectionError, BrokenPipeError) as send_error:
                                # 客户端断开连接，立即停止
                                print(f"[Router] Client disconnected, stopping generation: {type(send_error).__name__}", flush=True)
                                # 关闭生成器以停止LLM
                                if hasattr(answer_stream, 'aclose'):
                                    await answer_stream.aclose()
                                return  # 直接返回，不再继续
                            except Exception as send_error:
                                # 其他异常
                                print(f"[Router] Send error: {send_error}, stopping", flush=True)
                                if hasattr(answer_stream, 'aclose'):
                                    await answer_stream.aclose()
                                return
                except GeneratorExit:
                    # 生成器被关闭
                    print(f"[Router] Generator closed by client", flush=True)
                    return
                except Exception as stream_error:
                    print(f"[Router] Stream error: {stream_error}, stopping", flush=True)
                    return
            else:
                # 如果没有流式输出，发送错误信息
                error_data = {
                    "type": "error",
                    "data": "未能获取到流式回答"
                }
                yield f"data: {json.dumps(error_data, ensure_ascii=True)}\n\n"
            
            # 发送结束信号
            end_data = {
                "type": "end",
                "data": "完成"
            }
            try:
                yield f"data: {json.dumps(end_data, ensure_ascii=True)}\n\n"
            except (GeneratorExit, StopAsyncIteration):
                print("[Router] Client disconnected before sending end signal", flush=True)
                return
            
        except Exception as e:
            # 安全处理错误信息,避免编码问题
            error_message = str(e).encode('utf-8', errors='replace').decode('utf-8', errors='replace')
            error_data = {
                "type": "error",
                "data": f"查询失败: {error_message}"
            }
            yield f"data: {json.dumps(error_data, ensure_ascii=True)}\n\n"
    
    return StreamingResponse(
        generate_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "Access-Control-Allow-Origin": "*",
        }
    )

@router.get("/list", response_model=KnowledgeListResponse, summary="获取知识库列表")
async def get_knowledge_list(
    limit: int = 20,
    offset: int = 0
):
    """
    获取知识库中的知识列表
    
    - **limit**: 返回条数限制（默认20，最大100）
    - **offset**: 偏移量（用于分页，默认0）
    
    返回知识库中的知识条目列表，按创建时间倒序排列
    """
    if limit > 100:
        limit = 100
    if limit < 1:
        limit = 1
    if offset < 0:
        offset = 0
        
    try:
        knowledge_list = await knowledge_service.get_knowledge_list(limit, offset)
        
        items = [
            KnowledgeItem(
                id=item['id'],
                content=item['content'],
                source=item['source'],
                created_at=item['created_at']
            )
            for item in knowledge_list
        ]
        
        return KnowledgeListResponse(
            success=True,
            message="获取知识列表成功",
            data=items,
            total=len(items)
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"获取知识列表失败: {str(e)}"
        )

@router.get("/health", summary="知识库服务健康检查")
async def health_check():
    """检查知识库服务是否正常运行"""
    try:
        # 尝试初始化服务（如果还没有初始化的话）
        if not knowledge_service.initialized:
            await knowledge_service.initialize()
        
        if knowledge_service.initialized:
            return {
                "status": "healthy",
                "message": "知识库服务运行正常",
                "initialized": True
            }
        else:
            return {
                "status": "degraded",
                "message": "知识库服务部分功能不可用，请检查依赖包安装",
                "initialized": False,
                "missing_dependencies": "请运行: pip install sentence-transformers chromadb llama-cpp-python"
            }
    except Exception as e:
        raise HTTPException(
            status_code=503,
            detail=f"知识库服务不可用: {str(e)}"
        )

@router.get("/performance", summary="性能统计")
async def get_performance_stats():
    """
    获取知识库性能统计信息
    
    返回包含知识库大小、性能等级、响应时间预估等信息
    """
    try:
        stats = await knowledge_service.get_performance_stats()
        return {
            "success": True,
            "data": stats
        }
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"获取性能统计失败: {str(e)}"
        )


@router.post("/feedback/like", response_model=FeedbackResponse, summary="点赞AI回答")
async def like_answer(request: FeedbackRequest):
    """
    记录用户对AI回答的点赞
    
    - **question**: 用户问题
    - **answer**: AI回答内容
    - **relevant_docs**: 相关文档列表
    
    系统会记录这个问答对，当遇到相似问题时会优先使用点赞过的回答
    """
    try:
        feedback_id = await knowledge_service.record_feedback(
            question=request.question,
            answer=request.answer,
            relevant_docs=request.relevant_docs
        )
        
        return FeedbackResponse(
            success=True,
            message="感谢您的反馈！这将帮助AI提供更准确的回答",
            feedback_id=feedback_id
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"记录反馈失败: {str(e)}"
        )

@router.post("/query-stream-with-queue", summary="智能问答（流式输出+排队）")
async def query_knowledge_stream_with_queue(request: QueryRequest):
    """
    带排队功能的智能问答 - 流式输出版本
    
    - **question**: 用户问题（必填，1-1000字符）
    
    特性：
    1. 如果有其他查询正在处理，会自动排队
    2. 会优先使用点赞过的相似问答
    3. 流式返回结果，用户体验更好
    """
    
    async def generate_stream() -> AsyncGenerator[str, None]:
        try:
            import sys
            print(f"[Router] Received queued query request: {request.question}", flush=True)
            sys.stdout.flush()
            
            # 调用带排队的查询服务
            result = await knowledge_service.query_knowledge_stream_with_queue(request.question)
            
            print(f"[Router] Received service result with keys: {list(result.keys())}", flush=True)
            
            # 如果是排队状态
            if result.get('queued'):
                queue_info_data = {
                    "type": "queue_info",
                    "data": {
                        "queued": True,
                        "position": result.get('queue_position', 0)
                    }
                }
                yield f"data: {json.dumps(queue_info_data, ensure_ascii=True)}\n\n"
            
            # 如果来自点赞反馈
            if result.get('from_feedback'):
                feedback_info_data = {
                    "type": "feedback_info",
                    "data": {
                        "from_feedback": True,
                        "like_count": result.get('feedback_like_count', 0),
                        "similarity": result.get('feedback_similarity', 0),
                        "feedback_id": result.get('feedback_id', '')  # 传递 feedback_id
                    }
                }
                yield f"data: {json.dumps(feedback_info_data, ensure_ascii=True)}\n\n"
            
            # 发送相关文档
            relevant_docs = result.get("relevant_docs", [])
            relevant_docs_data = {
                "type": "docs",
                "data": relevant_docs
            }
            print(f"[Router] Sending {len(relevant_docs)} relevant documents", flush=True)
            try:
                yield f"data: {json.dumps(relevant_docs_data, ensure_ascii=True)}\n\n"
            except (GeneratorExit, StopAsyncIteration):
                print("[Router] Client disconnected while sending docs", flush=True)
                return
            
            # 流式生成回答
            print("[Router] Starting to iterate answer_stream...", flush=True)
            answer_stream = result.get("answer_stream")
            if answer_stream:
                try:
                    async for chunk in answer_stream:
                        if chunk:
                            # 安全处理chunk文本
                            safe_chunk = chunk.encode('utf-8', errors='replace').decode('utf-8', errors='replace')
                            response_data = {
                                "type": "token",
                                "data": safe_chunk
                            }
                            
                            try:
                                yield f"data: {json.dumps(response_data, ensure_ascii=True)}\n\n"
                                sys.stdout.flush()
                                await asyncio.sleep(0)  # 让出控制权，确保数据被发送
                            except (GeneratorExit, StopAsyncIteration, ConnectionError, BrokenPipeError) as send_error:
                                print(f"[Router] Client disconnected: {type(send_error).__name__}", flush=True)
                                if hasattr(answer_stream, 'aclose'):
                                    await answer_stream.aclose()
                                return
                            except Exception as send_error:
                                print(f"[Router] Send error: {send_error}", flush=True)
                                if hasattr(answer_stream, 'aclose'):
                                    await answer_stream.aclose()
                                return
                except GeneratorExit:
                    print(f"[Router] Generator closed by client", flush=True)
                    return
                except Exception as stream_error:
                    print(f"[Router] Stream error: {stream_error}", flush=True)
                    return
            else:
                error_data = {
                    "type": "error",
                    "data": "未能获取到流式回答"
                }
                yield f"data: {json.dumps(error_data, ensure_ascii=True)}\n\n"
            
            # 发送结束信号
            end_data = {
                "type": "end",
                "data": "完成"
            }
            try:
                yield f"data: {json.dumps(end_data, ensure_ascii=True)}\n\n"
            except (GeneratorExit, StopAsyncIteration):
                print("[Router] Client disconnected before sending end signal", flush=True)
                return
            
        except Exception as e:
            error_message = str(e).encode('utf-8', errors='replace').decode('utf-8', errors='replace')
            error_data = {
                "type": "error",
                "data": f"查询失败: {error_message}"
            }
            yield f"data: {json.dumps(error_data, ensure_ascii=True)}\n\n"
    
    return StreamingResponse(
        generate_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "Access-Control-Allow-Origin": "*",
        }
    )
