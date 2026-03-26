"""
AI驱动的Function Calling处理器
让AI自己决定调用什么函数，传递什么参数
"""
import json
import asyncio
from typing import Dict, Any, List, Optional
from datetime import datetime, timedelta

from services.function_definitions import get_function_definitions, get_current_date_info
try:
    from services.function_registry import function_registry
except ImportError as e:
    print(f"Warning: Could not import function_registry: {e}")
    function_registry = None
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
from config import DEBUG_FUNCTION_CALLING

def debug_print(message: str):
    """条件调试打印"""
    if DEBUG_FUNCTION_CALLING:
        encoding_fix.safe_print(message)

class AIFunctionCaller:
    """AI驱动的函数调用器"""
    
    def __init__(self, llm_model):
        self.llm_model = llm_model
        self.function_definitions = get_function_definitions()
        
        # 检查function_registry是否可用
        if function_registry is None:
            debug_print("⚠️ [AI Function Calling] function_registry不可用，功能受限")
            self.available = False
        else:
            self.available = True
        
    def _build_function_calling_prompt(self, user_query: str) -> str:
        """构建Function Calling的系统提示"""
        
        current_date_info = get_current_date_info()
        
        # 构建函数描述
        functions_desc = []
        for func_def in self.function_definitions:
            func = func_def["function"]
            name = func["name"]
            desc = func["description"]
            params = func["parameters"]["properties"]
            required = func["parameters"]["required"]
            
            param_desc = []
            for param_name, param_info in params.items():
                param_type = param_info["type"]
                param_description = param_info["description"]
                is_required = param_name in required
                required_text = " (必需)" if is_required else " (可选)"
                param_desc.append(f"  - {param_name} ({param_type}){required_text}: {param_description}")
            
            functions_desc.append(f"""
函数名: {name}
描述: {desc}
参数:
{chr(10).join(param_desc)}""")
        
        return f"""你是一个智能助手，可以调用以下函数来查询生产数据：

{current_date_info}

可用函数列表（你只能从这个列表中选择）：
{chr(10).join(functions_desc)}

用户问题: {user_query}

请分析用户的问题，如果需要查询数据库，请选择合适的函数并提取正确的参数。

⚠️ 严格规则：
1. 你只能调用上面列出的函数，不能创造或使用其他函数名
2. 如果用户问题无法用上述任何函数解决，必须返回 function_call: null
3. 相对时间需要转换为具体日期，如"上周三"需要计算出具体的YYYY-MM-DD日期
4. 机型名称需要标准化，如"苹果15"→"iPhone15"，"华为P60"→"HuaweiP60"
5. 如果用户询问的数据类型（如销量、库存等）不在可用函数范围内，返回 null

请按以下格式回答：

如果需要调用函数：
```json
{{
    "function_call": {{
        "name": "函数名（必须从上面列表中选择）",
        "arguments": {{
            "参数名": "参数值"
        }}
    }},
    "reasoning": "选择此函数的原因和参数提取逻辑"
}}
```

如果不需要调用函数（包括问题超出函数范围的情况）：
```json
{{
    "function_call": null,
    "reasoning": "不需要查询数据库的原因，或者没有合适的函数可用"
}}
```
"""

    async def analyze_and_call(self, user_query: str) -> Dict[str, Any]:
        """分析用户问题并决定是否调用函数"""
        debug_print(f"\n🤖 [AI Function Calling] 开始分析用户问题: {user_query}")
        
        # 检查是否可用
        if not self.available:
            debug_print("❌ [AI Function Calling] 功能不可用，跳过")
            return {"needs_function_call": False, "reason": "function_calling_unavailable"}
        
        try:
            # 1. 让AI分析问题并决定是否调用函数
            prompt = self._build_function_calling_prompt(user_query)
            debug_print("🧠 [AI思考] 发送Function Calling分析请求给LLM...")
            
            # 调用LLM进行分析
            loop = asyncio.get_event_loop()
            ai_response = await loop.run_in_executor(
                None,
                self._generate_response,
                prompt
            )
            
            debug_print(f"🤖 [AI回答] LLM分析结果: {ai_response}")
            
            # 2. 解析AI的回答
            function_call_info = self._parse_ai_response(ai_response)
            
            if not function_call_info.get("function_call"):
                debug_print("💭 [AI决策] AI认为不需要调用函数")
                return {
                    "needs_function_call": False,
                    "reasoning": function_call_info.get("reasoning", "AI判断不需要函数调用"),
                    "ai_analysis": ai_response
                }
            
            # 3. 执行AI选择的函数
            function_call = function_call_info["function_call"]
            function_name = function_call["name"]
            arguments = function_call["arguments"]
            reasoning = function_call_info.get("reasoning", "")
            
            debug_print(f"🎯 [AI选择] 函数: {function_name}")
            debug_print(f"📝 [AI参数] 参数: {arguments}")
            debug_print(f"💡 [AI推理] 原因: {reasoning}")
            
            # 4. 执行函数调用
            result = await self._execute_function(function_name, arguments)
            
            if result["success"]:
                debug_print(f"✅ [函数执行] 成功执行 {function_name}")
                return {
                    "needs_function_call": True,
                    "function_called": True,
                    "function_name": function_name,
                    "arguments": arguments,
                    "result": result,
                    "reasoning": reasoning,
                    "ai_analysis": ai_response
                }
            else:
                debug_print(f"❌ [函数执行] 执行失败: {result.get('error')}")
                return {
                    "needs_function_call": True,
                    "function_called": False,
                    "error": result.get("error"),
                    "reasoning": reasoning,
                    "ai_analysis": ai_response
                }
                
        except Exception as e:
            debug_print(f"❌ [AI Function Calling] 处理失败: {e}")
            return {
                "needs_function_call": False,
                "error": str(e),
                "fallback_to_knowledge_base": True
            }
    
    def _generate_response(self, prompt: str) -> str:
        """调用LLM生成回答"""
        try:
            return self.llm_model(
                prompt,
                max_tokens=500,
                temperature=0.1,
                stop=None
            )
        except Exception as e:
            debug_print(f"❌ [LLM调用] 生成回答失败: {e}")
            return "ERROR: LLM调用失败"
    
    def _parse_ai_response(self, ai_response) -> Dict[str, Any]:
        """解析AI的回答，提取函数调用信息"""
        try:
            # 从LLM响应中提取文本
            if isinstance(ai_response, dict):
                if 'choices' in ai_response and len(ai_response['choices']) > 0:
                    response_text = ai_response['choices'][0].get('text', '')
                else:
                    debug_print(f"❌ [解析错误] 无效的LLM响应格式: {ai_response}")
                    return {"function_call": None, "reasoning": "LLM响应格式错误"}
            else:
                response_text = str(ai_response)
            
            debug_print(f"📝 [解析] 提取到响应文本: {response_text}")
            
            # 查找JSON代码块
            import re
            json_match = re.search(r'```json\s*(\{.*?\})\s*```', response_text, re.DOTALL)
            
            if json_match:
                json_str = json_match.group(1)
                debug_print(f"📋 [解析] 提取到JSON: {json_str}")
                
                # 移除JSON中的注释（// 风格的注释）
                import re
                json_str_clean = re.sub(r'//.*?(?=\n|$)', '', json_str, flags=re.MULTILINE)
                debug_print(f"🧹 [清理] 移除注释后: {json_str_clean}")
                
                return json.loads(json_str_clean)
            else:
                # 尝试直接解析整个回答
                debug_print("🔍 [解析] 未找到JSON代码块，尝试直接解析")
                # 移除可能的注释
                response_clean = re.sub(r'//.*?(?=\n|$)', '', response_text.strip(), flags=re.MULTILINE)
                return json.loads(response_clean)
                
        except json.JSONDecodeError as e:
            debug_print(f"❌ [解析错误] JSON解析失败: {e}")
            debug_print(f"原始回答: {response_text if 'response_text' in locals() else ai_response}")
            return {"function_call": None, "reasoning": "AI回答格式错误"}
        except Exception as e:
            debug_print(f"❌ [解析错误] 解析失败: {e}")
            return {"function_call": None, "reasoning": f"解析错误: {str(e)}"}
    
    async def _execute_function(self, function_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
        """执行指定的函数"""
        try:
            debug_print(f"🔧 [函数执行] 准备执行 {function_name}({arguments})")
            
            # 使用现有的function_registry执行函数
            if function_registry is None:
                raise Exception("function_registry不可用")
            
            # ✅ 验证函数是否存在
            available_functions = [func["function"]["name"] for func in self.function_definitions]
            if function_name not in available_functions:
                error_msg = f"函数 '{function_name}' 不存在。可用函数: {', '.join(available_functions)}"
                debug_print(f"❌ [函数验证] {error_msg}")
                return {
                    "success": False,
                    "error": error_msg,
                    "function_name": function_name,
                    "parameters": arguments,
                    "available_functions": available_functions
                }
            
            result = await function_registry.execute_function(function_name, arguments)
            
            debug_print(f"📊 [执行结果] 查询到 {result.get('count', 0)} 条记录")
            return result
            
        except Exception as e:
            debug_print(f"❌ [函数执行] 执行失败: {e}")
            return {
                "success": False,
                "error": str(e),
                "function_name": function_name,
                "parameters": arguments
            }
    
    def format_database_result(self, result: Dict[str, Any]) -> str:
        """格式化数据库查询结果"""
        if not result.get("success"):
            return f"查询失败: {result.get('error', '未知错误')}"
        
        function_name = result.get("function_name", "")
        
        # 使用现有的格式化逻辑
        if function_name == "query_defect_rate":
            return self._format_defect_rate_result(result)
        elif function_name == "query_production_stats":
            return self._format_production_stats_result(result)
        elif function_name == "query_quality_trend":
            return self._format_quality_trend_result(result)
        elif function_name == "query_production_plan":
            return self._format_production_plan_result(result)
        elif function_name == "query_production_plan_stats":
            return self._format_production_plan_stats_result(result)
        elif function_name == "query_production_line_efficiency":
            return self._format_production_line_efficiency_result(result)
        else:
            return f"查询结果: {result.get('data', [])}"
    
    def _format_defect_rate_result(self, result: Dict[str, Any]) -> str:
        """格式化不良率查询结果"""
        data = result.get("data", [])
        if not data:
            return "未找到相关数据"
        
        formatted_lines = []
        for row in data:
            # 处理字典格式的数据
            if isinstance(row, dict):
                date = row.get('date')
                model_name = row.get('model_name')
                total_count = row.get('total_count')
                defect_count = row.get('defect_count')
                defect_rate = row.get('defect_rate')
            else:
                # 兼容原来的数组格式
                date = row[0]
                model_name = row[1] 
                total_count = row[2]
                defect_count = row[3]
                defect_rate = row[4]
            
            formatted_lines.append(f"""📅 日期: {date}
📱 机型: {model_name}
🏭 总产量: {total_count:,} 台
❌ 不良品: {defect_count:,} 台  
📈 不良率: {defect_rate}%""")
        
        return "\n\n".join(formatted_lines)
    
    def _format_production_stats_result(self, result: Dict[str, Any]) -> str:
        """格式化生产统计结果"""
        data = result.get("data", [])
        if not data:
            return "未找到相关数据"
        
        formatted_lines = []
        for row in data:
            # 处理字典格式的数据
            if isinstance(row, dict):
                date = row.get('date')
                model_name = row.get('model_name')
                total_production = row.get('total_production')
                total_defects = row.get('total_defects')
                avg_defect_rate = row.get('avg_defect_rate')
            else:
                # 兼容原来的数组格式
                date = row[0] if len(row) > 0 else None
                model_name = row[1] if len(row) > 1 else None
                total_production = row[2] if len(row) > 2 else None
                total_defects = row[3] if len(row) > 3 else None
                avg_defect_rate = row[4] if len(row) > 4 else None
            
            formatted_lines.append(f"""📅 日期: {date}
📱 机型: {model_name}
🏭 总产量: {total_production:,} 台
❌ 总不良品: {total_defects:,} 台
📊 平均不良率: {avg_defect_rate}%""")
        
        return "\n\n".join(formatted_lines)
    
    def _format_quality_trend_result(self, result: Dict[str, Any]) -> str:
        """格式化质量趋势结果"""
        data = result.get("data", [])
        if not data:
            return "未找到相关数据"
        
        formatted_lines = []
        for row in data:
            # 处理字典格式的数据
            if isinstance(row, dict):
                date = row.get('date')
                model_name = row.get('model_name')
                defect_rate = row.get('defect_rate')
            else:
                date = row[0]
                model_name = row[1]
                defect_rate = row[2]
            
            formatted_lines.append(f"📅 {date} - 📱 {model_name}: {defect_rate}%")
        
        return "\n".join(formatted_lines)
    
    def _format_production_plan_result(self, result: Dict[str, Any]) -> str:
        """格式化生产计划结果"""
        data = result.get("data", [])
        if not data:
            return "未找到相关数据"
        
        formatted_lines = []
        for row in data:
            # 处理字典格式的数据
            if isinstance(row, dict):
                date = row.get('plan_date')
                model_name = row.get('model_name')
                planned_qty = row.get('planned_quantity')
                actual_qty = row.get('actual_quantity')
                production_line = row.get('production_line')
                shift_type = row.get('shift_type')
                status = row.get('status')
                completion_rate = row.get('completion_rate', 0)
            else:
                date = row[0]
                model_name = row[1]
                planned_qty = row[2]
                actual_qty = row[3]
                production_line = row[4]
                shift_type = row[5]
                status = row[6]
                completion_rate = (actual_qty / planned_qty * 100) if planned_qty > 0 else 0
            
            formatted_lines.append(f"""📅 日期: {date}
📱 机型: {model_name}
🎯 计划产量: {planned_qty:,} 台
✅ 实际产量: {actual_qty:,} 台
📊 完成率: {completion_rate:.1f}%
🏭 生产线: {production_line}
⏰ 班次: {shift_type}
📋 状态: {status}""")
        
        return "\n\n".join(formatted_lines)
    
    def _format_production_plan_stats_result(self, result: Dict[str, Any]) -> str:
        """格式化生产计划统计结果"""
        data = result.get("data", [])
        if not data:
            return "未找到相关数据"
        
        formatted_lines = []
        for row in data:
            # 处理字典格式的数据
            if isinstance(row, dict):
                plan_date = row.get('plan_date')
                total_planned = row.get('total_planned')
                total_actual = row.get('total_actual')
                avg_completion_rate = row.get('avg_completion_rate')
                
                formatted_lines.append(f"""📅 日期: {plan_date}
🎯 计划总产量: {total_planned:,} 台
✅ 实际总产量: {total_actual:,} 台
📊 平均完成率: {avg_completion_rate}%""")
            else:
                production_line = row[0]
                total_planned = row[1]
                total_actual = row[2]
                completion_rate = row[3]
                
                formatted_lines.append(f"""🏭 生产线: {production_line}
🎯 计划总产量: {total_planned:,} 台
✅ 实际总产量: {total_actual:,} 台
📊 完成率: {completion_rate}%""")
        
        return "\n\n".join(formatted_lines)
    
    def _format_production_line_efficiency_result(self, result: Dict[str, Any]) -> str:
        """格式化生产线效率结果"""
        data = result.get("data", [])
        if not data:
            return "未找到相关数据"
        
        formatted_lines = []
        for row in data:
            # 处理字典格式的数据
            if isinstance(row, dict):
                production_line = row.get('production_line')
                shift_type = row.get('shift_type')
                total_planned = row.get('total_planned')
                total_actual = row.get('total_actual')
                avg_efficiency = row.get('avg_efficiency')
                
                formatted_lines.append(f"""🏭 生产线: {production_line} - ⏰ {shift_type}
🎯 计划: {total_planned:,} 台
✅ 实际: {total_actual:,} 台  
⚡ 平均效率: {avg_efficiency}%""")
            else:
                date = row[0]
                production_line = row[1]
                planned_qty = row[2]
                actual_qty = row[3]
                efficiency = row[4]
                
                formatted_lines.append(f"""📅 {date} - 🏭 {production_line}
🎯 计划: {planned_qty:,} 台
✅ 实际: {actual_qty:,} 台  
⚡ 效率: {efficiency}%""")
        
        return "\n".join(formatted_lines)
