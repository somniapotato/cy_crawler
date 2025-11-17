#!/usr/bin/env python3
import subprocess
import sys
import os

def main():
    """包装器脚本，确保没有任何日志输出到控制台"""
    # 设置环境变量，确保没有控制台输出
    env = os.environ.copy()
    env['PYTHONUNBUFFERED'] = '1'
    
    # 构建命令
    cmd = [sys.executable, 'scripts/crawler.py'] + sys.argv[1:]
    
    try:
        # 运行命令，捕获输出
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            env=env,
            check=True
        )
        
        # 只输出标准输出（应该是纯净的JSON）
        if result.stdout.strip():
            print(result.stdout.strip())
        else:
            print("{}")
            
    except subprocess.CalledProcessError as e:
        # 如果出错，输出空JSON
        print("{}")

if __name__ == "__main__":
    main()