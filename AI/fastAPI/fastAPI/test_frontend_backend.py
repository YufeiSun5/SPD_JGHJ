#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
前后端集成测试脚本
用于验证前端和后端是否正常配合工作
"""

import requests
import json
import sys
from typing import Dict, Any

# 配置
API_BASE_URL = "http://127.0.0.1:8000"
TIMEOUT = 10

# 颜色输出
class Colors:
    GREEN = '\033[92m'
    RED = '\033[91m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    RESET = '\033[0m'

def print_success(msg: str):
    print(f"{Colors.GREEN}✅ {msg}{Colors.RESET}")

def print_error(msg: str):
    print(f"{Colors.RED}❌ {msg}{Colors.RESET}")

def print_warning(msg: str):
    print(f"{Colors.YELLOW}⚠️  {msg}{Colors.RESET}")

def print_info(msg: str):
    print(f"{Colors.BLUE}ℹ️  {msg}{Colors.RESET}")

def test_health_check() -> bool:
    """测试健康检查接口"""
    print("\n" + "="*60)
    print("测试 1: 健康检查接口")
    print("="*60)
    
    try:
        url = f"{API_BASE_URL}/api/knowledge/health"
        print_info(f"请求: GET {url}")
        
        response = requests.get(url, timeout=TIMEOUT)
        print_info(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            print_info(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")
            
            if data.get("status") == "healthy":
                print_success("健康检查通过")
                return True
            else:
                print_warning(f"服务状态: {data.get('status')}")
                return False
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            return False
            
    except requests.exceptions.ConnectionError:
        print_error(f"无法连接到服务器 {API_BASE_URL}")
        print_warning("请确保后端服务已启动（端口 8000）")
        return False
    except Exception as e:
        print_error(f"测试失败: {e}")
        return False

def test_knowledge_list() -> bool:
    """测试知识库列表接口"""
    print("\n" + "="*60)
    print("测试 2: 知识库列表接口")
    print("="*60)
    
    try:
        url = f"{API_BASE_URL}/api/knowledge/list?limit=5&offset=0"
        print_info(f"请求: GET {url}")
        
        response = requests.get(url, timeout=TIMEOUT)
        print_info(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            
            if data.get("success"):
                total = data.get("total", 0)
                items_count = len(data.get("data", []))
                
                print_info(f"返回的 total: {total}")
                print_info(f"实际返回条目数: {items_count}")
                
                if items_count > 0:
                    print_info(f"第一条知识预览:")
                    first_item = data["data"][0]
                    print(f"  - ID: {first_item.get('id')}")
                    print(f"  - 来源: {first_item.get('source', '未分类')}")
                    print(f"  - 内容: {first_item.get('content', '')[:50]}...")
                
                print_success("知识库列表获取成功")
                
                # 检查 total 是否合理
                if total == items_count and items_count == 5:
                    print_warning("total 可能不准确（等于当前页数量）")
                    print_warning("建议查看《后端优化建议.md》进行优化")
                
                return True
            else:
                print_error(f"请求失败: {data.get('message')}")
                return False
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            return False
            
    except Exception as e:
        print_error(f"测试失败: {e}")
        return False

def test_performance_stats() -> bool:
    """测试性能统计接口"""
    print("\n" + "="*60)
    print("测试 3: 性能统计接口")
    print("="*60)
    
    try:
        url = f"{API_BASE_URL}/api/knowledge/performance"
        print_info(f"请求: GET {url}")
        
        response = requests.get(url, timeout=TIMEOUT)
        print_info(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            
            if data.get("success"):
                stats = data.get("data", {})
                print_info(f"性能统计:")
                print(f"  - 总知识数: {stats.get('total_knowledge_count')}")
                print(f"  - 性能等级: {stats.get('performance_level')}")
                print(f"  - 预估响应时间: {stats.get('estimated_response_time')}")
                
                # 检查性能等级格式
                perf_level = stats.get('performance_level')
                if perf_level in ['excellent', 'good', 'fair', 'poor']:
                    print_success("性能等级格式正确（英文）")
                elif perf_level in ['优秀', '良好', '一般', '需要优化']:
                    print_warning("性能等级是中文格式，前端可能无法正确识别")
                    print_warning("建议查看《后端优化建议.md》进行修复")
                else:
                    print_warning(f"未知的性能等级格式: {perf_level}")
                
                return True
            else:
                print_error(f"请求失败: {data.get('message')}")
                return False
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            return False
            
    except Exception as e:
        print_error(f"测试失败: {e}")
        return False

def test_add_knowledge() -> str:
    """测试添加知识接口"""
    print("\n" + "="*60)
    print("测试 4: 添加知识接口")
    print("="*60)
    
    try:
        url = f"{API_BASE_URL}/api/knowledge/add"
        payload = {
            "content": "这是一条测试知识，用于验证前后端集成。测试时间：2025-11-05",
            "source": "集成测试"
        }
        
        print_info(f"请求: POST {url}")
        print_info(f"数据: {json.dumps(payload, ensure_ascii=False)}")
        
        response = requests.post(
            url, 
            json=payload,
            headers={"Content-Type": "application/json"},
            timeout=TIMEOUT
        )
        print_info(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            
            if data.get("success"):
                knowledge_id = data.get("knowledge_id")
                print_info(f"知识ID: {knowledge_id}")
                print_success("知识添加成功")
                return knowledge_id
            else:
                print_error(f"添加失败: {data.get('message')}")
                return None
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            return None
            
    except Exception as e:
        print_error(f"测试失败: {e}")
        return None

def test_delete_knowledge(knowledge_id: str) -> bool:
    """测试删除知识接口"""
    print("\n" + "="*60)
    print("测试 5: 删除知识接口")
    print("="*60)
    
    if not knowledge_id:
        print_warning("跳过删除测试（没有可删除的知识ID）")
        return False
    
    try:
        url = f"{API_BASE_URL}/api/knowledge/delete/{knowledge_id}"
        print_info(f"请求: DELETE {url}")
        
        response = requests.delete(url, timeout=TIMEOUT)
        print_info(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            data = response.json()
            
            if data.get("success"):
                print_success("知识删除成功")
                return True
            else:
                print_error(f"删除失败: {data.get('message')}")
                return False
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            return False
            
    except Exception as e:
        print_error(f"测试失败: {e}")
        return False

def test_query_stream() -> bool:
    """测试流式查询接口"""
    print("\n" + "="*60)
    print("测试 6: 流式查询接口（SSE）")
    print("="*60)
    
    try:
        url = f"{API_BASE_URL}/api/knowledge/query-stream"
        payload = {"question": "测试问题"}
        
        print_info(f"请求: POST {url}")
        print_info(f"数据: {json.dumps(payload, ensure_ascii=False)}")
        print_info("开始接收流式数据...")
        
        response = requests.post(
            url,
            json=payload,
            headers={
                "Content-Type": "application/json",
                "Accept": "text/event-stream"
            },
            stream=True,
            timeout=30
        )
        
        print_info(f"状态码: {response.status_code}")
        
        if response.status_code == 200:
            event_count = 0
            event_types = set()
            
            for line in response.iter_lines():
                if line:
                    line_str = line.decode('utf-8')
                    if line_str.startswith('data: '):
                        try:
                            data = json.loads(line_str[6:])
                            event_type = data.get('type')
                            event_types.add(event_type)
                            event_count += 1
                            
                            if event_type == 'thinking':
                                print(f"  💭 思考: {data.get('data', '')[:50]}...")
                            elif event_type == 'token':
                                print(f"  📝 生成: {data.get('data', '')}", end='', flush=True)
                            elif event_type == 'docs':
                                docs = data.get('data', [])
                                print(f"\n  📄 找到 {len(docs)} 个相关文档")
                            elif event_type == 'end':
                                print(f"\n  ✅ 完成")
                                break
                            elif event_type == 'error':
                                print(f"\n  ❌ 错误: {data.get('data')}")
                                break
                        except json.JSONDecodeError:
                            pass
            
            print()
            print_info(f"接收到 {event_count} 个事件")
            print_info(f"事件类型: {', '.join(event_types)}")
            
            if 'thinking' in event_types:
                print_success("支持思考过程显示")
            else:
                print_warning("未检测到思考过程事件")
            
            if 'token' in event_types:
                print_success("流式输出正常")
            else:
                print_warning("未检测到流式输出")
            
            return True
        else:
            print_error(f"请求失败: HTTP {response.status_code}")
            return False
            
    except Exception as e:
        print_error(f"测试失败: {e}")
        return False

def print_summary(results: Dict[str, bool]):
    """打印测试摘要"""
    print("\n" + "="*60)
    print("测试摘要")
    print("="*60)
    
    total = len(results)
    passed = sum(1 for v in results.values() if v)
    failed = total - passed
    
    for test_name, result in results.items():
        status = "✅ 通过" if result else "❌ 失败"
        print(f"{status} - {test_name}")
    
    print(f"\n总计: {total} 个测试")
    print(f"通过: {passed} 个")
    print(f"失败: {failed} 个")
    
    if failed == 0:
        print_success("\n所有测试通过！前后端集成正常 🎉")
        return True
    else:
        print_error(f"\n有 {failed} 个测试失败，请检查配置")
        return False

def main():
    """主函数"""
    print(f"""
{'='*60}
前后端集成测试
{'='*60}
后端地址: {API_BASE_URL}
测试时间: 2025-11-05
{'='*60}
    """)
    
    results = {}
    
    # 测试 1: 健康检查
    results["健康检查"] = test_health_check()
    if not results["健康检查"]:
        print_error("\n后端服务未启动或无法访问，终止测试")
        print_info("请先启动后端服务: python main_rag_only.py")
        sys.exit(1)
    
    # 测试 2: 知识库列表
    results["知识库列表"] = test_knowledge_list()
    
    # 测试 3: 性能统计
    results["性能统计"] = test_performance_stats()
    
    # 测试 4 & 5: 添加和删除知识
    knowledge_id = test_add_knowledge()
    results["添加知识"] = knowledge_id is not None
    
    if knowledge_id:
        results["删除知识"] = test_delete_knowledge(knowledge_id)
    else:
        results["删除知识"] = False
    
    # 测试 6: 流式查询
    results["流式查询"] = test_query_stream()
    
    # 打印摘要
    all_passed = print_summary(results)
    
    # 额外建议
    print("\n" + "="*60)
    print("建议")
    print("="*60)
    
    if all_passed:
        print_info("✅ 可以使用修复后的前端页面：knowledge_chat_fixed.html")
        print_info("✅ 在浏览器中打开该文件即可开始使用")
    else:
        print_warning("⚠️  请根据失败的测试项进行排查")
        print_warning("⚠️  查看《前端问题修复说明.md》了解详情")
    
    print_info("📖 更多优化建议请查看《后端优化建议.md》")
    
    sys.exit(0 if all_passed else 1)

if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n测试被用户中断")
        sys.exit(1)






