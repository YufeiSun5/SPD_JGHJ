# test_logging_system.py - 测试日志系统
import asyncio
import time
from logger_config import setup_logging, get_logger, SafePrint

def test_logging_system():
    """测试日志系统的各种功能"""
    
    print("=" * 60)
    print("🧪 开始测试日志系统")
    print("=" * 60)
    
    # 设置日志系统
    setup_logging()
    
    # 获取不同的日志器
    main_logger = get_logger('test_main')
    rag_logger = get_logger('test_rag')
    sql_logger = get_logger('test_sql')
    
    # 创建安全打印实例
    safe_print = SafePrint('test_app')
    
    print("\n1. 测试基本日志级别:")
    main_logger.debug("这是调试信息 - 应该只在文件中看到")
    main_logger.info("这是普通信息 - 控制台和文件都有")
    main_logger.warning("这是警告信息 - 黄色显示")
    main_logger.error("这是错误信息 - 红色显示")
    
    print("\n2. 测试SafePrint类:")
    safe_print.info("普通信息输出")
    safe_print.success("成功操作 ✅")
    safe_print.progress("进度信息 🔄")
    safe_print.warning("警告信息 ⚠️")
    safe_print.failed("失败信息 ❌")
    
    print("\n3. 测试实时输出（每秒一条消息）:")
    for i in range(5):
        safe_print.progress(f"处理进度: {i+1}/5")
        time.sleep(1)
    
    safe_print.success("实时输出测试完成")
    
    print("\n4. 测试不同模块的日志:")
    rag_logger.info("RAG服务 - 嵌入模型加载中...")
    sql_logger.info("SQL服务 - 数据库连接成功")
    rag_logger.warning("RAG服务 - 未找到相关文档")
    sql_logger.error("SQL服务 - 查询语法错误")
    
    print("\n5. 测试emoji和特殊字符:")
    safe_print.info("🚀 系统启动成功")
    safe_print.info("📊 数据统计: 用户数量 1,234 人")
    safe_print.info("🔍 搜索结果: 找到 56 条记录")
    safe_print.info("💾 数据保存: 文件大小 2.5MB")
    
    print("\n6. 测试长消息:")
    long_message = "这是一条很长的日志消息，" * 20
    safe_print.info(f"长消息测试: {long_message}")
    
    print("\n" + "=" * 60)
    print("✅ 日志系统测试完成！")
    print("📁 日志文件位置: D:/fastapi/fastapi.log")
    print("🔍 请检查控制台输出是否实时显示")
    print("📝 请检查日志文件是否正确记录")
    print("=" * 60)

async def test_async_logging():
    """测试异步环境下的日志"""
    logger = get_logger('async_test')
    safe_print = SafePrint('async_test')
    
    print("\n🔄 开始异步日志测试...")
    
    # 模拟异步操作
    tasks = []
    for i in range(3):
        async def async_task(task_id):
            safe_print.progress(f"异步任务 {task_id} 开始")
            await asyncio.sleep(1)
            safe_print.success(f"异步任务 {task_id} 完成")
        
        tasks.append(async_task(i+1))
    
    await asyncio.gather(*tasks)
    safe_print.success("所有异步任务完成")

if __name__ == "__main__":
    # 测试同步日志
    test_logging_system()
    
    # 测试异步日志
    print("\n" + "=" * 40)
    asyncio.run(test_async_logging())
    print("=" * 40)
