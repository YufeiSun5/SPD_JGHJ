import os

# 1. 配置：只保留核心逻辑文件。
# 如果你觉得 .json 噪音还是太大（比如有很多 package.json），可以把 '.json' 也删掉。
EXTS = ('.go', '.vue', '.js', '.ts') 

# 2. 深度黑名单：只要路径里包含这些词，通通跳过
BLACK_LIST_KEYWORDS = {
    'node_modules', '.git', 'dist', 'vendor', 'bin', '.vfox', 
    'pkg', 'mod', 'obj', '.idea', '.vscode', 'assets',
    'embedding_cache', 'models--', 'snapshots', 'AI',  # 新增：针对 AI 模型的过滤
    'venv', '__pycache__', '.ipynb_checkpoints'        # 新增：Python 环境过滤
}

OUTPUT_FILE = 'copyright_code_clean.txt'

def export_code():
    count = 0
    print("--- 开始精准扫描（已排除 AI 缓存及 SQL） ---")
    
    with open(OUTPUT_FILE, 'w', encoding='utf-8') as f_out:
        for root, dirs, files in os.walk('.'):
            # 第一层过滤：检查当前文件夹名是否在黑名单中
            # 这一步能大幅提高扫描速度
            dirs[:] = [d for d in dirs if d not in BLACK_LIST_KEYWORDS]
            
            # 第二层过滤：检查完整路径里是否包含黑名单关键词（防止嵌套目录）
            path_parts = root.split(os.sep)
            if any(key in path_parts for key in BLACK_LIST_KEYWORDS):
                continue

            for file in files:
                if file.endswith(EXTS):
                    if file in [OUTPUT_FILE, 'export_code.py']:
                        continue
                        
                    file_path = os.path.join(root, file)
                    
                    try:
                        # 重点过滤：文件大小超过 100KB 的通常不是手写代码（排除掉压缩后的 tokenizer.json 等）
                        # 正常的 .go 或 .vue 文件很少超过 100KB
                        if os.path.getsize(file_path) > 100 * 1024:
                            print(f"  [跳过过大文件] {file_path}")
                            continue
                            
                        with open(file_path, 'r', encoding='utf-8') as f_in:
                            content = f_in.read()
                        
                        # 检查内容是否包含明显的机器生成特征（可选）
                        if len(content) > 0 and content.count('\n') < 5 and len(content) > 1000:
                            # 如果代码几乎没有换行且特别长，通常是混淆后的 JS 或 JSON，跳过
                            continue

                        f_out.write(f"\n// {'='*60}\n")
                        f_out.write(f"// Path: {file_path}\n")
                        f_out.write(f"// {'='*60}\n\n")
                        f_out.write(content)
                        f_out.write("\n\n")
                        
                        print(f"  [提取成功] {file_path}")
                        count += 1
                    except:
                        pass

    print(f"\n--- 扫描完成！ ---")
    print(f"共提取了 {count} 个文件。")
    print(f"结果文件: {os.path.abspath(OUTPUT_FILE)}")
    input("按回车键退出...")

if __name__ == "__main__":
    export_code()