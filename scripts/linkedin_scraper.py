import json
import jmespath
from typing import Dict, List
from scrapfly import ScrapeConfig, ScrapflyClient, ScrapeApiResponse
import asyncio
import os
import toml
from .custom_logger import get_logger

# 为当前文件创建专用的logger
log = get_logger()

class ScrapflyConfig:
    def __init__(self):
        self.api_key = None
        self._load_config()
    
    def _load_config(self) -> None:
        """从环境变量或配置文件加载配置"""
        # 首先尝试从环境变量获取
        self.api_key = os.getenv('SCRAPFLY_API_KEY')
        
        # 如果环境变量不存在，尝试从TOML配置文件读取
        if not self.api_key:
            try:
                # 从项目根目录的 configs 文件夹查找 config.toml
                current_dir = os.path.dirname(__file__)
                project_root = os.path.dirname(current_dir)  # scripts 的父目录（项目根目录）
                config_path = os.path.join(project_root, 'configs', 'config.toml')
                
                log.debug(f"Looking for config at: {config_path}")
                
                if os.path.exists(config_path):
                    with open(config_path, 'r', encoding='utf-8') as f:
                        config = toml.load(f)
                    
                    scrapfly_config = config.get('scrapfly', {})
                    self.api_key = scrapfly_config.get('api_key')
                    log.debug(f"Loaded Scrapfly config: API Key exists: {bool(self.api_key)}")
                else:
                    log.warning(f"Config file not found at: {config_path}")
            except Exception as e:
                log.error(f"Failed to load config from TOML file: {e}")
        
        # 验证配置是否完整
        if not self.api_key:
            raise ValueError("Scrapfly API key not found. Please set SCRAPFLY_API_KEY environment variable or add it to config.toml")

# 初始化Scrapfly客户端
scrapfly_config = ScrapflyConfig()
SCRAPFLY = ScrapflyClient(key=scrapfly_config.api_key)

BASE_CONFIG = {
    "asp": True,
    "country": "US",
    "headers": {
        "Accept-Language": "en-US,en;q=0.5"
    }
}

def strip_text(text):
    """remove extra spaces while handling None values"""
    return text.strip() if text != None else text

def parse_company_life(response: ScrapeApiResponse) -> Dict:
    """parse company life page"""
    selector = response.selector
    leaders = []
    for element in selector.xpath("//section[@data-test-id='leaders-at']/div/ul/li"):
        name = element.xpath(".//a/div/h3/text()").get()
        title = element.xpath(".//a/div/h4/text()").get()
        link = element.xpath(".//a/@href").get()
        
        if name and title and link:
            leaders.append({
                "name": name.strip(),
                "title": title.strip(),
                "linkedinProfileLink": link
            })
    
    affiliated_pages = []
    for element in selector.xpath("//section[@data-test-id='affiliated-pages']/div/div/ul/li"):
        name = element.xpath(".//a/div/h3/text()").get()
        industry = element.xpath(".//a/div/p[1]/text()").get()
        address = element.xpath(".//a/div/p[2]/text()").get()
        linkedin_url = element.xpath(".//a/@href").get()
        
        if name and linkedin_url:
            affiliated_pages.append({
                "name": name.strip(),
                "industry": strip_text(industry),
                "address": strip_text(address),
                "linkeinUrl": linkedin_url.split("?")[0] if linkedin_url else None
            })
    
    similar_pages = []
    for element in selector.xpath("//section[@data-test-id='similar-pages']/div/div/ul/li"):
        name = element.xpath(".//a/div/h3/text()").get()
        industry = element.xpath(".//a/div/p[1]/text()").get()
        address = element.xpath(".//a/div/p[2]/text()").get()
        linkedin_url = element.xpath(".//a/@href").get()
        
        if name and linkedin_url:
            similar_pages.append({
                "name": name.strip(),
                "industry": strip_text(industry),
                "address": strip_text(address),
                "linkeinUrl": linkedin_url.split("?")[0] if linkedin_url else None
            })
    
    company_life = {}
    company_life["leaders"] = leaders
    company_life["affiliatedPages"] = affiliated_pages
    company_life["similarPages"] = similar_pages
    return company_life

def parse_company_overview(response: ScrapeApiResponse) -> Dict:
    """parse company main overview page"""
    selector = response.selector
    
    try:
        # 尝试获取JSON-LD数据
        script_element = selector.xpath("//script[@type='application/ld+json']/text()")
        if not script_element:
            log.warning("No JSON-LD data found")
            return {}
            
        _script_data = json.loads(script_element.get())
        _company_types = [item for item in _script_data.get('@graph', []) if item.get('@type') == 'Organization']
        
        if not _company_types:
            log.warning("No Organization data found in JSON-LD")
            return {}
            
        microdata = jmespath.search(
            """{
            name: name,
            url: url,
            mainAddress: address,
            description: description,
            numberOfEmployees: numberOfEmployees.value,
            logo: logo
            }""",
            _company_types[0],
        )
    except Exception as e:
        log.error(f"Error parsing JSON-LD data: {e}")
        microdata = {}
    
    company_about = {}
    try:
        for element in selector.xpath("//div[contains(@data-test-id, 'about-us')]"):
            name = element.xpath(".//dt/text()").get()
            value = element.xpath(".//dd/text()").get()
            
            if name:
                name = name.strip()
                if not value:
                    value = ' '.join(element.xpath(".//dd//text()").getall()).strip().split('\n')[0]
                else:
                    value = value.strip()
                company_about[name] = value
    except Exception as e:
        log.error(f"Error parsing about section: {e}")
    
    company_overview = {**microdata, **company_about}
    log.debug(f"Parsed company overview with {len(company_about)} about fields")
    return company_overview

async def scrape_company(urls: List[str]) -> List[Dict]:
    """scrape public linkedin company pages"""
    log.info(f"Starting to scrape {len(urls)} company pages")
    to_scrape = [ScrapeConfig(url, **BASE_CONFIG) for url in urls]
    data = []
    
    async for response in SCRAPFLY.concurrent_scrape(to_scrape):
        try:
            # create the life page URL from the overview page response
            company_id = str(response.context["url"]).split("/")[-1]
            company_life_url = f"https://linkedin.com/company/{company_id}/life"
            
            # request the company life page
            life_page_response = await SCRAPFLY.async_scrape(ScrapeConfig(company_life_url, **BASE_CONFIG))
            overview = parse_company_overview(response)
            life = parse_company_life(life_page_response)
            data.append({"overview": overview, "life": life})
            log.info(f"Successfully scraped company: {overview.get('name', 'Unknown')}")
        except Exception as e:
            log.error("An error occurred while scraping company pages", exc_info=True)
            continue

    log.success(f"scraped {len(data)} companies from Linkedin")
    return data

async def scrape_company_overview(urls: List[str]) -> List[Dict]:
    """scrape public linkedin company pages - overview only"""
    log.info(f"Starting to scrape overview for {len(urls)} company pages")
    to_scrape = [ScrapeConfig(url, **BASE_CONFIG) for url in urls]
    data = []
    
    async for response in SCRAPFLY.concurrent_scrape(to_scrape):
        try:
            overview = parse_company_overview(response)
            data.append({"overview": overview})
            log.info(f"Successfully scraped company overview: {overview.get('name', 'Unknown')}")
        except Exception as e:
            log.error("An error occurred while scraping company pages", exc_info=True)
            continue

    log.success(f"scraped {len(data)} companies from Linkedin")
    return data