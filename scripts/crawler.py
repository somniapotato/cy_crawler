#!/usr/bin/env python3
import asyncio
import json
import sys
import argparse
import os

# 添加项目根目录到 Python 路径，以便能够找到 scripts 模块
sys.path.append(os.path.dirname(os.path.dirname(__file__)))

from scripts.google_search import search_company_on_linkedin_get_link
from scripts.linkedin_scraper import scrape_company_overview
from scripts.custom_logger import get_logger

# 为crawler.py创建专用的logger
log = get_logger()

def parse_arguments():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(description='LinkedIn Company Crawler')
    
    parser.add_argument('--type', required=True, choices=['company', 'person'],
                       help='类型: company 或 person')
    parser.add_argument('--name', required=True, 
                       help='名称: 字符串')
    parser.add_argument('--url', required=False, default='',
                       help='网址: 字符串 (可选)')
    parser.add_argument('--email', required=False, default='',
                       help='邮箱: 字符串 (可选)')
    parser.add_argument('--country', required=False, default='',
                       help='国家: 字符串 (可选)')
    
    return parser.parse_args()

async def process_company(company_name: str):
    """处理公司类型的请求"""
    try:
        log.info(f"开始处理公司: {company_name}")
        
        # 搜索公司LinkedIn链接
        log.info(f"正在搜索 {company_name} 的LinkedIn链接...")
        linkedin_url = search_company_on_linkedin_get_link(company_name)
        
        if not linkedin_url:
            log.error(f"未找到公司 {company_name} 的LinkedIn链接")
            return None
        
        log.success(f"找到LinkedIn链接: {linkedin_url}")
        
        # 抓取公司数据
        log.info(f"开始抓取公司数据...")
        companies_data = await scrape_company_overview([linkedin_url])
        
        if not companies_data:
            log.error(f"未能抓取到公司 {company_name} 的数据")
            return None
        
        log.success(f"成功抓取公司数据")
        return companies_data[0]  # 返回第一个公司的数据
        
    except Exception as e:
        log.error(f"处理公司 {company_name} 时发生错误: {e}", exc_info=True)
        return None

async def main():
    """主函数 - 根据命令行参数处理请求"""
    # 解析命令行参数
    args = parse_arguments()
    
    # 构建日志信息，只显示提供的参数
    log_info_parts = [f"类型: {args.type}", f"名称: {args.name}"]
    if args.url:
        log_info_parts.append(f"网址: {args.url}")
    if args.email:
        log_info_parts.append(f"邮箱: {args.email}")
    if args.country:
        log_info_parts.append(f"国家: {args.country}")
    
    log.info(f"收到请求 - {', '.join(log_info_parts)}")
    
    # 根据类型处理
    if args.type == 'person':
        log.info("类型为 person，跳过处理")
        # 输出空的JSON对象表示不处理
        print(json.dumps({}, ensure_ascii=False))
        return
    
    elif args.type == 'company':
        # 处理公司类型
        result = await process_company(args.name)
        
        if result:
            # 输出JSON结果到标准输出
            print(json.dumps(result, ensure_ascii=False))
        else:
            # 输出空的JSON对象表示处理失败
            print(json.dumps({}, ensure_ascii=False))
    
    else:
        log.error(f"未知的类型: {args.type}")
        print(json.dumps({}, ensure_ascii=False))

# python crawler.py --type company --name "biogenex"
if __name__ == "__main__":
    # 运行异步主函数
    asyncio.run(main())