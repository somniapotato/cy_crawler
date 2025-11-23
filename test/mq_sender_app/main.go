package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
)

// MQ配置
type MQConfig struct {
	Endpoints  string
	AccessKey  string
	SecretKey  string
	InstanceID string
	Topic      string
}

// TaskMessage 消息结构
type TaskMessage struct {
	RequestID          string `json:"requestId"`
	RequestTime        string `json:"requestTime"`
	TenantID           string `json:"tenantId"`
	CompanyName        string `json:"companyName"`
	CompanyWebsite     string `json:"companyWebsite"`
	ContactPersonName  string `json:"contactPersonName"`
	EmailAddress       string `json:"emailAddress"`
	Type               int    `json:"type"`
	Location           string `json:"location"`
	Position           string `json:"position"`
	ImportExperience   string `json:"importExperience"`
	IndustryExperience string `json:"industryExperience"`
}

var (
	mqConfig   *MQConfig
	mqProducer rocketmq.Producer
	ctx        = context.Background()
	serverHost = "0.0.0.0"
	serverPort = "8080"
)

// 获取环境变量，支持默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// 初始化配置
func initConfig() *MQConfig {
	accessKey := os.Getenv("ROCKETMQ_ACCESS_KEY")
	secretKey := os.Getenv("ROCKETMQ_SECRET_KEY")

	if accessKey == "" || secretKey == "" {
		log.Fatal("请设置环境变量 ROCKETMQ_ACCESS_KEY 和 ROCKETMQ_SECRET_KEY")
	}

	// 从环境变量获取服务器配置
	serverHost = getEnv("SERVER_HOST", "0.0.0.0")
	serverPort = getEnv("SERVER_PORT", "8080")

	return &MQConfig{
		Endpoints:  "http://MQ_INST_1625550118601853_BZenvcEe.cn-hangzhou.mq.aliyuncs.com:80",
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		InstanceID: "MQ_INST_1625550118601853_BZenvcEe",
		Topic:      "dev_search_task_request",
	}
}

// 初始化MQ生产者
func initMQProducer() (rocketmq.Producer, error) {
	// 设置RocketMQ日志级别
	rlog.SetLogLevel("error")

	p, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{mqConfig.Endpoints})),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: mqConfig.AccessKey,
			SecretKey: mqConfig.SecretKey,
		}),
		producer.WithInstanceName(mqConfig.InstanceID),
		producer.WithGroupName("GID_DEV_SEARCH_TASK_REQUEST"),
		producer.WithRetry(2),
		producer.WithNamespace(mqConfig.InstanceID), // 添加命名空间配置
	)
	if err != nil {
		return nil, fmt.Errorf("创建生产者失败: %v", err)
	}

	err = p.Start()
	if err != nil {
		return nil, fmt.Errorf("启动生产者失败: %v", err)
	}

	log.Println("RocketMQ生产者启动成功")
	return p, nil
}

// 发送消息到MQ
func sendMessageToMQ(message *TaskMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	msg := primitive.NewMessage(mqConfig.Topic, jsonData)
	msg.WithTag("web_send")
	msg.WithKeys([]string{message.RequestID})

	log.Printf("准备发送消息到Topic: %s, RequestID: %s", mqConfig.Topic, message.RequestID)

	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := mqProducer.SendSync(sendCtx, msg)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}

	log.Printf("消息发送成功: MsgID=%s, Status=%s", result.MsgID, result.Status)
	return nil
}

// 获取服务器访问地址（用于页面显示）
func getServerAddress(r *http.Request) string {
	// 尝试从X-Forwarded-For或Host头获取真实访问地址
	host := r.Host
	if host == "" {
		host = serverHost + ":" + serverPort
	}

	// 如果是本地访问，显示localhost
	if host == "0.0.0.0:8080" || host == ":8080" {
		return "localhost:8080"
	}

	return host
}

// Web处理器
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>MQ消息发送工具</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { box-sizing: border-box; }
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f5f5f5; }
        .container { max-width: 900px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; margin-bottom: 30px; }
        .form-group { margin-bottom: 20px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; color: #555; }
        textarea { width: 100%; height: 300px; padding: 12px; border: 1px solid #ddd; border-radius: 5px; font-family: monospace; font-size: 14px; resize: vertical; }
        button { background: #007cba; color: white; padding: 12px 30px; border: none; border-radius: 5px; cursor: pointer; font-size: 16px; width: 100%; }
        button:hover { background: #005a87; }
        .result { margin-top: 20px; padding: 15px; border-radius: 5px; }
        .success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .example { background: #f8f9fa; padding: 15px; border-radius: 5px; margin-bottom: 20px; border-left: 4px solid #007cba; }
        .example pre { margin: 0; font-size: 12px; overflow-x: auto; }
        .status { padding: 10px; border-radius: 5px; margin-bottom: 20px; text-align: center; }
        .status.connected { background: #d4edda; color: #155724; }
        .status.disconnected { background: #f8d7da; color: #721c24; }
        .config-info { background: #e9ecef; padding: 15px; border-radius: 5px; margin-bottom: 20px; font-size: 14px; }
        .server-info { background: #d1ecf1; padding: 10px; border-radius: 5px; margin-bottom: 15px; text-align: center; }
        @media (max-width: 768px) {
            body { padding: 10px; }
            .container { padding: 20px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>MQ消息发送工具</h1>
        
        <div class="server-info">
            <strong>服务器地址:</strong> {{.ServerAddress}}
        </div>
        
        <div class="config-info">
            <strong>MQ配置信息:</strong><br>
            Topic: dev_search_task_request<br>
            Instance: MQ_INST_1625550118601853_BZenvcEe<br>
            Endpoint: {{.Endpoints}}
        </div>
        
        <div class="status connected">
            ✅ MQ连接状态: 已就绪
        </div>

        <div class="example">
            <h3>消息格式示例:</h3>
            <pre>{
    "requestId": "req_{{.Timestamp}}",
    "requestTime": "{{.CurrentTime}}",
    "tenantId": "tenant_001",
    "companyName": "示例公司",
    "companyWebsite": "https://example.com",
    "contactPersonName": "张三",
    "emailAddress": "zhangsan@example.com",
    "type": 1,
    "location": "北京",
    "position": "软件工程师",
    "importExperience": "3年以上",
    "industryExperience": "互联网"
}</pre>
        </div>

        <form method="POST" action="/send">
            <div class="form-group">
                <label for="message">请输入消息 (JSON格式):</label>
                <textarea id="message" name="message" placeholder="请输入JSON格式的消息..." required>{{.ExampleMessage}}</textarea>
            </div>
            <button type="submit">发送消息</button>
        </form>

        {{if .Result}}
        <div class="result {{if .Success}}success{{else}}error{{end}}">
            {{.Result}}
        </div>
        {{end}}
    </div>
</body>
</html>`

		t, err := template.New("home").Parse(tmpl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		currentTime := time.Now().Format(time.RFC3339)
		timestamp := time.Now().Unix()

		data := struct {
			Endpoints      string
			Timestamp      int64
			CurrentTime    string
			ExampleMessage string
			Result         string
			Success        bool
			ServerAddress  string
		}{
			Endpoints:     mqConfig.Endpoints,
			Timestamp:     timestamp,
			CurrentTime:   currentTime,
			ServerAddress: getServerAddress(r),
			ExampleMessage: fmt.Sprintf(`{
    "requestId": "req_%d",
    "requestTime": "%s",
    "tenantId": "tenant_001",
    "companyName": "示例公司",
    "companyWebsite": "https://example.com",
    "contactPersonName": "张三",
    "emailAddress": "zhangsan@example.com",
    "type": 1,
    "location": "北京",
    "position": "软件工程师",
    "importExperience": "3年以上",
    "industryExperience": "互联网"
}`, timestamp, currentTime),
		}

		t.Execute(w, data)
	}
}

func sendHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	messageJSON := r.FormValue("message")

	type ResponseData struct {
		Result        string
		Success       bool
		Endpoints     string
		ServerAddress string
	}

	var responseData ResponseData
	responseData.Endpoints = mqConfig.Endpoints
	responseData.ServerAddress = getServerAddress(r)

	var taskMessage TaskMessage
	err := json.Unmarshal([]byte(messageJSON), &taskMessage)
	if err != nil {
		responseData.Result = fmt.Sprintf("JSON解析失败: %v", err)
		responseData.Success = false
	} else {
		err = sendMessageToMQ(&taskMessage)
		if err != nil {
			responseData.Result = fmt.Sprintf("发送到MQ失败: %v", err)
			responseData.Success = false
		} else {
			responseData.Result = "消息发送成功!"
			responseData.Success = true
		}
	}

	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>发送结果</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background-color: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .result { margin-top: 20px; padding: 15px; border-radius: 5px; }
        .success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .back-btn { display: inline-block; margin-top: 20px; padding: 10px 20px; background: #6c757d; color: white; text-decoration: none; border-radius: 5px; }
        .back-btn:hover { background: #545b62; }
        .config-info { background: #e9ecef; padding: 10px; border-radius: 5px; margin-bottom: 15px; font-size: 14px; }
        .server-info { background: #d1ecf1; padding: 10px; border-radius: 5px; margin-bottom: 15px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <h1>发送结果</h1>
        
        <div class="server-info">
            <strong>服务器地址:</strong> {{.ServerAddress}}
        </div>
        
        <div class="config-info">
            <strong>MQ配置信息:</strong><br>
            Topic: dev_search_task_request<br>
            Instance: MQ_INST_1625550118601853_BZenvcEe<br>
            Endpoint: {{.Endpoints}}
        </div>
        
        <div class="result {{if .Success}}success{{else}}error{{end}}">
            {{.Result}}
        </div>
        <a href="/" class="back-btn">返回继续发送</a>
    </div>
</body>
</html>`

	t, err := template.New("result").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, responseData)
}

func main() {
	// 初始化配置
	mqConfig = initConfig()
	log.Printf("MQ配置初始化成功, Endpoints: %s, Topic: %s\n", mqConfig.Endpoints, mqConfig.Topic)

	// 初始化MQ生产者
	var err error
	mqProducer, err = initMQProducer()
	if err != nil {
		log.Fatalf("初始化MQ生产者失败: %v", err)
	}
	defer func() {
		err = mqProducer.Shutdown()
		if err != nil {
			log.Printf("关闭MQ生产者失败: %v", err)
		} else {
			log.Println("MQ生产者已关闭")
		}
	}()

	// 设置HTTP路由
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/send", sendHandler)

	// 启动Web服务器 - 监听所有网络接口
	addr := serverHost + ":" + serverPort
	log.Printf("Web服务启动成功，监听地址: %s", addr)
	log.Printf("您可以通过以下地址访问:")
	log.Printf("  本地访问: http://localhost:%s", serverPort)
	log.Printf("  网络访问: http://云机器IP:%s", serverPort)

	log.Fatal(http.ListenAndServe(addr, nil))
}
