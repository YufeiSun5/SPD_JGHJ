#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
编码修复模块
解决Windows环境下的Unicode编码问题
"""

import sys
import os
import io

def setup_encoding():
    """设置正确的编码配置，支持emoji和Unicode字符"""

    # 设置标准输出编码为UTF-8
    if sys.platform.startswith('win'):
        # Windows平台特殊处理
        try:
            # 设置控制台输出编码为UTF-8
            if hasattr(sys.stdout, 'buffer'):
                sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='backslashreplace')
            if hasattr(sys.stderr, 'buffer'):
                sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8', errors='backslashreplace')
            if hasattr(sys.stdin, 'buffer'):
                sys.stdin = io.TextIOWrapper(sys.stdin.buffer, encoding='utf-8', errors='backslashreplace')
        except Exception as e:
            print(f"警告: 设置控制台编码失败: {e}")

    # 设置环境变量
    os.environ['PYTHONIOENCODING'] = 'utf-8'
    os.environ['PYTHONUTF8'] = '1'
    
    # 设置控制台代码页为UTF-8 (Windows)
    if sys.platform.startswith('win'):
        try:
            import subprocess
            # 尝试设置控制台代码页为UTF-8
            subprocess.run(['chcp', '65001'], shell=True, capture_output=True)
        except Exception:
            pass
        
        # 确保文件系统使用UTF-8编码
        os.environ['PYTHONIOENCODING'] = 'utf-8:surrogateescape'

def safe_print(*args, **kwargs):
    """安全的打印函数，支持emoji字符，并强制刷新输出"""
    try:
        # 使用errors='replace'避免编码错误
        print(*args, **kwargs, file=sys.stdout)
        sys.stdout.flush()  # 强制刷新缓冲区
    except UnicodeEncodeError:
        # 如果仍有编码错误，使用备用打印方法
        try:
            # 替换无法编码的字符
            safe_args = []
            for arg in args:
                if isinstance(arg, str):
                    safe_arg = arg.encode('utf-8', errors='replace').decode('utf-8', errors='replace')
                    safe_args.append(safe_arg)
                else:
                    safe_args.append(arg)
            print(*safe_args, **kwargs, file=sys.stdout)
            sys.stdout.flush()  # 强制刷新缓冲区
        except Exception:
            # 最后的备用方案：忽略所有编码错误
            try:
                print(*args, **kwargs, file=sys.stdout, errors='ignore')
                sys.stdout.flush()  # 强制刷新缓冲区
            except Exception as e:
                print(f"打印失败: {e}", file=sys.stderr)
                sys.stderr.flush()

def safe_encode_for_model(text: str) -> str:
    """
    安全编码文本用于模型处理，避免GBK编码错误
    特别处理emoji和特殊Unicode字符
    """
    try:
        # 首先尝试正常的UTF-8编码
        return text.encode('utf-8', errors='replace').decode('utf-8', errors='replace')
    except Exception:
        # 如果仍有问题，使用更保守的方法
        import re
        # 保留基本ASCII、中文字符，替换其他特殊字符
        safe_text = re.sub(r'[^\u0000-\u007F\u4e00-\u9fff\u3000-\u303f\uff00-\uffef]', '[特殊字符]', text)
        return safe_text

def check_text_encoding_safety(text: str) -> bool:
    """检查文本是否包含可能导致编码问题的字符"""
    try:
        text.encode('gbk')
        return True
    except UnicodeEncodeError:
        return False

# 全局设置
setup_encoding()

# 测试打印
if __name__ == "__main__":
    safe_print("🧠 智能知识库 API 启动中...")
    safe_print("✅ 测试emoji字符")
    safe_print("🔍 搜索功能正常")
    safe_print("🎉 成功解决编码问题！")