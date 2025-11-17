package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

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

func main() {
	// RocketMQ配置
	nameServer := "127.0.0.1:9876"
	topic := "crawler_tasks"
	group := "test_producer"

	// 创建生产者
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{nameServer}),
		producer.WithRetry(2),
		producer.WithGroupName(group),
	)
	if err != nil {
		fmt.Printf("Create producer error: %s\n", err.Error())
		os.Exit(1)
	}

	err = p.Start()
	if err != nil {
		fmt.Printf("Start producer error: %s\n", err.Error())
		os.Exit(1)
	}
	defer p.Shutdown()

	// 测试消息 - 公司类型
	testMessages := []TaskMessage{
		{
			RequestID:          "6352d81f-1217-4c73-aa11-4031a1daf7c0",
			RequestTime:        "2025-11-23 22:22:22",
			TenantID:           "122",
			CompanyName:        "BioGenex",
			CompanyWebsite:     "www.baidu.com",
			ContactPersonName:  "张三",
			EmailAddress:       "duxu111@126.com",
			Type:               1, // 公司
			Location:           "意大利",
			Position:           "General Manager",
			ImportExperience:   "有",
			IndustryExperience: "互联网",
		},
		// {
		// 	RequestID:          "7352d81f-1217-4c73-aa11-4031a1daf7c1",
		// 	RequestTime:        "2025-11-23 22:23:00",
		// 	TenantID:           "123",
		// 	CompanyName:        "TECH CORP",
		// 	CompanyWebsite:     "www.techcorp.com",
		// 	ContactPersonName:  "李四",
		// 	EmailAddress:       "lisi@techcorp.com",
		// 	Type:               2, // 个人
		// 	Location:           "中国",
		// 	Position:           "Software Engineer",
		// 	ImportExperience:   "无",
		// 	IndustryExperience: "IT",
		// },
	}

	// 发送消息
	for i, msg := range testMessages {
		data, _ := json.Marshal(msg)

		message := &primitive.Message{
			Topic: topic,
			Body:  data,
		}

		res, err := p.SendSync(context.Background(), message)
		if err != nil {
			fmt.Printf("Send message error: %s\n", err.Error())
		} else {
			fmt.Printf("Send message %d success: %s\n", i+1, res.MsgID)
			fmt.Printf("Message content: %s\n", string(data))
		}
	}

	fmt.Println("All test messages sent!")
}
