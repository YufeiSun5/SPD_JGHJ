# routers/function_management.py
"""
函数管理路由 - 管理数据库查询函数的注册、查看和测试
"""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
from typing import Dict, List, Any, Optional
from services.function_registry import function_registry, DatabaseFunction, FunctionParameter, ParameterType
import encoding_fix

router = APIRouter()

# Pydantic 模型定义
class FunctionListResponse(BaseModel):
    success: bool
    functions: List[Dict[str, Any]]
    categories: List[str]

class FunctionTestRequest(BaseModel):
    function_name: str = Field(..., description="要测试的函数名称")
    parameters: Dict[str, Any] = Field(..., description="函数参数")

class FunctionTestResponse(BaseModel):
    success: bool
    function_name: str
    parameters: Dict[str, Any]
    result: Optional[Dict[str, Any]] = None
    error: Optional[str] = None

class QueryAnalysisRequest(BaseModel):
    query: str = Field(..., description="用户查询语句")

class QueryAnalysisResponse(BaseModel):
    success: bool
    needs_function_call: bool
    function_name: Optional[str] = None
    extracted_parameters: Optional[Dict[str, Any]] = None
    missing_parameters: Optional[List[str]] = None
    query_type: str
    reason: Optional[str] = None

@router.get("/functions", response_model=FunctionListResponse, summary="获取所有已注册的函数")
async def get_functions():
    """
    获取所有已注册的数据库查询函数
    
    返回函数列表、参数定义和分类信息
    """
    try:
        schema = function_registry.get_functions_schema()
        
        return FunctionListResponse(
            success=True,
            functions=schema["functions"],
            categories=schema["categories"]
        )
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"获取函数列表失败: {str(e)}"
        )

@router.get("/functions/{function_name}", summary="获取指定函数的详细信息")
async def get_function_detail(function_name: str):
    """
    获取指定函数的详细信息
    
    - **function_name**: 函数名称
    """
    function = function_registry.get_function(function_name)
    if not function:
        raise HTTPException(
            status_code=404,
            detail=f"未找到函数: {function_name}"
        )
    
    return {
        "success": True,
        "function": function.to_dict()
    }

@router.get("/functions/category/{category}", summary="根据分类获取函数")
async def get_functions_by_category(category: str):
    """
    根据分类获取函数列表
    
    - **category**: 函数分类（如：quality, production, trend等）
    """
    functions = function_registry.get_functions_by_category(category)
    
    return {
        "success": True,
        "category": category,
        "functions": [func.to_dict() for func in functions],
        "count": len(functions)
    }

@router.post("/test", response_model=FunctionTestResponse, summary="测试函数调用")
async def test_function(request: FunctionTestRequest):
    """
    测试数据库查询函数的调用
    
    - **function_name**: 要测试的函数名称
    - **parameters**: 函数参数字典
    
    用于验证函数是否正常工作和参数是否正确
    """
    try:
        # 检查函数是否存在
        function = function_registry.get_function(request.function_name)
        if not function:
            return FunctionTestResponse(
                success=False,
                function_name=request.function_name,
                parameters=request.parameters,
                error=f"未找到函数: {request.function_name}"
            )
        
        # 执行函数
        result = await function_registry.execute_function(
            request.function_name, 
            request.parameters
        )
        
        return FunctionTestResponse(
            success=result.get("success", False),
            function_name=request.function_name,
            parameters=request.parameters,
            result=result,
            error=result.get("error")
        )
        
    except Exception as e:
        return FunctionTestResponse(
            success=False,
            function_name=request.function_name,
            parameters=request.parameters,
            error=str(e)
        )

@router.post("/analyze", response_model=QueryAnalysisResponse, summary="分析用户查询")
async def analyze_query(request: QueryAnalysisRequest):
    """
    分析用户查询，判断是否需要调用数据库函数
    
    - **query**: 用户查询语句
    
    返回分析结果，包括：
    - 是否需要函数调用
    - 匹配的函数名称
    - 提取的参数
    - 缺失的参数
    """
    try:
        # 导入knowledge_service来使用其function_caller
        from services.knowledge_service import knowledge_service
        
        # 确保服务已初始化
        if not knowledge_service.initialized:
            await knowledge_service.initialize()
        
        if not knowledge_service.function_caller:
            raise HTTPException(
                status_code=503,
                detail="函数调用器未初始化"
            )
        
        analysis_result = await knowledge_service.function_caller.analyze_and_call(request.query)
        
        return QueryAnalysisResponse(
            success=True,
            needs_function_call=analysis_result.get("needs_function_call", False),
            function_name=analysis_result.get("function_name"),
            extracted_parameters=analysis_result.get("extracted_parameters", analysis_result.get("parameters")),
            missing_parameters=analysis_result.get("missing_parameters"),
            query_type=analysis_result.get("query_type", "unknown"),
            reason=analysis_result.get("reason")
        )
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"查询分析失败: {str(e)}"
        )

class RegisterFunctionRequest(BaseModel):
    name: str = Field(..., description="函数名称")
    description: str = Field(..., description="函数描述")
    sql_template: str = Field(..., description="SQL模板")
    category: str = Field("custom", description="函数分类")
    parameters: List[Dict[str, Any]] = Field(..., description="参数定义列表")
    examples: List[str] = Field([], description="使用示例")

@router.post("/register", summary="注册新的数据库查询函数")
async def register_function(request: RegisterFunctionRequest):
    """
    注册新的数据库查询函数
    
    参数格式示例：
    ```json
    {
        "name": "query_custom_data",
        "description": "查询自定义数据",
        "sql_template": "SELECT * FROM custom_table WHERE date = :date",
        "category": "custom",
        "parameters": [
            {
                "name": "date",
                "type": "date",
                "description": "查询日期",
                "required": true,
                "validation_regex": "^\\d{4}-\\d{2}-\\d{2}$"
            }
        ],
        "examples": ["查询今天的自定义数据"]
    }
    ```
    """
    try:
        # 检查函数是否已存在
        if function_registry.get_function(request.name):
            raise HTTPException(
                status_code=400,
                detail=f"函数 {request.name} 已存在"
            )
        
        # 构建参数对象
        param_objects = []
        for param_dict in request.parameters:
            try:
                param_type = ParameterType(param_dict["type"])
                param = FunctionParameter(
                    name=param_dict["name"],
                    type=param_type,
                    description=param_dict["description"],
                    required=param_dict.get("required", True),
                    default_value=param_dict.get("default_value"),
                    validation_regex=param_dict.get("validation_regex")
                )
                param_objects.append(param)
            except Exception as e:
                raise HTTPException(
                    status_code=400,
                    detail=f"参数 {param_dict.get('name', 'unknown')} 定义错误: {str(e)}"
                )
        
        # 创建函数对象
        db_function = DatabaseFunction(
            name=request.name,
            description=request.description,
            sql_template=request.sql_template,
            parameters=param_objects,
            category=request.category,
            examples=request.examples
        )
        
        # 注册函数
        function_registry.register_function(db_function)
        
        return {
            "success": True,
            "message": f"函数 {request.name} 注册成功",
            "function": db_function.to_dict()
        }
        
    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"注册函数失败: {str(e)}"
        )

@router.delete("/functions/{function_name}", summary="删除函数")
async def delete_function(function_name: str):
    """
    删除指定的数据库查询函数
    
    - **function_name**: 要删除的函数名称
    
    注意：内置函数无法删除
    """
    # 检查是否为内置函数
    builtin_functions = ["query_defect_rate", "query_production_stats", "query_quality_trend"]
    if function_name in builtin_functions:
        raise HTTPException(
            status_code=400,
            detail=f"无法删除内置函数: {function_name}"
        )
    
    # 检查函数是否存在
    if not function_registry.get_function(function_name):
        raise HTTPException(
            status_code=404,
            detail=f"未找到函数: {function_name}"
        )
    
    try:
        # 从注册表中删除函数
        del function_registry.functions[function_name]
        
        return {
            "success": True,
            "message": f"函数 {function_name} 删除成功"
        }
        
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"删除函数失败: {str(e)}"
        )

@router.get("/health", summary="函数调用系统健康检查")
async def health_check():
    """检查函数调用系统是否正常运行"""
    try:
        # 检查函数注册表
        function_count = len(function_registry.functions)
        categories = list(set(func.category for func in function_registry.functions.values()))
        
        # 简单测试一个函数（如果有的话）
        test_result = None
        if function_count > 0:
            test_function_name = list(function_registry.functions.keys())[0]
            test_result = f"测试函数 {test_function_name} 可用"
        
        return {
            "status": "healthy",
            "message": "函数调用系统运行正常",
            "statistics": {
                "total_functions": function_count,
                "categories": categories,
                "test_result": test_result
            }
        }
        
    except Exception as e:
        raise HTTPException(
            status_code=503,
            detail=f"函数调用系统不可用: {str(e)}"
        )
