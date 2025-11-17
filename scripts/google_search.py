import os
import requests
import toml
from typing import Dict, Any, Optional, List
import urllib.parse
from .custom_logger import get_logger

# 为当前文件创建专用的logger
log = get_logger()

class GoogleSearchAPI:
    def __init__(self):
        self.api_key = None
        self.search_engine_id = None
        self.base_url = "https://www.googleapis.com/customsearch/v1"
        self._load_config()
    
    def _load_config(self) -> None:
        """从环境变量或配置文件加载配置"""
        # 首先尝试从环境变量获取
        self.api_key = os.getenv('GOOGLE_SEARCH_API_KEY')
        self.search_engine_id = os.getenv('GOOGLE_SEARCH_ENGINE_ID')
        
        # 如果环境变量不存在，尝试从TOML配置文件读取
        if not self.api_key or not self.search_engine_id:
            try:
                # 从项目根目录的 configs 文件夹查找 config.toml
                current_dir = os.path.dirname(__file__)
                project_root = os.path.dirname(current_dir)  # scripts 的父目录（项目根目录）
                config_path = os.path.join(project_root, 'configs', 'config.toml')
                
                log.debug(f"Looking for config at: {config_path}")
                
                if os.path.exists(config_path):
                    with open(config_path, 'r', encoding='utf-8') as f:
                        config = toml.load(f)
                    
                    google_config = config.get('google_search', {})
                    self.api_key = self.api_key or google_config.get('api_key')
                    self.search_engine_id = self.search_engine_id or google_config.get('search_engine_id')
                    log.debug(f"Loaded config: API Key exists: {bool(self.api_key)}, Search Engine ID exists: {bool(self.search_engine_id)}")
                else:
                    log.warning(f"Config file not found at: {config_path}")
            except Exception as e:
                log.error(f"Failed to load config from TOML file: {e}")
        
        # 验证配置是否完整
        if not self.api_key:
            raise ValueError("Google Search API key not found. Please set GOOGLE_SEARCH_API_KEY environment variable or add it to config.toml")
        
        if not self.search_engine_id:
            raise ValueError("Google Search Engine ID not found. Please set GOOGLE_SEARCH_ENGINE_ID environment variable or add it to config.toml")
    
    def search_linkedin(self, query: str, search_type: str = 'company') -> Dict[str, Any]:
        """
        搜索LinkedIn上的信息
        
        Args:
            query: 搜索查询
            search_type: 搜索类型，'company' 或 'person'
            
        Returns:
            Google Custom Search API的响应结果
        """
        # 根据类型构建查询字符串
        if search_type == 'company':
            search_query = f"{query} site:linkedin.com/company"
        elif search_type == 'person':
            search_query = f"{query} site:linkedin.com/in"
        else:
            search_query = f"{query} site:linkedin.com"
        
        # 构建请求参数
        params = {
            'key': self.api_key,
            'cx': self.search_engine_id,
            'q': search_query
        }
        
        try:
            log.info(f"Searching for {search_type}: {query}")
            response = requests.get(self.base_url, params=params)
            response.raise_for_status()
            log.success(f"Successfully searched for {search_type}: {query}")
            return response.json()
        except requests.exceptions.RequestException as e:
            log.error(f"Error making request to Google Search API for {query}: {e}")
            raise
    
    def search_company_linkedin(self, company_name: str) -> Dict[str, Any]:
        """
        搜索公司在LinkedIn上的信息
        
        Args:
            company_name: 公司名称，如 "nokia"
            
        Returns:
            Google Custom Search API的响应结果
        """
        return self.search_linkedin(company_name, 'company')
    
    def search_person_linkedin(self, person_name: str) -> Dict[str, Any]:
        """
        搜索个人在LinkedIn上的信息
        
        Args:
            person_name: 个人名称，如 "John Doe"
            
        Returns:
            Google Custom Search API的响应结果
        """
        return self.search_linkedin(person_name, 'person')
    
    def search_linkedin_get_top3(self, query: str, search_type: str = 'company') -> List[Dict[str, Any]]:
        """
        搜索LinkedIn上的信息，并返回前3个结果
        
        Args:
            query: 搜索查询
            search_type: 搜索类型，'company' 或 'person'
            
        Returns:
            前3个搜索结果的列表，如果没有结果则返回空列表
        """
        try:
            result = self.search_linkedin(query, search_type)
            
            # 检查是否有items字段且不为空
            if 'items' in result and result['items']:
                top3_items = result['items'][:3]  # 取前3个结果
                log.info(f"Found {len(top3_items)} LinkedIn {search_type} results for {query}")
                return top3_items
            else:
                log.warning(f"No LinkedIn {search_type} search results found for '{query}'")
                return []
                
        except Exception as e:
            log.error(f"Error searching for {search_type} '{query}': {e}")
            return []

    def search_company_linkedin_get_top3(self, company_name: str) -> List[Dict[str, Any]]:
        """
        搜索公司在LinkedIn上的信息，并返回前3个结果
        
        Args:
            company_name: 公司名称，如 "nokia"
            
        Returns:
            前3个搜索结果的列表，如果没有结果则返回空列表
        """
        return self.search_linkedin_get_top3(company_name, 'company')

    def search_person_linkedin_get_top3(self, person_name: str) -> List[Dict[str, Any]]:
        """
        搜索个人在LinkedIn上的信息，并返回前3个结果
        
        Args:
            person_name: 个人名称，如 "John Doe"
            
        Returns:
            前3个搜索结果的列表，如果没有结果则返回空列表
        """
        return self.search_linkedin_get_top3(person_name, 'person')
    
    def search_company_linkedin_get_link(self, company_name: str) -> Optional[str]:
        """
        搜索公司在LinkedIn上的信息，并返回第一个结果的链接
        
        Args:
            company_name: 公司名称，如 "nokia"
            
        Returns:
            第一个搜索结果的链接，如果没有结果则返回None
        """
        try:
            result = self.search_company_linkedin(company_name)
            
            # 检查是否有items字段且不为空
            if 'items' in result and result['items']:
                first_item = result['items'][0]
                link = first_item.get('link')
                log.info(f"Found LinkedIn link for {company_name}: {link}")
                return link
            else:
                log.warning(f"No LinkedIn search results found for company '{company_name}'")
                return None
                
        except Exception as e:
            log.error(f"Error searching for company '{company_name}': {e}")
            return None
    
    def get_search_url(self, query: str, search_type: str = 'company') -> str:
        """
        获取搜索URL（用于调试或直接访问）
        
        Args:
            query: 搜索查询
            search_type: 搜索类型，'company' 或 'person'
            
        Returns:
            完整的搜索URL
        """
        if search_type == 'company':
            search_query = f"{query} site:linkedin.com/company"
        elif search_type == 'person':
            search_query = f"{query} site:linkedin.com/in"
        else:
            search_query = f"{query} site:linkedin.com"
            
        encoded_query = urllib.parse.quote_plus(search_query)
        
        return f"{self.base_url}?key={self.api_key}&cx={self.search_engine_id}&q={encoded_query}"

# 创建全局实例
search_api = GoogleSearchAPI()

def search_company_on_linkedin(company_name: str) -> Dict[str, Any]:
    """
    搜索公司在LinkedIn上的信息
    
    Args:
        company_name: 公司名称，如 "nokia", "microsoft"等
        
    Returns:
        Google Custom Search API的响应结果
    """
    return search_api.search_company_linkedin(company_name)

def search_person_on_linkedin(person_name: str) -> Dict[str, Any]:
    """
    搜索个人在LinkedIn上的信息
    
    Args:
        person_name: 个人名称，如 "John Doe"
        
    Returns:
        Google Custom Search API的响应结果
    """
    return search_api.search_person_linkedin(person_name)

def search_company_on_linkedin_get_link(company_name: str) -> Optional[str]:
    """
    搜索公司在LinkedIn上的信息，并返回第一个结果的链接
    
    Args:
        company_name: 公司名称，如 "nokia", "microsoft"等
        
    Returns:
        第一个搜索结果的链接，如果没有结果则返回None
    """
    return search_api.search_company_linkedin_get_link(company_name)

def search_company_on_linkedin_get_top3(company_name: str) -> List[Dict[str, Any]]:
    """
    搜索公司在LinkedIn上的信息，并返回前3个结果
    
    Args:
        company_name: 公司名称，如 "nokia", "microsoft"等
        
    Returns:
        前3个搜索结果的列表
    """
    return search_api.search_company_linkedin_get_top3(company_name)

def search_person_on_linkedin_get_top3(person_name: str) -> List[Dict[str, Any]]:
    """
    搜索个人在LinkedIn上的信息，并返回前3个结果
    
    Args:
        person_name: 个人名称，如 "John Doe"
        
    Returns:
        前3个搜索结果的列表
    """
    return search_api.search_person_linkedin_get_top3(person_name)

def get_search_url(query: str, search_type: str = 'company') -> str:
    """
    获取搜索URL
    
    Args:
        query: 搜索查询
        search_type: 搜索类型，'company' 或 'person'
        
    Returns:
        完整的搜索URL字符串
    """
    return search_api.get_search_url(query, search_type)