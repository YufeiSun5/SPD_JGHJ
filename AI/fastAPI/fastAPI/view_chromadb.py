#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
ChromaDB数据查看工具
"""

import chromadb
from chromadb.config import Settings
import json

def view_chromadb():
    """查看ChromaDB中的数据"""
    
    # 连接到ChromaDB
    client = chromadb.PersistentClient(
        path="./chroma_db",
        settings=Settings(
            anonymized_telemetry=False,
            allow_reset=False
        )
    )
    
    print("🔍 ChromaDB数据查看器")
    print("=" * 50)
    
    # 1. 查看所有集合
    collections = client.list_collections()
    print(f"📚 总集合数: {len(collections)}")
    
    for collection in collections:
        print(f"\n📖 集合名称: {collection.name}")
        print(f"   📊 文档数量: {collection.count()}")
        
        # 2. 获取集合中的所有数据
        if collection.count() > 0:
            # 获取前10条数据作为示例
            results = collection.get(
                include=["documents", "metadatas", "embeddings"],
                limit=10
            )
            
            print(f"   📋 前{min(10, len(results['ids']))}条数据:")
            
            for i, (doc_id, metadata, embedding) in enumerate(zip(
                results['ids'], 
                results['metadatas'] if results['metadatas'] else [],
                results['embeddings'] if results['embeddings'] else []
            )):
                print(f"\n   📄 文档 {i+1}:")
                print(f"      🆔 ID: {doc_id}")
                
                if metadata:
                    print(f"      📝 内容: {metadata.get('content', 'N/A')[:100]}...")
                    print(f"      📂 来源: {metadata.get('source', 'N/A')}")
                
                if embedding:
                    print(f"      🧮 向量维度: {len(embedding)}")
                    print(f"      🔢 向量预览: [{embedding[0]:.3f}, {embedding[1]:.3f}, {embedding[2]:.3f}, ...]")
        
        # 3. 测试相似度搜索
        if collection.count() > 0:
            print(f"\n   🔍 测试搜索功能:")
            try:
                # 使用第一个文档的向量进行测试搜索
                first_result = collection.get(limit=1, include=["embeddings"])
                if first_result['embeddings']:
                    test_embedding = first_result['embeddings'][0]
                    
                    search_results = collection.query(
                        query_embeddings=[test_embedding],
                        n_results=3,
                        include=["metadatas", "distances"]
                    )
                    
                    print(f"      ✅ 搜索成功，找到 {len(search_results['ids'][0])} 个相似文档")
                    for j, (res_id, distance) in enumerate(zip(
                        search_results['ids'][0],
                        search_results['distances'][0]
                    )):
                        similarity = 1 - distance
                        print(f"         {j+1}. ID: {res_id}, 相似度: {similarity:.3f}")
            except Exception as e:
                print(f"      ❌ 搜索测试失败: {e}")

def query_specific_content():
    """查询特定内容"""
    client = chromadb.PersistentClient(path="./chroma_db")
    
    try:
        collection = client.get_collection("knowledge_base")
        
        print("\n🔍 交互式查询")
        print("=" * 30)
        
        while True:
            query = input("\n请输入搜索关键词 (输入'quit'退出): ").strip()
            if query.lower() == 'quit':
                break
                
            if not query:
                continue
            
            # 简单的文本搜索（通过metadata）
            try:
                results = collection.get(
                    include=["metadatas"],
                    where={"content": {"$contains": query}}
                )
                
                if results['ids']:
                    print(f"找到 {len(results['ids'])} 个匹配项:")
                    for i, (doc_id, metadata) in enumerate(zip(results['ids'], results['metadatas'])):
                        print(f"{i+1}. ID: {doc_id}")
                        print(f"   内容: {metadata['content'][:200]}...")
                        print(f"   来源: {metadata.get('source', 'N/A')}")
                else:
                    print("未找到匹配的内容")
                    
            except Exception as e:
                print(f"搜索出错: {e}")
                # 尝试获取所有数据进行模糊匹配
                all_results = collection.get(include=["metadatas"])
                matches = []
                for doc_id, metadata in zip(all_results['ids'], all_results['metadatas']):
                    if query.lower() in metadata.get('content', '').lower():
                        matches.append((doc_id, metadata))
                
                if matches:
                    print(f"模糊匹配找到 {len(matches)} 个结果:")
                    for i, (doc_id, metadata) in enumerate(matches[:5]):  # 只显示前5个
                        print(f"{i+1}. ID: {doc_id}")
                        print(f"   内容: {metadata['content'][:200]}...")
                        print(f"   来源: {metadata.get('source', 'N/A')}")
                else:
                    print("未找到匹配的内容")
    
    except Exception as e:
        print(f"连接ChromaDB失败: {e}")

if __name__ == "__main__":
    print("ChromaDB 数据查看工具")
    print("1. 查看所有数据")
    print("2. 交互式搜索")
    
    choice = input("请选择功能 (1/2): ").strip()
    
    if choice == "1":
        view_chromadb()
    elif choice == "2":
        query_specific_content()
    else:
        print("选择无效，显示所有数据:")
        view_chromadb()




























