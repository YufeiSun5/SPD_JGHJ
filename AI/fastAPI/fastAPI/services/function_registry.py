# services/function_registry.py
"""
Function Registry - 管理所有可调用的数据库查询函数
支持动态注册、参数解析和函数调用
"""

import json
import asyncio
from typing import Dict, List, Any, Optional, Callable
from dataclasses import dataclass, asdict
from enum import Enum
from datetime import datetime, date
import re
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import text
from db import async_session
import encoding_fix


class ParameterType(Enum):
    """参数类型枚举"""
    STRING = "string"
    DATE = "date"
    DATETIME = "datetime"
    INTEGER = "integer"
    FLOAT = "float"
    BOOLEAN = "boolean"


@dataclass
class FunctionParameter:
    """函数参数定义"""
    name: str
    type: ParameterType
    description: str
    required: bool = True
    default_value: Any = None
    validation_regex: Optional[str] = None
    
    def to_dict(self) -> Dict:
        """转换为字典格式"""
        return {
            "name": self.name,
            "type": self.type.value,
            "description": self.description,
            "required": self.required,
            "default_value": self.default_value,
            "validation_regex": self.validation_regex
        }


@dataclass
class DatabaseFunction:
    """数据库查询函数定义"""
    name: str
    description: str
    sql_template: str
    parameters: List[FunctionParameter]
    category: str = "query"
    examples: List[str] = None
    
    def __post_init__(self):
        if self.examples is None:
            self.examples = []
    
    def to_dict(self) -> Dict:
        """转换为字典格式，用于AI模型理解"""
        return {
            "name": self.name,
            "description": self.description,
            "category": self.category,
            "parameters": [param.to_dict() for param in self.parameters],
            "examples": self.examples
        }


class FunctionRegistry:
    """函数注册中心"""
    
    def __init__(self):
        self.functions: Dict[str, DatabaseFunction] = {}
        self._initialize_builtin_functions()
    
    def _initialize_builtin_functions(self):
        """初始化内置函数"""
        # 不良率查询函数
        defect_rate_function = DatabaseFunction(
            name="query_defect_rate",
            description="查询指定日期和机型的不良率数据",
            sql_template="""
                SELECT 
                    date,
                    model_name,
                    total_count,
                    defect_count,
                    ROUND((defect_count * 100.0 / total_count), 2) as defect_rate
                FROM production_quality 
                WHERE date = :date 
                    AND model_name = :model_name
                ORDER BY date DESC
            """,
            parameters=[
                FunctionParameter(
                    name="date",
                    type=ParameterType.DATE,
                    description="查询日期，格式：YYYY-MM-DD",
                    validation_regex=r"^\d{4}-\d{2}-\d{2}$"
                ),
                FunctionParameter(
                    name="model_name",
                    type=ParameterType.STRING,
                    description="机型名称，如：iPhone15、SamsungS23等"
                )
            ],
            category="quality",
            examples=[
                "查询今天iPhone15的不良率",
                "帮我看看2024-01-15 SamsungS23的质量情况",
                "今日不良率怎么样？机型是iPhone15"
            ]
        )
        
        # 生产统计查询函数
        production_stats_function = DatabaseFunction(
            name="query_production_stats",
            description="查询指定日期范围内的生产统计数据",
            sql_template="""
                SELECT 
                    date,
                    model_name,
                    SUM(total_count) as total_production,
                    SUM(defect_count) as total_defects,
                    ROUND(AVG(defect_count * 100.0 / total_count), 2) as avg_defect_rate
                FROM production_quality 
                WHERE date BETWEEN :start_date AND :end_date
                    AND (:model_name IS NULL OR model_name = :model_name)
                GROUP BY date, model_name
                ORDER BY date DESC
            """,
            parameters=[
                FunctionParameter(
                    name="start_date",
                    type=ParameterType.DATE,
                    description="开始日期，格式：YYYY-MM-DD"
                ),
                FunctionParameter(
                    name="end_date",
                    type=ParameterType.DATE,
                    description="结束日期，格式：YYYY-MM-DD"
                ),
                FunctionParameter(
                    name="model_name",
                    type=ParameterType.STRING,
                    description="机型名称（可选），如果不指定则查询所有机型",
                    required=False,
                    default_value=None
                )
            ],
            category="production",
            examples=[
                "查询本周的生产统计",
                "看看最近7天iPhone15的生产情况",
                "本月所有机型的生产数据"
            ]
        )
        
        # 质量趋势查询函数
        quality_trend_function = DatabaseFunction(
            name="query_quality_trend",
            description="查询指定机型的质量趋势数据",
            sql_template="""
                SELECT 
                    date,
                    model_name,
                    total_count,
                    defect_count,
                    ROUND((defect_count * 100.0 / total_count), 2) as defect_rate,
                    LAG(ROUND((defect_count * 100.0 / total_count), 2)) OVER (ORDER BY date) as prev_defect_rate
                FROM production_quality 
                WHERE model_name = :model_name
                    AND date >= :start_date
                ORDER BY date ASC
            """,
            parameters=[
                FunctionParameter(
                    name="model_name",
                    type=ParameterType.STRING,
                    description="机型名称"
                ),
                FunctionParameter(
                    name="start_date",
                    type=ParameterType.DATE,
                    description="开始统计日期，格式：YYYY-MM-DD"
                )
            ],
            category="trend",
            examples=[
                "iPhone15最近的质量趋势如何",
                "分析SamsungS23本月的不良率变化",
                "查看质量趋势"
            ]
        )
        
        # 生产计划查询函数
        production_plan_function = DatabaseFunction(
            name="query_production_plan",
            description="查询指定日期的生产计划",
            sql_template="""
                SELECT 
                    plan_date,
                    model_name,
                    planned_quantity,
                    actual_quantity,
                    production_line,
                    shift_type,
                    status,
                    CASE 
                        WHEN planned_quantity > 0 THEN ROUND((actual_quantity * 100.0 / planned_quantity), 2)
                        ELSE 0 
                    END as completion_rate
                FROM production_plan 
                WHERE plan_date = :date 
                    AND (:model_name IS NULL OR model_name = :model_name)
                    AND (:production_line IS NULL OR production_line = :production_line)
                ORDER BY shift_type, production_line
            """,
            parameters=[
                FunctionParameter(
                    name="date",
                    type=ParameterType.DATE,
                    description="计划日期，格式：YYYY-MM-DD",
                    validation_regex=r"^\d{4}-\d{2}-\d{2}$"
                ),
                FunctionParameter(
                    name="model_name",
                    type=ParameterType.STRING,
                    description="机型名称（可选），如：iPhone15、SamsungS23等",
                    required=False,
                    default_value=None
                ),
                FunctionParameter(
                    name="production_line",
                    type=ParameterType.STRING,
                    description="生产线（可选），如：Line-A、Line-B、Line-C",
                    required=False,
                    default_value=None
                )
            ],
            category="planning",
            examples=[
                "查询今天的生产计划",
                "看看明天iPhone15的生产安排",
                "Line-A今天的计划完成情况",
                "今日生产计划怎么样"
            ]
        )
        
        # 生产计划统计函数
        production_plan_stats_function = DatabaseFunction(
            name="query_production_plan_stats",
            description="查询指定时间范围的生产计划统计",
            sql_template="""
                SELECT 
                    plan_date,
                    COUNT(*) as total_plans,
                    SUM(planned_quantity) as total_planned,
                    SUM(actual_quantity) as total_actual,
                    ROUND(AVG(CASE 
                        WHEN planned_quantity > 0 THEN (actual_quantity * 100.0 / planned_quantity)
                        ELSE 0 
                    END), 2) as avg_completion_rate,
                    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_plans,
                    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_plans
                FROM production_plan 
                WHERE plan_date BETWEEN :start_date AND :end_date
                    AND (:model_name IS NULL OR model_name = :model_name)
                GROUP BY plan_date
                ORDER BY plan_date DESC
            """,
            parameters=[
                FunctionParameter(
                    name="start_date",
                    type=ParameterType.DATE,
                    description="开始日期，格式：YYYY-MM-DD"
                ),
                FunctionParameter(
                    name="end_date",
                    type=ParameterType.DATE,
                    description="结束日期，格式：YYYY-MM-DD"
                ),
                FunctionParameter(
                    name="model_name",
                    type=ParameterType.STRING,
                    description="机型名称（可选）",
                    required=False,
                    default_value=None
                )
            ],
            category="planning",
            examples=[
                "本周的生产计划统计",
                "最近7天iPhone15的计划完成情况",
                "本月生产计划执行情况"
            ]
        )
        
        # 生产线效率分析函数
        production_line_efficiency_function = DatabaseFunction(
            name="query_production_line_efficiency",
            description="查询生产线效率分析",
            sql_template="""
                SELECT 
                    production_line,
                    shift_type,
                    COUNT(*) as total_tasks,
                    SUM(planned_quantity) as total_planned,
                    SUM(actual_quantity) as total_actual,
                    ROUND(AVG(CASE 
                        WHEN planned_quantity > 0 THEN (actual_quantity * 100.0 / planned_quantity)
                        ELSE 0 
                    END), 2) as avg_efficiency,
                    COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks
                FROM production_plan 
                WHERE plan_date BETWEEN :start_date AND :end_date
                    AND (:production_line IS NULL OR production_line = :production_line)
                GROUP BY production_line, shift_type
                ORDER BY avg_efficiency DESC, production_line, shift_type
            """,
            parameters=[
                FunctionParameter(
                    name="start_date",
                    type=ParameterType.DATE,
                    description="开始日期，格式：YYYY-MM-DD"
                ),
                FunctionParameter(
                    name="end_date",
                    type=ParameterType.DATE,
                    description="结束日期，格式：YYYY-MM-DD"
                ),
                FunctionParameter(
                    name="production_line",
                    type=ParameterType.STRING,
                    description="生产线名称（可选），如：Line-A",
                    required=False,
                    default_value=None
                )
            ],
            category="efficiency",
            examples=[
                "分析各生产线的效率",
                "Line-A本周的生产效率如何",
                "哪条生产线效率最高",
                "生产线效率对比"
            ]
        )

        # 注册内置函数
        self.register_function(defect_rate_function)
        self.register_function(production_stats_function)
        self.register_function(quality_trend_function)
        self.register_function(production_plan_function)
        self.register_function(production_plan_stats_function)
        self.register_function(production_line_efficiency_function)
    
    def register_function(self, function: DatabaseFunction):
        """注册一个新的数据库查询函数"""
        self.functions[function.name] = function
        encoding_fix.safe_print(f"✅ 注册函数: {function.name} - {function.description}")
    
    def get_function(self, name: str) -> Optional[DatabaseFunction]:
        """获取指定名称的函数"""
        return self.functions.get(name)
    
    def list_functions(self) -> List[DatabaseFunction]:
        """获取所有已注册的函数"""
        return list(self.functions.values())
    
    def get_functions_by_category(self, category: str) -> List[DatabaseFunction]:
        """根据分类获取函数"""
        return [func for func in self.functions.values() if func.category == category]
    
    def get_functions_schema(self) -> Dict:
        """获取所有函数的schema，用于AI模型理解"""
        return {
            "functions": [func.to_dict() for func in self.functions.values()],
            "categories": list(set(func.category for func in self.functions.values()))
        }
    
    def parse_parameters(self, function_name: str, raw_params: Dict[str, Any]) -> Dict[str, Any]:
        """解析和验证函数参数"""
        function = self.get_function(function_name)
        if not function:
            raise ValueError(f"未找到函数: {function_name}")
        
        parsed_params = {}
        
        for param in function.parameters:
            raw_value = raw_params.get(param.name)
            
            # 检查必需参数
            if param.required and raw_value is None:
                if param.default_value is not None:
                    parsed_params[param.name] = param.default_value
                else:
                    raise ValueError(f"缺少必需参数: {param.name}")
                continue
            
            # 如果参数为空且不是必需的，使用默认值
            if raw_value is None:
                parsed_params[param.name] = param.default_value
                continue
            
            # 类型转换和验证
            try:
                parsed_value = self._convert_parameter_value(raw_value, param)
                parsed_params[param.name] = parsed_value
            except Exception as e:
                raise ValueError(f"参数 {param.name} 转换失败: {str(e)}")
        
        return parsed_params
    
    def _convert_parameter_value(self, value: Any, param: FunctionParameter) -> Any:
        """转换参数值到指定类型"""
        if value is None:
            return None
        
        # 如果已经是正确类型，直接返回
        if param.type == ParameterType.STRING:
            str_value = str(value)
            if param.validation_regex:
                if not re.match(param.validation_regex, str_value):
                    raise ValueError(f"参数格式不正确，应匹配: {param.validation_regex}")
            return str_value
        
        elif param.type == ParameterType.INTEGER:
            return int(value)
        
        elif param.type == ParameterType.FLOAT:
            return float(value)
        
        elif param.type == ParameterType.BOOLEAN:
            if isinstance(value, bool):
                return value
            if isinstance(value, str):
                return value.lower() in ('true', '1', 'yes', 'on')
            return bool(value)
        
        elif param.type == ParameterType.DATE:
            if isinstance(value, date):
                return value.strftime('%Y-%m-%d')
            elif isinstance(value, datetime):
                return value.strftime('%Y-%m-%d')
            elif isinstance(value, str):
                # 验证日期格式
                if param.validation_regex and not re.match(param.validation_regex, value):
                    raise ValueError(f"日期格式不正确，应为: YYYY-MM-DD")
                return value
            else:
                raise ValueError("无法转换为日期格式")
        
        elif param.type == ParameterType.DATETIME:
            if isinstance(value, datetime):
                return value.strftime('%Y-%m-%d %H:%M:%S')
            elif isinstance(value, str):
                return value
            else:
                raise ValueError("无法转换为日期时间格式")
        
        return value
    
    async def execute_function(self, function_name: str, parameters: Dict[str, Any]) -> Dict[str, Any]:
        """执行数据库查询函数"""
        function = self.get_function(function_name)
        if not function:
            raise ValueError(f"未找到函数: {function_name}")
        
        # 解析参数
        parsed_params = self.parse_parameters(function_name, parameters)
        
        encoding_fix.safe_print(f"🔧 执行函数: {function_name}")
        encoding_fix.safe_print(f"📊 参数: {parsed_params}")
        
        try:
            async with async_session() as session:
                # 执行SQL查询
                result = await session.execute(text(function.sql_template), parsed_params)
                rows = result.mappings().all()
                
                # 转换结果为可序列化的格式
                data = []
                for row in rows:
                    row_dict = dict(row)
                    # 处理日期类型
                    for key, value in row_dict.items():
                        if isinstance(value, (date, datetime)):
                            row_dict[key] = value.isoformat()
                    data.append(row_dict)
                
                encoding_fix.safe_print(f"✅ 查询完成，返回 {len(data)} 条记录")
                
                return {
                    "success": True,
                    "function_name": function_name,
                    "parameters": parsed_params,
                    "data": data,
                    "count": len(data)
                }
                
        except Exception as e:
            encoding_fix.safe_print(f"❌ 函数执行失败: {str(e)}")
            return {
                "success": False,
                "function_name": function_name,
                "parameters": parsed_params,
                "error": str(e),
                "data": []
            }
    
    def match_function_by_query(self, query: str) -> Optional[str]:
        """根据用户查询匹配最适合的函数"""
        query_lower = query.lower()
        
        # 简单的关键词匹配逻辑
        for function in self.functions.values():
            # 检查函数描述
            if any(keyword in query_lower for keyword in function.description.lower().split()):
                return function.name
            
            # 检查示例
            for example in function.examples:
                if any(keyword in query_lower for keyword in example.lower().split()):
                    return function.name
        
        # 特定关键词匹配
        if any(keyword in query_lower for keyword in ['不良率', '质量', 'defect', 'quality']):
            if any(keyword in query_lower for keyword in ['趋势', '变化', 'trend']):
                return 'query_quality_trend'
            else:
                return 'query_defect_rate'
        
        if any(keyword in query_lower for keyword in ['生产', '统计', 'production', 'stats']):
            return 'query_production_stats'
        
        # 生产计划相关匹配
        if any(keyword in query_lower for keyword in ['计划', '安排', 'plan', 'schedule']):
            if any(keyword in query_lower for keyword in ['统计', '汇总', 'stats', 'summary']):
                return 'query_production_plan_stats'
            else:
                return 'query_production_plan'
        
        # 效率分析匹配
        if any(keyword in query_lower for keyword in ['效率', '生产线', 'efficiency', 'line']):
            return 'query_production_line_efficiency'
        
        return None


# 全局函数注册中心实例
function_registry = FunctionRegistry()
