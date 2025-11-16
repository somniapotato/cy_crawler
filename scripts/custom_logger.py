import logging
import sys
import os
from datetime import datetime

class CustomLogger:
    def __init__(self, name=None):
        # 如果没有指定name，自动获取调用者的文件名
        if name is None:
            # 获取调用这个logger的文件的文件名
            frame = sys._getframe(2)  # 跳过两层调用栈
            filename = frame.f_code.co_filename
            name = os.path.splitext(os.path.basename(filename))[0]
        
        self.logger = logging.getLogger(name)
        self.logger.setLevel(logging.INFO)
        
        # 避免重复添加handler
        if not self.logger.handlers:
            # 创建formatter - 显示文件名
            formatter = logging.Formatter(
                '%(asctime)s - %(name)s - %(levelname)s - %(message)s',
                datefmt='%Y-%m-%d %H:%M:%S'
            )
            
            # 只保留文件handler，移除控制台handler
            # 这样日志只会写入文件，不会输出到控制台
            
            # 文件handler - 现在日志文件放在项目根目录的 logs 文件夹
            current_dir = os.path.dirname(__file__)
            project_root = os.path.dirname(current_dir)  # 项目根目录
            log_dir = os.path.join(project_root, 'logs')
            os.makedirs(log_dir, exist_ok=True)
            log_file = os.path.join(log_dir, 'scraper.log')
            
            file_handler = logging.FileHandler(log_file, encoding='utf-8')
            file_handler.setLevel(logging.INFO)
            file_handler.setFormatter(formatter)
            
            # 只添加文件handler，不添加控制台handler
            self.logger.addHandler(file_handler)
    
    def debug(self, message):
        self.logger.debug(message)
    
    def info(self, message):
        self.logger.info(message)
    
    def warning(self, message):
        self.logger.warning(message)
    
    def error(self, message, exc_info=False):
        self.logger.error(message, exc_info=exc_info)
    
    def success(self, message):
        self.logger.info(f"✅ {message}")
    
    def critical(self, message):
        self.logger.critical(message)

# 创建函数来获取logger，这样每个文件都能有自己的logger
def get_logger(name=None):
    """获取指定名称的logger，如果未指定则自动使用文件名"""
    return CustomLogger(name)

# 为了向后兼容，保留全局logger（不推荐在新代码中使用）
log = get_logger("main")