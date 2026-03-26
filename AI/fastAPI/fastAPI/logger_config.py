# logger_config.py - 日志配置模块
import logging
import sys
import os
from datetime import datetime
from logging.handlers import TimedRotatingFileHandler
from pathlib import Path

class ColoredFormatter(logging.Formatter):
    """带颜色的控制台日志格式化器"""
    
    # 颜色代码
    COLORS = {
        'DEBUG': '\033[36m',    # 青色
        'INFO': '\033[32m',     # 绿色
        'WARNING': '\033[33m',  # 黄色
        'ERROR': '\033[31m',    # 红色
        'CRITICAL': '\033[35m', # 紫色
    }
    RESET = '\033[0m'
    
    def format(self, record):
        # 添加颜色
        if record.levelname in self.COLORS:
            record.levelname = f"{self.COLORS[record.levelname]}{record.levelname}{self.RESET}"
        
        return super().format(record)

class RealTimeHandler(logging.StreamHandler):
    """实时输出处理器，确保立即刷新"""
    
    def emit(self, record):
        try:
            msg = self.format(record)
            stream = self.stream
            stream.write(msg + self.terminator)
            stream.flush()  # 强制刷新缓冲区
        except Exception:
            self.handleError(record)

def setup_logging():
    """设置日志系统"""
    
    # 创建日志目录
    log_dir = Path("D:/fastapi")
    log_dir.mkdir(parents=True, exist_ok=True)
    
    # 获取根日志器
    root_logger = logging.getLogger()
    root_logger.setLevel(logging.DEBUG)
    
    # 清除现有的处理器
    for handler in root_logger.handlers[:]:
        root_logger.removeHandler(handler)
    
    # 1. 控制台处理器（实时输出）
    console_handler = RealTimeHandler(sys.stdout)
    console_handler.setLevel(logging.INFO)
    
    # 控制台格式（带颜色）
    console_formatter = ColoredFormatter(
        '%(asctime)s | %(levelname)s | %(name)s | %(message)s',
        datefmt='%H:%M:%S'
    )
    console_handler.setFormatter(console_formatter)
    
    # 2. 文件处理器（每天轮转）
    log_file = log_dir / "fastapi.log"
    file_handler = TimedRotatingFileHandler(
        filename=str(log_file),
        when='midnight',
        interval=1,
        backupCount=30,  # 保留30天的日志
        encoding='utf-8'
    )
    file_handler.setLevel(logging.DEBUG)
    
    # 文件格式（详细信息）
    file_formatter = logging.Formatter(
        '%(asctime)s | %(levelname)s | %(name)s | %(filename)s:%(lineno)d | %(funcName)s | %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S'
    )
    file_handler.setFormatter(file_formatter)
    
    # 添加处理器到根日志器
    root_logger.addHandler(console_handler)
    root_logger.addHandler(file_handler)
    
    # 设置第三方库的日志级别
    logging.getLogger('uvicorn').setLevel(logging.INFO)
    logging.getLogger('uvicorn.access').setLevel(logging.WARNING)
    logging.getLogger('sqlalchemy').setLevel(logging.WARNING)
    logging.getLogger('chromadb').setLevel(logging.WARNING)
    
    return root_logger

def get_logger(name: str = None):
    """获取日志器"""
    if name is None:
        name = __name__
    return logging.getLogger(name)

# 创建应用专用的日志器
app_logger = get_logger('fastapi_app')
rag_logger = get_logger('rag_service')
sql_logger = get_logger('sql_service')

class SafePrint:
    """安全的打印类，同时输出到控制台和日志"""
    
    def __init__(self, logger_name='app'):
        self.logger = get_logger(logger_name)
    
    def info(self, message):
        """信息级别输出"""
        self.logger.info(message)
        # 强制刷新标准输出
        sys.stdout.flush()
    
    def debug(self, message):
        """调试级别输出"""
        self.logger.debug(message)
        sys.stdout.flush()
    
    def warning(self, message):
        """警告级别输出"""
        self.logger.warning(message)
        sys.stdout.flush()
    
    def error(self, message):
        """错误级别输出"""
        self.logger.error(message)
        sys.stderr.flush()
    
    def success(self, message):
        """成功信息（使用info级别，但添加✅标记）"""
        self.logger.info(f"✅ {message}")
        sys.stdout.flush()
    
    def progress(self, message):
        """进度信息（使用info级别，但添加🔄标记）"""
        self.logger.info(f"🔄 {message}")
        sys.stdout.flush()
    
    def failed(self, message):
        """失败信息（使用error级别，但添加❌标记）"""
        self.logger.error(f"❌ {message}")
        sys.stderr.flush()

# 创建全局安全打印实例
safe_print = SafePrint('app')

def force_flush_all():
    """强制刷新所有输出流"""
    sys.stdout.flush()
    sys.stderr.flush()

if __name__ == "__main__":
    # 测试日志系统
    setup_logging()
    
    logger = get_logger('test')
    logger.info("这是一条测试信息")
    logger.warning("这是一条警告信息")
    logger.error("这是一条错误信息")
    
    # 测试安全打印
    test_print = SafePrint('test')
    test_print.info("测试信息输出")
    test_print.success("测试成功输出")
    test_print.progress("测试进度输出")
    test_print.failed("测试失败输出")
