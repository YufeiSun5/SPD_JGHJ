"""
OpenAI格式的函数定义
用于标准Function Calling
"""
from typing import List, Dict, Any
from datetime import datetime

def get_function_definitions() -> List[Dict[str, Any]]:
    """获取所有可用函数的OpenAI格式定义"""
    
    return [
        {
            "type": "function",
            "function": {
                "name": "query_defect_rate",
                "description": "查询指定日期和机型的不良率数据。可以查询具体某天的生产质量情况，包括总产量、不良品数量和不良率。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "date": {
                            "type": "string",
                            "description": "查询日期，格式为YYYY-MM-DD。支持相对日期如'今天'、'昨天'、'上周三'等，需要转换为具体日期。",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "model_name": {
                            "type": "string",
                            "description": "机型名称，如iPhone15、SamsungS23、HuaweiP60等",
                            "enum": ["iPhone15", "SamsungS23", "HuaweiP60", "XiaomiMi13"]
                        }
                    },
                    "required": ["date", "model_name"]
                }
            }
        },
        {
            "type": "function", 
            "function": {
                "name": "query_production_stats",
                "description": "查询指定日期范围内的生产统计数据。可以查询某个时间段内的总体生产情况，包括各机型的产量统计。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "start_date": {
                            "type": "string",
                            "description": "开始日期，格式为YYYY-MM-DD",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "end_date": {
                            "type": "string", 
                            "description": "结束日期，格式为YYYY-MM-DD",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "model_name": {
                            "type": "string",
                            "description": "机型名称，可选。如果不指定则查询所有机型",
                            "enum": ["iPhone15", "SamsungS23", "HuaweiP60", "XiaomiMi13"]
                        }
                    },
                    "required": ["start_date", "end_date"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "query_quality_trend", 
                "description": "查询指定日期范围内的质量趋势数据。分析某个时间段内不良率的变化趋势，帮助识别质量问题。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "start_date": {
                            "type": "string",
                            "description": "开始日期，格式为YYYY-MM-DD",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "end_date": {
                            "type": "string",
                            "description": "结束日期，格式为YYYY-MM-DD", 
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "model_name": {
                            "type": "string",
                            "description": "机型名称，可选。如果不指定则查询所有机型",
                            "enum": ["iPhone15", "SamsungS23", "HuaweiP60", "XiaomiMi13"]
                        }
                    },
                    "required": ["start_date", "end_date"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "query_production_plan",
                "description": "查询指定日期的生产计划信息。可以查看某天的生产安排，包括计划产量、实际产量、生产线分配等。",
                "parameters": {
                    "type": "object", 
                    "properties": {
                        "date": {
                            "type": "string",
                            "description": "查询日期，格式为YYYY-MM-DD",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "model_name": {
                            "type": "string",
                            "description": "机型名称，可选。如果不指定则查询所有机型",
                            "enum": ["iPhone15", "SamsungS23", "HuaweiP60", "XiaomiMi13"]
                        },
                        "production_line": {
                            "type": "string",
                            "description": "生产线名称，可选。如果不指定则查询所有生产线",
                            "enum": ["Line-A", "Line-B", "Line-C"]
                        }
                    },
                    "required": ["date"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "query_production_plan_stats",
                "description": "查询指定日期范围内的生产计划统计数据。分析计划完成情况，包括计划达成率、各生产线效率等。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "start_date": {
                            "type": "string",
                            "description": "开始日期，格式为YYYY-MM-DD",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "end_date": {
                            "type": "string",
                            "description": "结束日期，格式为YYYY-MM-DD",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "production_line": {
                            "type": "string",
                            "description": "生产线名称，可选。如果不指定则查询所有生产线",
                            "enum": ["Line-A", "Line-B", "Line-C"]
                        }
                    },
                    "required": ["start_date", "end_date"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "query_production_line_efficiency",
                "description": "查询指定生产线在指定日期范围内的效率数据。分析生产线的运行效率，包括产能利用率、完成率等指标。",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "production_line": {
                            "type": "string", 
                            "description": "生产线名称",
                            "enum": ["Line-A", "Line-B", "Line-C"]
                        },
                        "start_date": {
                            "type": "string",
                            "description": "开始日期，格式为YYYY-MM-DD",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        },
                        "end_date": {
                            "type": "string",
                            "description": "结束日期，格式为YYYY-MM-DD",
                            "pattern": "^\\d{4}-\\d{2}-\\d{2}$"
                        }
                    },
                    "required": ["production_line", "start_date", "end_date"]
                }
            }
        }
    ]

def get_current_date_info() -> str:
    """获取当前日期信息，用于AI理解相对时间"""
    now = datetime.now()
    weekday_names = ['一', '二', '三', '四', '五', '六', '日']
    weekday = weekday_names[now.weekday()]
    
    return f"""当前日期信息：
- 今天是: {now.strftime('%Y-%m-%d')} (星期{weekday})
- 昨天是: {(now - timedelta(days=1)).strftime('%Y-%m-%d')}
- 明天是: {(now + timedelta(days=1)).strftime('%Y-%m-%d')}

时间计算规则：
- "上周三" = 上一周的星期三
- "本周一" = 本周的星期一  
- "前天" = 2天前
- "大前天" = 3天前
"""

# 导入timedelta用于日期计算
from datetime import timedelta
