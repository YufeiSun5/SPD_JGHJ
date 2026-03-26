# services/function_caller.py
"""
智能函数调用器 - 分析用户问题并决定是否需要调用数据库查询函数
支持参数提取、函数匹配和智能调用
"""

import re
import json
import asyncio
from typing import Dict, List, Any, Optional, Tuple
from datetime import datetime, date, timedelta
from services.function_registry import function_registry, DatabaseFunction
import encoding_fix
import config


def debug_print(message: str):
    """条件调试打印"""
    if config.DEBUG_FUNCTION_CALLING:
        encoding_fix.safe_print(message)


class ParameterExtractor:
    """参数提取器 - 从用户问题中提取函数参数"""
    
    def __init__(self):
        # 日期相关的正则表达式
        self.date_patterns = {
            'today': r'今天|今日|当天',
            'yesterday': r'昨天|昨日',
            'specific_date': r'(\d{4}[-/]\d{1,2}[-/]\d{1,2})',
            'relative_days': r'(\d+)天前',
            'this_week': r'本周|这周',
            'this_month': r'本月|这个月',
            'last_week': r'上周|上星期',
            'last_month': r'上月|上个月',
            'last_week_day': r'上周([一二三四五六日])',
            'this_week_day': r'本周([一二三四五六日])'
        }
        
        # 机型相关的正则表达式
        self.model_patterns = {
            'iphone': r'(?:iPhone|苹果|apple)\s*(\d+)',
            'samsung': r'(?:Samsung|三星)\s*S?(\d+)',
            'huawei': r'华为|Huawei',
            'xiaomi': r'小米|Xiaomi',
            'oppo': r'OPPO',
            'vivo': r'VIVO'
        }
        
        # 生产线相关的正则表达式
        self.production_line_patterns = {
            'line_a': r'Line[-\s]*A|A线|一线',
            'line_b': r'Line[-\s]*B|B线|二线',
            'line_c': r'Line[-\s]*C|C线|三线',
            'line_d': r'Line[-\s]*D|D线|四线'
        }
    
    def extract_date_parameter(self, query: str) -> Optional[str]:
        """从查询中提取日期参数"""
        query_lower = query.lower()
        
        # 今天
        if re.search(self.date_patterns['today'], query):
            return datetime.now().strftime('%Y-%m-%d')
        
        # 昨天
        if re.search(self.date_patterns['yesterday'], query):
            return (datetime.now() - timedelta(days=1)).strftime('%Y-%m-%d')
        
        # 具体日期
        date_match = re.search(self.date_patterns['specific_date'], query)
        if date_match:
            date_str = date_match.group(1)
            # 标准化日期格式
            date_str = date_str.replace('/', '-')
            return date_str
        
        # N天前
        days_ago_match = re.search(self.date_patterns['relative_days'], query)
        if days_ago_match:
            days = int(days_ago_match.group(1))
            return (datetime.now() - timedelta(days=days)).strftime('%Y-%m-%d')
        
        # 上周具体星期几
        last_week_day_match = re.search(self.date_patterns['last_week_day'], query)
        if last_week_day_match:
            day_name = last_week_day_match.group(1)
            weekday_map = {'一': 0, '二': 1, '三': 2, '四': 3, '五': 4, '六': 5, '日': 6}
            if day_name in weekday_map:
                target_weekday = weekday_map[day_name]
                today = datetime.now()
                # 计算上周指定星期几的日期
                days_back = today.weekday() + 7 - target_weekday
                target_date = today - timedelta(days=days_back)
                return target_date.strftime('%Y-%m-%d')
        
        # 本周具体星期几
        this_week_day_match = re.search(self.date_patterns['this_week_day'], query)
        if this_week_day_match:
            day_name = this_week_day_match.group(1)
            weekday_map = {'一': 0, '二': 1, '三': 2, '四': 3, '五': 4, '六': 5, '日': 6}
            if day_name in weekday_map:
                target_weekday = weekday_map[day_name]
                today = datetime.now()
                # 计算本周指定星期几的日期
                days_diff = target_weekday - today.weekday()
                target_date = today + timedelta(days=days_diff)
                return target_date.strftime('%Y-%m-%d')
        
        return None
    
    def extract_date_range_parameters(self, query: str) -> Tuple[Optional[str], Optional[str]]:
        """从查询中提取日期范围参数"""
        query_lower = query.lower()
        today = datetime.now()
        
        # 本周
        if re.search(self.date_patterns['this_week'], query):
            start_of_week = today - timedelta(days=today.weekday())
            end_of_week = start_of_week + timedelta(days=6)
            return start_of_week.strftime('%Y-%m-%d'), end_of_week.strftime('%Y-%m-%d')
        
        # 本月
        if re.search(self.date_patterns['this_month'], query):
            start_of_month = today.replace(day=1)
            if today.month == 12:
                end_of_month = today.replace(year=today.year+1, month=1, day=1) - timedelta(days=1)
            else:
                end_of_month = today.replace(month=today.month+1, day=1) - timedelta(days=1)
            return start_of_month.strftime('%Y-%m-%d'), end_of_month.strftime('%Y-%m-%d')
        
        # 上周
        if re.search(self.date_patterns['last_week'], query):
            start_of_last_week = today - timedelta(days=today.weekday() + 7)
            end_of_last_week = start_of_last_week + timedelta(days=6)
            return start_of_last_week.strftime('%Y-%m-%d'), end_of_last_week.strftime('%Y-%m-%d')
        
        # 最近N天
        recent_days_match = re.search(r'最近(\d+)天|近(\d+)天', query)
        if recent_days_match:
            days = int(recent_days_match.group(1) or recent_days_match.group(2))
            start_date = today - timedelta(days=days)
            return start_date.strftime('%Y-%m-%d'), today.strftime('%Y-%m-%d')
        
        return None, None
    
    def extract_model_parameter(self, query: str) -> Optional[str]:
        """从查询中提取机型参数"""
        # iPhone
        iphone_match = re.search(self.model_patterns['iphone'], query, re.IGNORECASE)
        if iphone_match:
            return f"iPhone{iphone_match.group(1)}"
        
        # Samsung
        samsung_match = re.search(self.model_patterns['samsung'], query, re.IGNORECASE)
        if samsung_match:
            return f"SamsungS{samsung_match.group(1)}"
        
        # 其他品牌的简单匹配
        for brand, pattern in self.model_patterns.items():
            if brand not in ['iphone', 'samsung']:
                if re.search(pattern, query, re.IGNORECASE):
                    return brand.capitalize()
        
        return None
    
    def extract_production_line_parameter(self, query: str) -> Optional[str]:
        """从查询中提取生产线参数"""
        for line_key, pattern in self.production_line_patterns.items():
            if re.search(pattern, query, re.IGNORECASE):
                return f"Line-{line_key.split('_')[1].upper()}"
        return None
    
    def extract_parameters_for_function(self, query: str, function: DatabaseFunction) -> Dict[str, Any]:
        """为特定函数提取参数"""
        parameters = {}
        
        for param in function.parameters:
            if param.name == 'date':
                date_value = self.extract_date_parameter(query)
                if date_value:
                    parameters[param.name] = date_value
            
            elif param.name in ['start_date', 'end_date']:
                start_date, end_date = self.extract_date_range_parameters(query)
                if param.name == 'start_date' and start_date:
                    parameters[param.name] = start_date
                elif param.name == 'end_date' and end_date:
                    parameters[param.name] = end_date
            
            elif param.name == 'model_name':
                model_value = self.extract_model_parameter(query)
                if model_value:
                    parameters[param.name] = model_value
            
            elif param.name == 'production_line':
                line_value = self.extract_production_line_parameter(query)
                if line_value:
                    parameters[param.name] = line_value
        
        return parameters


class QueryAnalyzer:
    """查询分析器 - 分析用户问题的意图"""
    
    def __init__(self):
        self.database_query_keywords = [
            '查询', '查看', '看看', '统计', '分析', '数据', 
            '不良率', '质量', '生产', '趋势', '情况', '计划', '安排', '效率',
            'query', 'check', 'analyze', 'data', 'defect', 'quality', 'production', 'plan', 'efficiency'
        ]
        
        self.knowledge_base_keywords = [
            '什么是', '如何', '怎么', '为什么', '介绍', '说明', '解释',
            'what', 'how', 'why', 'explain', 'describe'
        ]
    
    def needs_database_query(self, query: str) -> bool:
        """判断是否需要数据库查询"""
        query_lower = query.lower()
        
        # 检查是否包含数据库查询关键词
        has_db_keywords = any(keyword in query_lower for keyword in self.database_query_keywords)
        
        # 检查是否包含具体的数据指标
        has_data_indicators = any(indicator in query_lower for indicator in [
            '不良率', '生产量', '统计', '数据', '报表', '趋势', '计划', '安排', '效率', '完成率'
        ])
        
        # 检查是否包含时间相关词汇
        has_time_indicators = any(time_word in query_lower for time_word in [
            '今天', '昨天', '本周', '本月', '最近', '日期'
        ])
        
        # 如果同时满足多个条件，很可能需要数据库查询
        score = sum([has_db_keywords, has_data_indicators, has_time_indicators])
        
        return score >= 2 or (has_db_keywords and has_data_indicators)
    
    def is_knowledge_base_query(self, query: str) -> bool:
        """判断是否是知识库查询"""
        query_lower = query.lower()
        
        # 检查是否包含知识库查询关键词
        has_kb_keywords = any(keyword in query_lower for keyword in self.knowledge_base_keywords)
        
        # 检查是否是概念性问题
        is_conceptual = any(pattern in query_lower for pattern in [
            '是什么', '什么意思', '如何理解', '原理', '定义'
        ])
        
        return has_kb_keywords or is_conceptual


class FunctionCaller:
    """智能函数调用器"""
    
    def __init__(self, llm_model=None):
        self.parameter_extractor = ParameterExtractor()
        self.query_analyzer = QueryAnalyzer()
        
        # 如果提供了LLM模型，启用AI参数提取
        if llm_model:
            try:
                from services.ai_parameter_extractor import HybridParameterExtractor
                self.hybrid_extractor = HybridParameterExtractor(llm_model)
                self.use_ai_extraction = True
                encoding_fix.safe_print("✅ 启用AI参数提取")
            except ImportError:
                self.use_ai_extraction = False
                encoding_fix.safe_print("⚠️ AI参数提取模块未找到，使用规则提取")
        else:
            self.use_ai_extraction = False
    
    async def analyze_and_call(self, query: str) -> Dict[str, Any]:
        """分析用户问题并决定是否调用函数"""
        debug_print(f"\n🔍 [函数调用分析] 分析用户问题: {query}")
        
        # 1. 判断是否需要数据库查询
        debug_print("🧠 [思考过程] 开始分析用户意图...")
        needs_db_query = self.query_analyzer.needs_database_query(query)
        is_kb_query = self.query_analyzer.is_knowledge_base_query(query)
        
        debug_print(f"📊 [意图分析] 需要数据库查询: {needs_db_query}, 知识库查询: {is_kb_query}")
        
        # 详细分析过程
        if needs_db_query:
            debug_print("💡 [思考] 检测到数据库查询关键词，准备匹配查询函数")
        else:
            debug_print("💡 [思考] 未检测到数据库查询需求，可能是知识库问题")
        
        if not needs_db_query:
            return {
                "needs_function_call": False,
                "reason": "问题不需要数据库查询",
                "query_type": "knowledge_base" if is_kb_query else "general"
            }
        
        # 2. 匹配合适的函数
        debug_print("🎯 [思考] 开始匹配合适的查询函数...")
        function_name = function_registry.match_function_by_query(query)
        if not function_name:
            debug_print("❌ [函数匹配] 未找到合适的函数")
            debug_print("💭 [思考] 用户问题可能包含数据库查询意图，但没有匹配的函数")
            return {
                "needs_function_call": False,
                "reason": "未找到合适的数据库查询函数",
                "query_type": "unsupported_database_query"
            }
        
        function = function_registry.get_function(function_name)
        debug_print(f"✅ [函数匹配] 匹配到函数: {function_name} - {function.description}")
        debug_print(f"💡 [思考] 选择了 {function_name} 函数，需要提取以下参数: {[p.name for p in function.parameters]}")
        
        # 3. 提取参数 (使用AI或规则)
        if self.use_ai_extraction:
            parameters = await self.hybrid_extractor.extract_parameters_for_function(query, function)
            debug_print(f"🤖 [AI参数提取] 提取到参数: {parameters}")
        else:
            parameters = self.parameter_extractor.extract_parameters_for_function(query, function)
            debug_print(f"📋 [规则参数提取] 提取到参数: {parameters}")
        
        # 4. 检查必需参数是否完整
        debug_print("🔍 [思考] 检查参数完整性...")
        missing_params = []
        for param in function.parameters:
            if param.required and param.name not in parameters:
                if param.default_value is None:
                    missing_params.append(param.name)
        
        if missing_params:
            debug_print(f"⚠️ [参数检查] 缺少必需参数: {missing_params}")
            debug_print(f"💭 [思考] 无法执行查询，需要用户提供更多信息: {missing_params}")
            return {
                "needs_function_call": False,
                "reason": f"缺少必需参数: {', '.join(missing_params)}",
                "function_name": function_name,
                "extracted_parameters": parameters,
                "missing_parameters": missing_params,
                "query_type": "incomplete_database_query"
            }
        
        debug_print("✅ [参数检查] 所有必需参数已提取完整")
        
        # 5. 执行函数调用
        try:
            debug_print(f"🚀 [函数执行] 开始执行函数: {function_name}")
            result = await function_registry.execute_function(function_name, parameters)
            
            return {
                "needs_function_call": True,
                "function_called": True,
                "function_name": function_name,
                "parameters": parameters,
                "result": result,
                "query_type": "database_query"
            }
            
        except Exception as e:
            debug_print(f"❌ [函数执行] 执行失败: {str(e)}")
            return {
                "needs_function_call": True,
                "function_called": False,
                "function_name": function_name,
                "parameters": parameters,
                "error": str(e),
                "query_type": "failed_database_query"
            }
    
    def format_database_result(self, result: Dict[str, Any]) -> str:
        """格式化数据库查询结果为用户友好的文本"""
        if not result.get("success", False):
            return f"查询失败: {result.get('error', '未知错误')}"
        
        data = result.get("data", [])
        if not data:
            return "查询完成，但没有找到相关数据。"
        
        function_name = result.get("function_name", "")
        
        # 根据不同的函数类型格式化结果
        if function_name == "query_defect_rate":
            return self._format_defect_rate_result(data)
        elif function_name == "query_production_stats":
            return self._format_production_stats_result(data)
        elif function_name == "query_quality_trend":
            return self._format_quality_trend_result(data)
        elif function_name == "query_production_plan":
            return self._format_production_plan_result(data)
        elif function_name == "query_production_plan_stats":
            return self._format_production_plan_stats_result(data)
        elif function_name == "query_production_line_efficiency":
            return self._format_production_line_efficiency_result(data)
        else:
            return self._format_generic_result(data)
    
    def _format_defect_rate_result(self, data: List[Dict]) -> str:
        """格式化不良率查询结果"""
        if not data:
            return "未找到指定日期和机型的不良率数据。"
        
        result_text = "📊 不良率查询结果：\n\n"
        for item in data:
            defect_rate = item.get('defect_rate', 0)
            
            # 质量等级评估
            if defect_rate <= 1.0:
                quality_level = "🟢 优秀"
                quality_desc = "质量表现优异"
            elif defect_rate <= 2.0:
                quality_level = "🟡 良好"
                quality_desc = "质量表现良好"
            elif defect_rate <= 3.0:
                quality_level = "🟠 一般"
                quality_desc = "质量需要关注"
            else:
                quality_level = "🔴 较差"
                quality_desc = "质量需要改进"
            
            result_text += f"📅 日期: {item.get('date', 'N/A')}\n"
            result_text += f"📱 机型: {item.get('model_name', 'N/A')}\n"
            result_text += f"🏭 总产量: {item.get('total_count', 0):,} 台\n"
            result_text += f"❌ 不良品: {item.get('defect_count', 0):,} 台\n"
            result_text += f"📈 不良率: {defect_rate}%\n"
            result_text += f"🎯 质量等级: {quality_level} ({quality_desc})\n"
            result_text += "─" * 30 + "\n"
        
        return result_text
    
    def _format_production_stats_result(self, data: List[Dict]) -> str:
        """格式化生产统计结果"""
        if not data:
            return "未找到指定时间范围的生产统计数据。"
        
        result_text = "📊 生产统计结果：\n\n"
        total_production = 0
        total_defects = 0
        
        for item in data:
            production = item.get('total_production', 0)
            defects = item.get('total_defects', 0)
            total_production += production
            total_defects += defects
            
            result_text += f"📅 日期: {item.get('date', 'N/A')}\n"
            result_text += f"📱 机型: {item.get('model_name', 'N/A')}\n"
            result_text += f"🏭 总产量: {production:,} 台\n"
            result_text += f"❌ 不良品: {defects:,} 台\n"
            result_text += f"📈 平均不良率: {item.get('avg_defect_rate', 0)}%\n"
            result_text += "─" * 30 + "\n"
        
        # 添加汇总信息
        if len(data) > 1:
            overall_defect_rate = (total_defects / total_production * 100) if total_production > 0 else 0
            result_text += f"\n📋 汇总统计：\n"
            result_text += f"🏭 总产量: {total_production:,} 台\n"
            result_text += f"❌ 总不良品: {total_defects:,} 台\n"
            result_text += f"📈 整体不良率: {overall_defect_rate:.2f}%\n"
        
        return result_text
    
    def _format_quality_trend_result(self, data: List[Dict]) -> str:
        """格式化质量趋势结果"""
        if not data:
            return "未找到指定机型的质量趋势数据。"
        
        result_text = "📈 质量趋势分析：\n\n"
        
        for i, item in enumerate(data):
            defect_rate = item.get('defect_rate', 0)
            prev_rate = item.get('prev_defect_rate')
            
            result_text += f"📅 {item.get('date', 'N/A')}: {defect_rate}%"
            
            if prev_rate is not None:
                change = defect_rate - prev_rate
                if change > 0:
                    result_text += f" (↗️ +{change:.2f}%)"
                elif change < 0:
                    result_text += f" (↘️ {change:.2f}%)"
                else:
                    result_text += f" (➡️ 持平)"
            
            result_text += "\n"
        
        # 添加趋势分析
        if len(data) >= 2:
            first_rate = data[0].get('defect_rate', 0)
            last_rate = data[-1].get('defect_rate', 0)
            overall_change = last_rate - first_rate
            
            result_text += f"\n📊 趋势分析：\n"
            if overall_change > 0:
                result_text += f"📈 整体趋势上升 (+{overall_change:.2f}%)\n"
            elif overall_change < 0:
                result_text += f"📉 整体趋势下降 ({overall_change:.2f}%)\n"
            else:
                result_text += f"➡️ 整体趋势平稳\n"
        
        return result_text
    
    def _format_production_plan_result(self, data: List[Dict]) -> str:
        """格式化生产计划查询结果"""
        if not data:
            return "未找到指定日期的生产计划数据。"
        
        result_text = "📋 生产计划查询结果：\n\n"
        
        for item in data:
            completion_rate = item.get('completion_rate', 0)
            status = item.get('status', 'unknown')
            status_icon = {
                'pending': '⏳',
                'in_progress': '🔄', 
                'completed': '✅',
                'cancelled': '❌'
            }.get(status, '❓')
            
            result_text += f"📅 日期: {item.get('plan_date', 'N/A')}\n"
            result_text += f"📱 机型: {item.get('model_name', 'N/A')}\n"
            result_text += f"🏭 生产线: {item.get('production_line', 'N/A')}\n"
            result_text += f"⏰ 班次: {item.get('shift_type', 'N/A')}\n"
            result_text += f"📊 计划产量: {item.get('planned_quantity', 0):,} 台\n"
            result_text += f"✅ 实际产量: {item.get('actual_quantity', 0):,} 台\n"
            result_text += f"📈 完成率: {completion_rate}%\n"
            result_text += f"{status_icon} 状态: {status}\n"
            result_text += "─" * 30 + "\n"
        
        return result_text
    
    def _format_production_plan_stats_result(self, data: List[Dict]) -> str:
        """格式化生产计划统计结果"""
        if not data:
            return "未找到指定时间范围的生产计划统计数据。"
        
        result_text = "📊 生产计划统计结果：\n\n"
        
        total_planned = 0
        total_actual = 0
        total_completed = 0
        total_pending = 0
        
        for item in data:
            planned = item.get('total_planned', 0)
            actual = item.get('total_actual', 0)
            completed = item.get('completed_plans', 0)
            pending = item.get('pending_plans', 0)
            
            total_planned += planned
            total_actual += actual
            total_completed += completed
            total_pending += pending
            
            result_text += f"📅 日期: {item.get('plan_date', 'N/A')}\n"
            result_text += f"📋 计划数: {item.get('total_plans', 0)} 个\n"
            result_text += f"📊 计划产量: {planned:,} 台\n"
            result_text += f"✅ 实际产量: {actual:,} 台\n"
            result_text += f"📈 平均完成率: {item.get('avg_completion_rate', 0)}%\n"
            result_text += f"✅ 已完成: {completed} 个\n"
            result_text += f"⏳ 待执行: {pending} 个\n"
            result_text += "─" * 30 + "\n"
        
        # 添加汇总信息
        if len(data) > 1:
            overall_completion = (total_actual / total_planned * 100) if total_planned > 0 else 0
            result_text += f"\n📋 汇总统计：\n"
            result_text += f"📊 总计划产量: {total_planned:,} 台\n"
            result_text += f"✅ 总实际产量: {total_actual:,} 台\n"
            result_text += f"📈 整体完成率: {overall_completion:.2f}%\n"
            result_text += f"✅ 总完成计划: {total_completed} 个\n"
            result_text += f"⏳ 总待执行: {total_pending} 个\n"
        
        return result_text
    
    def _format_production_line_efficiency_result(self, data: List[Dict]) -> str:
        """格式化生产线效率分析结果"""
        if not data:
            return "未找到指定条件的生产线效率数据。"
        
        result_text = "⚡ 生产线效率分析：\n\n"
        
        for item in data:
            efficiency = item.get('avg_efficiency', 0)
            efficiency_icon = "🟢" if efficiency >= 95 else "🟡" if efficiency >= 85 else "🔴"
            
            result_text += f"🏭 生产线: {item.get('production_line', 'N/A')}\n"
            result_text += f"⏰ 班次: {item.get('shift_type', 'N/A')}\n"
            result_text += f"📋 任务数: {item.get('total_tasks', 0)} 个\n"
            result_text += f"📊 计划产量: {item.get('total_planned', 0):,} 台\n"
            result_text += f"✅ 实际产量: {item.get('total_actual', 0):,} 台\n"
            result_text += f"{efficiency_icon} 平均效率: {efficiency}%\n"
            result_text += f"✅ 完成任务: {item.get('completed_tasks', 0)} 个\n"
            result_text += "─" * 30 + "\n"
        
        # 找出效率最高和最低的生产线
        if len(data) > 1:
            best_line = max(data, key=lambda x: x.get('avg_efficiency', 0))
            worst_line = min(data, key=lambda x: x.get('avg_efficiency', 0))
            
            result_text += f"\n📊 效率分析：\n"
            result_text += f"🏆 效率最高: {best_line.get('production_line', 'N/A')} ({best_line.get('avg_efficiency', 0)}%)\n"
            result_text += f"⚠️ 效率最低: {worst_line.get('production_line', 'N/A')} ({worst_line.get('avg_efficiency', 0)}%)\n"
        
        return result_text
    
    def _format_generic_result(self, data: List[Dict]) -> str:
        """格式化通用查询结果"""
        result_text = f"📊 查询结果 (共 {len(data)} 条记录)：\n\n"
        
        for i, item in enumerate(data, 1):
            result_text += f"记录 {i}:\n"
            for key, value in item.items():
                result_text += f"  {key}: {value}\n"
            result_text += "─" * 30 + "\n"
        
        return result_text


# 注意：不再创建全局实例，改为在knowledge_service中按需创建
