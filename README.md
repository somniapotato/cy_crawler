# CyCrawler

一个基于Go和RocketMQ的分布式网络爬虫任务处理系统。该系统能够消费RocketMQ中的爬虫任务，调用Python爬虫脚本进行处理，并将结果发送回RocketMQ。

## 项目架构

```
cy_crawler/
├── cmd/cy_crawler/main.go          # 应用程序入口
├── internal/
│   ├── config/                     # 配置管理
│   ├── logger/                     # 日志系统
│   ├── mq/                         # RocketMQ客户端
│   ├── processor/                  # 任务处理器
│   └── types/                      # 数据类型定义
├── scripts/crawler.py              # Python爬虫脚本
├── configs/config.toml             # 配置文件
└── logs/                           # 日志目录
```

## 功能特性

- 🚀 **分布式架构**: 基于RocketMQ的消息队列系统
- 📝 **完善日志**: 结构化日志记录，支持文件轮转和心跳日志
- 🔧 **配置化管理**: 使用TOML配置文件，支持热加载
- 🐍 **Python集成**: 无缝调用Python爬虫脚本
- 🔄 **结果透传**: 自动将处理结果发送回消息队列
- ❌ **异常处理**: 完善的错误处理和重试机制
- 💓 **健康监控**: 每10秒心跳日志，实时监控应用状态

## 快速开始

### 前置要求

- Go 1.25+
- Python 3.6+
- RocketMQ 4.9+
- 以下Python包:
  ```bash
  pip install -r scripts/requirements.txt
  ```

### 安装和运行

1. **克隆项目**
   ```bash
   git clone <repository-url>
   cd cy_crawler
   ```

2. **安装依赖**
   ```bash
   make setup
   ```

3. **配置RocketMQ**
   修改 `configs/config.toml` 中的RocketMQ配置：
   ```toml
   [rocketmq]
   name_server = "127.0.0.1:9876"
   consumer_group = "crawler_consumer"
   producer_group = "crawler_producer"
   consumer_topic = "crawler_tasks"
   producer_topic = "crawler_tasks_result"
   ```

4. **配置环境变量**
  ```bash
export GOOGLE_SEARCH_API_KEY="your_google_api_key"
export GOOGLE_SEARCH_ENGINE_ID="your_search_engine_id"
export SCRAPFLY_API_KEY="your_scrapfly_api_key
   ```

5. **构建项目**
   ```bash
   make build
   ```

6. **运行应用**
   ```bash
   make run
   # 或者直接运行
   ./bin/cy_crawler
   ```

### 配置文件

`configs/config.toml` 示例：

```toml
[rocketmq]
name_server = "127.0.0.1:9876"
consumer_group = "crawler_consumer"
producer_group = "crawler_producer"
consumer_topic = "crawler_tasks"
producer_topic = "crawler_tasks_result"

[log]
level = "info"
file_path = "./logs/cy_crawler.log"
max_size = 100
max_backups = 10
max_age = 30

[application]
python_script_path = "./scripts/crawler.py"
heartbeat_interval = 10
```

## 消息格式

### 输入消息 (crawler_tasks)

```json
{
  "type": "company",
  "name": "biogenex",
  "url": "https://example.com",
  "email": "contact@example.com",
  "country": "US"
}
```

**字段说明:**
- `type`: 任务类型（必需）
- `name`: 公司/组织名称（必需）
- `url`: 目标网站URL（必需）
- `email`: 联系邮箱（可选）
- `country`: 国家代码（可选）

### 输出消息 (crawler_tasks_result)

成功响应：
```json
{
    "code": 200,
    "message": "success",
    "data": [
        //爬虫获得的数据
    ],
    "params": { //这部分都是从MQ 接受的json透传过来的
        "requestId": "6352d81f-1217-4c73-aa11-4031a1daf7c0",
        "requestTime": "2025-11-23 22:22:22",
        "tenantId": "122",
        "companyName": "LEXMARK INTERNATIONAL DE ARGENTINA INC SUCURSAL ARGENTINA",
        "companyWebsite": "www.baidu.com",
        "contactPersonName": "张三",
        "emailAddress": "duxu111@126.com",
        "type": 1,
        "location": "意大利",
        "position": "General Manager",
        "importExperience": "有",
        "industryExperience": "互联网"
    }
}
```

错误响应：
to do

## 测试工具

项目提供了多个测试工具来验证系统功能：

### 消费结果消息

```bash
# 实时消费 crawler_tasks_result 中的结果
go run test_consumer.go
```

### 测试Python脚本

```bash
# 单独测试Python爬虫脚本
python crawler.py --type company --name "biogenex"
```

## Makefile 命令

```bash
make build     # 构建项目
make run       # 运行项目
make clean     # 清理构建文件
make test      # 运行测试
make setup     # 安装依赖

make clean & make build & make run
```

## 故障排除

### 常见问题

1. **RocketMQ连接失败**
   - 检查NameServer地址和端口
   - 确认RocketMQ服务正在运行
   - 验证topic是否存在

2. **Python脚本执行失败**
   - 检查Python环境变量
   - 验证Python依赖是否安装
   - 查看应用程序日志获取详细错误信息

3. **JSON解析错误**
   - 确保Python脚本只输出纯JSON格式
   - 检查是否有额外的print语句或调试输出

### 日志查看

应用日志位于 `logs/cy_crawler.log`，包含：
- 应用启动和关闭信息
- 消息消费和处理记录
- Python脚本执行结果
- 每10秒的心跳日志
- 错误和异常信息

## 开发指南

### 扩展Python爬虫

修改 `scripts/crawler.py` 来实现具体的爬虫逻辑：

```python
def main():
    # 解析命令行参数
    args = parse_arguments()
    
    try:
        # 实现爬虫逻辑
        result = crawl_website(args.url, args.type)
        
        # 返回标准JSON格式
        print(json.dumps(result))
        
    except Exception as e:
        # 错误时返回标准错误格式
        error_result = {"status": "error", "error": str(e)}
        print(json.dumps(error_result))
        sys.exit(1)
```

### 添加新的消息类型

1. 在 `internal/types/types.go` 中定义新的消息结构
2. 更新处理器以支持新的消息格式
3. 修改Python脚本处理新的任务类型

## 监控和维护

### 系统监控

- 查看应用日志：`tail -f logs/cy_crawler.log`
- 监控RocketMQ状态：使用RocketMQ Dashboard
- 检查系统资源使用情况

### 性能优化建议

1. **调整消费者数量**: 根据负载调整并发消费者数量
2. **优化Python脚本**: 减少Python脚本执行时间
3. **消息批量处理**: 考虑实现消息批量处理机制
4. **连接池管理**: 优化数据库和HTTP连接池

## 贡献

欢迎提交Issue和Pull Request来改进这个项目。

## 支持

如果您遇到问题，请：
1. 查看本README文档
2. 检查应用程序日志
3. 提交Issue并提供详细的重现步骤