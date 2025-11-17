package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func main() {
	// RocketMQ配置
	nameServer := "127.0.0.1:9876"
	topic := "crawler_tasks_result"
	group := "result_checker_consumer"

	fmt.Printf("Starting result consumer...\n")
	fmt.Printf("Topic: %s\n", topic)
	fmt.Printf("NameServer: %s\n", nameServer)
	fmt.Printf("Consumer Group: %s\n\n", group)

	// 创建消费者
	c, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(group),
		consumer.WithNameServer([]string{nameServer}),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
	)
	if err != nil {
		fmt.Printf("Create consumer error: %s\n", err.Error())
		os.Exit(1)
	}

	// 注册消息处理函数 - 使用正确的selector
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "*",
	}

	err = c.Subscribe(topic, selector,
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range msgs {
				fmt.Printf("=== Received Result Message ===\n")
				fmt.Printf("Message ID: %s\n", msg.MsgId)
				fmt.Printf("Topic: %s\n", msg.Topic)
				fmt.Printf("Born Time: %s\n", time.Unix(msg.BornTimestamp/1000, 0).Format("2006-01-02 15:04:05"))
				fmt.Printf("Store Time: %s\n", time.Unix(msg.StoreTimestamp/1000, 0).Format("2006-01-02 15:04:05"))
				fmt.Printf("Queue ID: %d\n", msg.Queue.QueueId)
				fmt.Printf("Broker: %s\n", msg.Queue.BrokerName)
				fmt.Printf("Reconsume Times: %d\n", msg.ReconsumeTimes)
				fmt.Printf("Body Length: %d bytes\n", len(msg.Body))
				fmt.Printf("\n--- Raw Body Content ---\n")
				fmt.Printf("%s\n", string(msg.Body))
				fmt.Printf("--- End of Body ---\n")
				fmt.Printf("==============================\n\n")
			}
			return consumer.ConsumeSuccess, nil
		})

	if err != nil {
		fmt.Printf("Subscribe error: %s\n", err.Error())
		os.Exit(1)
	}

	// 启动消费者
	err = c.Start()
	if err != nil {
		fmt.Printf("Start consumer error: %s\n", err.Error())
		os.Exit(1)
	}
	defer c.Shutdown()

	fmt.Printf("Result consumer started successfully! Waiting for messages...\n")
	fmt.Printf("Press Ctrl+C to exit.\n\n")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Printf("\nShutting down result consumer...\n")
}
