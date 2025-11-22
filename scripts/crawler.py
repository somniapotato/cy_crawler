#!/usr/bin/env python3
import asyncio
import json
import sys
import argparse
import os
import logging
from typing import List, Dict, Any
from urllib.parse import urlparse

logging.basicConfig(level=logging.INFO)

# 添加项目根目录到 Python 路径，以便能够找到 scripts 模块
sys.path.append(os.path.dirname(os.path.dirname(__file__)))

from scripts.google_search import search_company_on_linkedin_get_top3, search_person_on_linkedin_get_top3, search_general_google
from scripts.linkedin_scraper import scrape_company_overview, scrape_profile

def parse_arguments():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(description='LinkedIn Company Crawler')
    
    parser.add_argument('--type', required=True, choices=['company', 'person'],
                       help='类型: company 或 person')
    parser.add_argument('--name', required=False, default='',
                       help='名称: 字符串 (可选)')
    parser.add_argument('--url', required=False, default='',
                       help='网址: 字符串 (可选)')
    parser.add_argument('--email', required=False, default='',
                       help='邮箱: 字符串 (可选)')
    parser.add_argument('--country', required=False, default='',
                       help='国家: 字符串 (可选)')
    
    return parser.parse_args()

def extract_urls_from_google_items(google_items: List[Dict], search_type: str) -> List[str]:
    """
    从Google搜索结果中提取URL，并根据类型过滤
    
    Args:
        google_items: Google搜索结果的items列表
        search_type: 搜索类型，'company' 或 'person'
        
    Returns:
        符合类型要求的URL列表
    """
    urls = []
    
    for item in google_items:
        if 'link' in item:
            url = item['link']
            
            # 根据类型过滤URL
            if search_type == 'company':
                # 公司类型：必须是 LinkedIn 公司页面
                if 'linkedin.com/company/' in url:
                    urls.append(url)
                    
            elif search_type == 'person':
                # 个人类型：必须是 LinkedIn 个人资料页面
                if 'linkedin.com/in/' in url:
                    urls.append(url)
    
    return urls

def extract_domain_from_url(url: str) -> str:
    """
    从URL中提取域名
    
    Args:
        url: 完整的URL，如 "https://biogenex.com/contact-us/"
        
    Returns:
        提取的域名，如 "biogenex.com"
    """
    if not url:
        return ""
    
    try:
        parsed = urlparse(url)
        # 移除www前缀
        domain = parsed.netloc
        if domain.startswith('www.'):
            domain = domain[4:]
        return domain
    except Exception:
        # 如果解析失败，尝试手动提取
        if url.startswith('https://'):
            url = url[8:]
        elif url.startswith('http://'):
            url = url[7:]
        
        # 提取域名部分（第一个/之前的部分）
        if '/' in url:
            domain = url.split('/')[0]
        else:
            domain = url
            
        # 移除www前缀
        if domain.startswith('www.'):
            domain = domain[4:]
        return domain

def build_google_search_query(name: str, url: str) -> str:
    """
    构建Google搜索查询参数
    
    Args:
        name: 名称
        url: 网址
        
    Returns:
        组合后的搜索查询字符串
    """
    if name and url:
        # 从URL中提取域名
        domain = extract_domain_from_url(url)
        return f"{name}+site:{domain}"
    elif name:
        return name
    elif url:
        # 从URL中提取域名
        domain = extract_domain_from_url(url)
        return f"site:{domain}"
    else:
        return ""

async def process_linkedin_chain(args, search_type: str) -> Dict[str, Any]:
    """
    处理LinkedIn链路：Google搜索 -> 提取URL -> LinkedIn爬取
    
    Args:
        args: 命令行参数
        search_type: 搜索类型，'company' 或 'person'
        
    Returns:
        LinkedIn爬取结果
    """
    linkedin_data = []
    
    try:
        # 1. 从Google搜索获取前3个结果（使用LinkedIn搜索）
        if search_type == 'company':
            google_items = search_company_on_linkedin_get_top3(args.name)
        else:
            google_items = search_person_on_linkedin_get_top3(args.name)
        
        if not google_items:
            # 如果没有Google结果，直接返回空结果
            return {"linkedin": []}
        
        # 2. 从Google结果中提取URL用于LinkedIn抓取
        linkedin_urls = extract_urls_from_google_items(google_items, search_type)
        
        if not linkedin_urls:
            # 如果没有提取到URL，直接返回空结果
            return {"linkedin": []}
        
        # 3. 抓取LinkedIn数据
        if search_type == 'company':
            linkedin_data = await scrape_company_overview(linkedin_urls)
        else:
            linkedin_data = await scrape_profile(linkedin_urls)
        
        return {"linkedin": linkedin_data}
        
    except Exception as e:
        # 发生错误时返回已有的结果
        return {"linkedin": linkedin_data}

async def process_google_chain(args) -> Dict[str, Any]:
    """
    处理Google链路：使用name和url组合搜索Google信息
    
    Args:
        args: 命令行参数
        
    Returns:
        Google搜索结果
    """
    try:
        # 构建Google搜索查询
        query = build_google_search_query(args.name, args.url)
        
        if not query:
            # 如果没有查询参数，返回空结果
            return {"google": []}
        
        # 执行Google搜索
        google_data = search_general_google(query)
        
        return {"google": google_data}
        
    except Exception as e:
        # 发生错误时返回空结果
        return {"google": []}

async def process_company(args) -> Dict[str, Any]:
    """
    处理公司类型的请求
    
    Returns:
        包含Google和LinkedIn数据的完整结果结构
    """
    result = {
        "sources": {
            "google": [],
            "linkedin": []
        }
    }
    
    try:
        # 并行执行两条链路
        linkedin_task = process_linkedin_chain(args, 'company')
        google_task = process_google_chain(args)
        
        # 等待两条链路完成
        linkedin_result, google_result = await asyncio.gather(
            linkedin_task, google_task, return_exceptions=True
        )
        
        # 处理LinkedIn结果
        if isinstance(linkedin_result, dict) and "linkedin" in linkedin_result:
            result["sources"]["linkedin"] = linkedin_result["linkedin"]
        
        # 处理Google结果
        if isinstance(google_result, dict) and "google" in google_result:
            result["sources"]["google"] = google_result["google"]
        
        return result
        
    except Exception as e:
        # 发生错误时返回已有的结果（可能包含部分数据）
        return result

async def process_person(args) -> Dict[str, Any]:
    """
    处理个人类型的请求
    
    Returns:
        包含Google和LinkedIn数据的完整结果结构
    """
    result = {
        "sources": {
            "google": [],
            "linkedin": []
        }
    }
    
    try:
        # 并行执行两条链路
        linkedin_task = process_linkedin_chain(args, 'person')
        google_task = process_google_chain(args)
        
        # 等待两条链路完成
        linkedin_result, google_result = await asyncio.gather(
            linkedin_task, google_task, return_exceptions=True
        )
        
        # 处理LinkedIn结果
        if isinstance(linkedin_result, dict) and "linkedin" in linkedin_result:
            result["sources"]["linkedin"] = linkedin_result["linkedin"]
        
        # 处理Google结果
        if isinstance(google_result, dict) and "google" in google_result:
            result["sources"]["google"] = google_result["google"]
        
        return result
        
    except Exception as e:
        # 发生错误时返回已有的结果（可能包含部分数据）
        return result

async def main():
    """主函数 - 根据命令行参数处理请求"""
    # 解析命令行参数
    args = parse_arguments()
    
    # 根据类型处理
    if args.type == 'person':
        # 处理个人类型
        result = await process_person(args)
        print(json.dumps(result, ensure_ascii=False))
    
    elif args.type == 'company':
        # 处理公司类型
        result = await process_company(args)
        print(json.dumps(result, ensure_ascii=False))
    
    else:
        print(json.dumps({"sources": {"google": [], "linkedin": []}}, ensure_ascii=False))

if __name__ == "__main__":
    # 运行异步主函数
    asyncio.run(main())
