# network_config.py - 网络配置和代理设置
import os
import requests
from urllib3.util.retry import Retry
from requests.adapters import HTTPAdapter

def setup_network_config():
    """设置网络配置，包括代理和重试策略"""
    
    # 检查是否有本地嵌入模型缓存
    import os
    current_dir = os.path.dirname(os.path.abspath(__file__))
    model_cache_path = os.path.join(current_dir, "models", "embedding_cache", "models--BAAI--bge-small-zh-v1.5")
    
    if os.path.exists(model_cache_path):
        print("🔍 发现本地嵌入模型缓存，启用离线模式")
        # 强制离线模式
        os.environ['TRANSFORMERS_OFFLINE'] = '1'
        os.environ['HF_DATASETS_OFFLINE'] = '1'
        os.environ['HF_HUB_OFFLINE'] = '1'
    else:
        print("⚠️ 未发现本地嵌入模型缓存，允许在线下载")
        # 允许在线下载
        os.environ['TRANSFORMERS_OFFLINE'] = '0'
        os.environ.pop('HF_DATASETS_OFFLINE', None)
        os.environ.pop('HF_HUB_OFFLINE', None)
    
    # 设置环境变量以禁用代理（如果不需要代理）
    os.environ['NO_PROXY'] = '*'
    os.environ['HTTP_PROXY'] = ''
    os.environ['HTTPS_PROXY'] = ''
    os.environ['http_proxy'] = ''
    os.environ['https_proxy'] = ''
    
    # 如果需要使用代理，请取消注释并配置以下行：
    # os.environ['HTTP_PROXY'] = 'http://your-proxy:port'
    # os.environ['HTTPS_PROXY'] = 'http://your-proxy:port'
    
    # 设置 Hugging Face 相关环境变量
    os.environ['HF_HUB_DISABLE_TELEMETRY'] = '1'
    
    print("✅ 网络配置已设置")

def create_session_with_retry():
    """创建带重试机制的requests会话"""
    session = requests.Session()
    
    # 配置重试策略
    retry_strategy = Retry(
        total=3,
        backoff_factor=1,
        status_forcelist=[429, 500, 502, 503, 504],
    )
    
    adapter = HTTPAdapter(max_retries=retry_strategy)
    session.mount("http://", adapter)
    session.mount("https://", adapter)
    
    # 设置超时
    session.timeout = 30
    
    return session

def test_huggingface_connectivity():
    """测试 Hugging Face 连接性"""
    try:
        session = create_session_with_retry()
        response = session.get("https://huggingface.co", timeout=10)
        if response.status_code == 200:
            print("✅ Hugging Face 连接正常")
            return True
        else:
            print(f"⚠️ Hugging Face 连接异常，状态码: {response.status_code}")
            return False
    except Exception as e:
        print(f"❌ Hugging Face 连接失败: {e}")
        return False

if __name__ == "__main__":
    setup_network_config()
    test_huggingface_connectivity()
