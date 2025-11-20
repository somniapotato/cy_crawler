package mq

import (
	"context"
	"cy_crawler/internal/logger"
	"cy_crawler/internal/types"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	client rocketmq.PushConsumer
	config *types.Config
}

// NewConsumer 创建支持阿里云的消费者
func NewConsumer(config *types.Config, messageHandler func(*types.TaskMessage) error) (*Consumer, error) {
	// 阿里云 RocketMQ 配置
	endpoints := config.RocketMQ.Common.Endpoints

	// 创建消费者选项
	opts := []consumer.Option{
		consumer.WithGroupName(config.RocketMQ.BGCheck.Consumer.Group),
		consumer.WithNameServer([]string{endpoints}),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset),
		consumer.WithCredentials(primitive.Credentials{
			AccessKey: config.RocketMQ.Common.AccessKey,
			SecretKey: config.RocketMQ.Common.SecretKey,
		}),
		consumer.WithNamespace(config.RocketMQ.Common.InstanceID),
	}

	// 创建消费者
	c, err := rocketmq.NewPushConsumer(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}

	consumerInst := &Consumer{
		client: c,
		config: config,
	}

	// 使用配置中的tag
	tag := config.RocketMQ.BGCheck.Consumer.Tag
	if tag == "" {
		tag = "*"
	}

	// 注册消息处理函数
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: tag,
	}

	err = c.Subscribe(config.RocketMQ.BGCheck.Consumer.Topic, selector,
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range msgs {
				if err := consumerInst.handleMessage(msg, messageHandler); err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"topic": msg.Topic,
						"msgId": msg.MsgId,
						"error": err.Error(),
					}).Error("Failed to handle message")
				}
			}
			return consumer.ConsumeSuccess, nil
		})

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %v", err)
	}

	return consumerInst, nil
}

// handleMessage 处理单个消息
func (c *Consumer) handleMessage(msg *primitive.MessageExt, handler func(*types.TaskMessage) error) error {
	logger.Logger.WithFields(logrus.Fields{
		"topic": msg.Topic,
		"msgId": msg.MsgId,
		"body":  string(msg.Body),
	}).Info("Received message")

	var taskMsg types.TaskMessage
	if err := json.Unmarshal(msg.Body, &taskMsg); err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"msgId": msg.MsgId,
			"body":  string(msg.Body),
			"error": err.Error(),
		}).Error("Failed to parse message body")
		return err
	}

	// 验证必需字段
	if err := c.validateTaskMessage(&taskMsg); err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"msgId": msg.MsgId,
			"task":  taskMsg,
			"error": err.Error(),
		}).Error("Invalid task message")
		return err
	}

	return handler(&taskMsg)
}

// validateTaskMessage 验证任务消息
func (c *Consumer) validateTaskMessage(msg *types.TaskMessage) error {
	if msg.RequestID == "" {
		return fmt.Errorf("requestId field is required")
	}
	if msg.TenantID == "" {
		return fmt.Errorf("tenantId field is required")
	}
	if msg.Type != "1" && msg.Type != "2" {
		return fmt.Errorf("type must be \"1\" (company) or \"2\" (person)")
	}
	if msg.Type == "1" && msg.CompanyName == "" {
		return fmt.Errorf("companyName is required when type is \"1\" (company)")
	}
	if msg.Type == "2" && msg.ContactPersonName == "" {
		return fmt.Errorf("contactPersonName is required when type is \"2\" (person)")
	}
	if msg.CompanyWebsite == "" {
		return fmt.Errorf("companyWebsite field is required")
	}
	return nil
}

// Start 启动消费者
func (c *Consumer) Start() error {
	// 添加启动重试逻辑
	var err error
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		err = c.client.Start()
		if err == nil {
			logger.Logger.Info("Consumer started successfully")
			return nil
		}

		logger.Logger.WithFields(logrus.Fields{
			"attempt":    i + 1,
			"maxRetries": maxRetries,
			"error":      err.Error(),
		}).Warn("Failed to start consumer, retrying...")

		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	return fmt.Errorf("failed to start consumer after %d attempts: %v", maxRetries, err)
}

// Shutdown 关闭消费者
func (c *Consumer) Shutdown() error {
	return c.client.Shutdown()
}
