#!/usr/bin/env python3
# test_knowledge_api.py
"""
知识库API测试脚本
用于测试各个API接口的功能
"""

import asyncio
import json
import aiohttp
import config

BASE_URL = f"http://{config.HOST}:{config.PORT}"

async def test_api():
    """测试知识库API的各个功能"""
    print("=" * 60)
    print("🧪 知识库API功能测试")
    print("=" * 60)
    
    async with aiohttp.ClientSession() as session:
        try:
            # 1. 测试健康检查
            print("\n1. 测试健康检查...")
            async with session.get(f"{BASE_URL}/api/knowledge/health") as resp:
                if resp.status == 200:
                    data = await resp.json()
                    print("✅ 健康检查通过")
                    print(f"   状态: {data.get('status')}")
                    print(f"   消息: {data.get('message')}")
                else:
                    print(f"❌ 健康检查失败: {resp.status}")
                    return
            
            # 2. 测试添加知识
            print("\n2. 测试添加知识...")
            test_knowledge = {
                "content": "人工智能（AI）是计算机科学的一个分支，致力于开发能够执行通常需要人类智能的任务的系统。这包括学习、推理、问题解决、感知、语言理解等能力。",
                "source": "AI基础知识"
            }
            
            async with session.post(
                f"{BASE_URL}/api/knowledge/add",
                json=test_knowledge
            ) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    if data.get('success'):
                        knowledge_id = data.get('knowledge_id')
                        print("✅ 知识添加成功")
                        print(f"   知识ID: {knowledge_id}")
                    else:
                        print(f"❌ 知识添加失败: {data.get('message')}")
                        return
                else:
                    print(f"❌ 知识添加请求失败: {resp.status}")
                    return
            
            # 3. 测试获取知识列表
            print("\n3. 测试获取知识列表...")
            async with session.get(f"{BASE_URL}/api/knowledge/list") as resp:
                if resp.status == 200:
                    data = await resp.json()
                    if data.get('success'):
                        items = data.get('data', [])
                        print(f"✅ 获取知识列表成功，共 {len(items)} 条知识")
                        for item in items[:3]:  # 只显示前3条
                            print(f"   - {item['id'][:8]}...: {item['content'][:50]}...")
                    else:
                        print(f"❌ 获取知识列表失败: {data.get('message')}")
                else:
                    print(f"❌ 获取知识列表请求失败: {resp.status}")
            
            # 4. 测试智能问答
            print("\n4. 测试智能问答...")
            query = {
                "question": "什么是人工智能？"
            }
            
            async with session.post(
                f"{BASE_URL}/api/knowledge/query",
                json=query
            ) as resp:
                if resp.status == 200:
                    data = await resp.json()
                    print("✅ 智能问答测试完成")
                    print(f"   问题: {query['question']}")
                    print(f"   回答: {data.get('answer', '无回答')}")
                    print(f"   使用上下文: {data.get('context_used', False)}")
                    print(f"   相关文档数: {len(data.get('relevant_docs', []))}")
                    
                    if data.get('error'):
                        print(f"   错误信息: {data['error']}")
                else:
                    print(f"❌ 智能问答请求失败: {resp.status}")
            
            # 5. 测试删除知识（如果有知识ID的话）
            if 'knowledge_id' in locals():
                print(f"\n5. 测试删除知识 (ID: {knowledge_id[:8]}...)...")
                async with session.delete(
                    f"{BASE_URL}/api/knowledge/delete/{knowledge_id}"
                ) as resp:
                    if resp.status == 200:
                        data = await resp.json()
                        if data.get('success'):
                            print("✅ 知识删除成功")
                        else:
                            print(f"❌ 知识删除失败: {data.get('message')}")
                    else:
                        print(f"❌ 知识删除请求失败: {resp.status}")
            
            print("\n" + "=" * 60)
            print("🎉 API测试完成!")
            print("=" * 60)
            
        except aiohttp.ClientConnectorError:
            print("❌ 无法连接到API服务器")
            print(f"   请确保服务器正在运行: {BASE_URL}")
            print("   你可以运行: python start_knowledge_api.py")
        except Exception as e:
            print(f"❌ 测试过程中出现错误: {e}")

async def main():
    """主函数"""
    print("准备测试知识库API...")
    print(f"目标服务器: {BASE_URL}")
    print("请确保API服务器正在运行...")
    
    await asyncio.sleep(1)
    await test_api()

if __name__ == "__main__":
    asyncio.run(main())





























