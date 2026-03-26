# services/ai_parameter_extractor.py
"""
AI参数提取器 - 使用LLM智能提取查询参数
"""

import json
import asyncio
from typing import Dict, List, Any, Optional
from datetime import datetime, date, timedelta
from services.function_registry import DatabaseFunction
import encoding_fix
import config


def debug_print(message: str):
    """条件调试打印"""
    if hasattr(config, 'DEBUG_FUNCTION_CALLING') and config.DEBUG_FUNCTION_CALLING:
        encoding_fix.safe_print(message)

class AIParameterExtractor:
    """使用AI模型进行智能参数提取"""
    
    def __init__(self, llm_model=None):
        self.llm_model = llm_model
    
    async def extract_parameters_with_ai(self, query: str, function: DatabaseFunction) -> Dict[str, Any]:
        """使用AI从用户查询中提取函数参数"""
        if not self.llm_model:
            encoding_fix.safe_print("⚠️ LLM模型未初始化，回退到规则提取")
            return {}
        
        debug_print("🤖 [AI思考] 开始使用AI分析用户查询...")
        debug_print(f"🎯 [AI思考] 目标函数: {function.name}")
        debug_print(f"📋 [AI思考] 需要提取的参数: {[p.name + '(' + p.type.value + ')' for p in function.parameters]}")
        
        # 构建参数提取的prompt
        prompt = self._build_parameter_extraction_prompt(query, function)
        
        try:
            debug_print("🧠 [AI思考] 正在调用LLM进行参数提取...")
            # 使用LLM提取参数
            loop = asyncio.get_event_loop()
            response = await loop.run_in_executor(
                None,
                self._generate_parameter_extraction,
                prompt
            )
            
            debug_print(f"📝 [AI思考] LLM原始回复: {response}")
            
            # 解析LLM返回的JSON
            parameters = self._parse_llm_response(response)
            debug_print(f"✅ [AI思考] 解析后的参数: {parameters}")
            
            return parameters
            
        except Exception as e:
            encoding_fix.safe_print(f"❌ AI参数提取失败: {e}")
            return {}
    
    def _build_parameter_extraction_prompt(self, query: str, function: DatabaseFunction) -> str:
        """构建参数提取的prompt"""
        
        # 获取当前日期信息
        today = datetime.now()
        today_str = today.strftime('%Y-%m-%d')
        
        # 构建参数描述
        param_descriptions = []
        for param in function.parameters:
            param_desc = f"- {param.name} ({param.type.value}): {param.description}"
            if not param.required:
                param_desc += " (可选)"
            if param.default_value is not None:
                param_desc += f" (默认值: {param.default_value})"
            param_descriptions.append(param_desc)
        
        prompt = f"""从用户查询中提取参数，只返回JSON格式。

用户查询: "{query}"
当前日期: {today_str}

需要提取的参数:
{chr(10).join(param_descriptions)}

提取规则:
- 日期: "今天"→{today_str}, "昨天"→{(today - timedelta(days=1)).strftime('%Y-%m-%d')}, "上周三"→{(today - timedelta(days=today.weekday() + 7 - 2)).strftime('%Y-%m-%d')}
- 机型: "苹果15"/"iPhone15"→"iPhone15", "三星23"/"SamsungS23"→"SamsungS23"
- 生产线: "A线"/"一线"→"Line-A", "B线"/"二线"→"Line-B"

只返回JSON，无其他文字:
{{
    "date": "2025-10-01",
    "model_name": "iPhone15"
}}"""
        
        return prompt
    
    def _generate_parameter_extraction(self, prompt: str) -> str:
        """使用LLM生成参数提取结果"""
        try:
            response = self.llm_model(
                prompt,
                max_tokens=config.AI_EXTRACTION_MAX_TOKENS,
                temperature=config.AI_EXTRACTION_TEMPERATURE,
                stop=["\n\n", "```"],  # 移除"}"作为stop token，因为JSON需要它
                echo=False
            )
            
            result = response['choices'][0]['text'].strip()
            
            # 确保返回完整的JSON
            if result and not result.endswith('}'):
                result += '}'
            
            return result
            
        except Exception as e:
            encoding_fix.safe_print(f"LLM参数提取生成失败: {e}")
            return "{}"
    
    def _parse_llm_response(self, response: str) -> Dict[str, Any]:
        """解析LLM返回的参数JSON"""
        try:
            # 清理响应文本
            response = response.strip()
            debug_print(f"🔍 [JSON解析] 原始响应: '{response}'")
            
            # 尝试多种方式找到JSON
            json_candidates = []
            
            # 方法1: 找到完整的{}块
            start_idx = response.find('{')
            end_idx = response.rfind('}')
            if start_idx != -1 and end_idx != -1:
                json_candidates.append(response[start_idx:end_idx+1])
            
            # 方法2: 如果响应就是JSON格式
            if response.startswith('{') and response.endswith('}'):
                json_candidates.append(response)
            
            # 方法3: 尝试提取引号中的JSON
            import re
            json_match = re.search(r'\{[^{}]*\}', response)
            if json_match:
                json_candidates.append(json_match.group())
            
            # 尝试解析每个候选JSON
            for json_str in json_candidates:
                try:
                    debug_print(f"🔍 [JSON解析] 尝试解析: '{json_str}'")
                    parameters = json.loads(json_str)
                    
                    # 验证和清理参数
                    cleaned_params = {}
                    for key, value in parameters.items():
                        if value is not None and value != "":
                            cleaned_params[key] = value
                    
                    debug_print(f"✅ [JSON解析] 成功解析: {cleaned_params}")
                    return cleaned_params
                    
                except json.JSONDecodeError:
                    continue
            
            # 如果所有方法都失败了
            debug_print("❌ [JSON解析] 所有解析方法都失败")
            debug_print(f"❌ [JSON解析] 原始响应: {response}")
            return {}
                
        except Exception as e:
            debug_print(f"❌ [JSON解析] 解析失败: {e}")
            return {}


class HybridParameterExtractor:
    """混合参数提取器 - 结合AI和规则提取"""
    
    def __init__(self, llm_model=None):
        from services.function_caller import ParameterExtractor
        self.rule_extractor = ParameterExtractor()
        self.ai_extractor = AIParameterExtractor(llm_model)
    
    async def extract_parameters_for_function(self, query: str, function: DatabaseFunction) -> Dict[str, Any]:
        """混合提取参数 - 优先使用AI，回退到规则"""
        
        debug_print("🔄 [混合提取] 开始混合参数提取策略...")
        
        # 1. 先尝试AI提取
        debug_print("🤖 [混合提取] 步骤1: 使用AI提取参数")
        ai_params = await self.ai_extractor.extract_parameters_with_ai(query, function)
        
        # 2. 使用规则提取作为补充
        debug_print("📋 [混合提取] 步骤2: 使用规则提取作为补充")
        rule_params = self.rule_extractor.extract_parameters_for_function(query, function)
        
        # 3. 合并结果，AI优先
        debug_print("🔀 [混合提取] 步骤3: 合并AI和规则提取结果")
        final_params = {}
        final_params.update(rule_params)  # 先添加规则提取的结果
        final_params.update(ai_params)    # AI结果覆盖规则结果
        
        debug_print(f"📋 [混合提取] 规则提取结果: {rule_params}")
        debug_print(f"🤖 [混合提取] AI提取结果: {ai_params}")
        debug_print(f"✅ [混合提取] 最终合并结果: {final_params}")
        
        # 分析提取效果
        if ai_params and rule_params:
            debug_print("💡 [混合提取] AI和规则都成功提取到参数，使用AI结果优先")
        elif ai_params:
            debug_print("🤖 [混合提取] 仅AI成功提取参数，规则提取失败")
        elif rule_params:
            debug_print("📋 [混合提取] 仅规则成功提取参数，AI提取失败")
        else:
            debug_print("❌ [混合提取] AI和规则都未能提取到参数")
        
        return final_params
