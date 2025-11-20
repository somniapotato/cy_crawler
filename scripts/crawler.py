#!/usr/bin/env python3
import asyncio
import json
import sys
import argparse
import os
import logging
from typing import List, Dict, Any

logging.basicConfig(level=logging.INFO)

# 添加项目根目录到 Python 路径，以便能够找到 scripts 模块
sys.path.append(os.path.dirname(os.path.dirname(__file__)))

from scripts.google_search import search_company_on_linkedin_get_top3, search_person_on_linkedin_get_top3
from scripts.linkedin_scraper import scrape_company_overview, scrape_profile

def parse_arguments():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(description='LinkedIn Company Crawler')
    
    parser.add_argument('--type', required=True, choices=['1', '2'],
                       help='类型: "1" 代表公司, "2" 代表个人')
    parser.add_argument('--name', required=True, 
                       help='名称: 字符串')
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

async def process_company(company_name: str) -> Dict[str, Any]:
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
        # 1. 从Google搜索获取前3个结果（使用公司搜索）
        google_items = search_company_on_linkedin_get_top3(company_name)
        
        if not google_items:
            # 如果没有Google结果，直接返回空结构
            return result
        
        # 将Google结果存入最终结果
        result["sources"]["google"] = google_items
        
        # 2. 从Google结果中提取URL用于LinkedIn抓取
        linkedin_urls = extract_urls_from_google_items(google_items, 'company')
        
        if not linkedin_urls:
            # 如果没有提取到URL，直接返回Google结果
            return result
        
        # 3. 抓取LinkedIn公司数据
        linkedin_data = await scrape_company_overview(linkedin_urls)
        
        if linkedin_data:
            # 将LinkedIn数据存入最终结果
            result["sources"]["linkedin"] = linkedin_data
        
        return result
        
    except Exception as e:
        # 发生错误时返回已有的结果（可能包含部分数据）
        return result

async def process_person(person_name: str) -> Dict[str, Any]:
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
        # 1. 从Google搜索获取前3个结果（使用个人搜索）
        google_items = search_person_on_linkedin_get_top3(person_name)
        
        if not google_items:
            # 如果没有Google结果，直接返回空结构
            return result
        
        # 将Google结果存入最终结果
        result["sources"]["google"] = google_items
        
        # 2. 从Google结果中提取URL用于LinkedIn抓取
        linkedin_urls = extract_urls_from_google_items(google_items, 'person')
        
        if not linkedin_urls:
            # 如果没有提取到URL，直接返回Google结果
            return result
        
        # 3. 抓取LinkedIn个人资料数据
        linkedin_data = await scrape_profile(linkedin_urls)
        
        if linkedin_data:
            # 将LinkedIn数据存入最终结果
            result["sources"]["linkedin"] = linkedin_data
        
        return result
        
    except Exception as e:
        # 发生错误时返回已有的结果（可能包含部分数据）
        return result

async def main():
    """主函数 - 根据命令行参数处理请求"""
    # 解析命令行参数
    args = parse_arguments()
    
    # 根据类型处理
    if args.type == '2':
        # 处理个人类型 (2 = person)
        result = await process_person(args.name)
        print(json.dumps(result, ensure_ascii=False))
    
    elif args.type == '1':
        # 处理公司类型 (1 = company)
        result = await process_company(args.name)
        print(json.dumps(result, ensure_ascii=False))
    
    else:
        print(json.dumps({"sources": {"google": [], "linkedin": []}}, ensure_ascii=False))

if __name__ == "__main__":
    # 运行异步主函数
    asyncio.run(main())
