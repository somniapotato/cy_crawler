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
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	client        rocketmq.PushConsumer
	config        *types.Config
	errorProducer *Producer
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

	// 创建生产者用于发送错误结果
	errorProducer, err := newProducerFromConsumerConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create error producer: %v", err)
	}

	consumerInst := &Consumer{
		client:        c,
		config:        config,
		errorProducer: errorProducer,
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
				if err := consumerInst.handleMessageWithErrorHandling(msg, messageHandler); err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"topic": msg.Topic,
						"msgId": msg.MsgId,
						"error": err.Error(),
					}).Error("Failed to handle message with error handling")
				}
			}
			return consumer.ConsumeSuccess, nil
		})

	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %v", err)
	}

	return consumerInst, nil
}

// newProducerFromConsumerConfig 创建用于错误处理的生产者
func newProducerFromConsumerConfig(config *types.Config) (*Producer, error) {
	// 阿里云 RocketMQ 配置
	endpoints := config.RocketMQ.Common.Endpoints

	// 创建生产者选项
	opts := []producer.Option{
		producer.WithGroupName(config.RocketMQ.BGCheck.Consumer.Group + "_error_producer"),
		producer.WithNameServer([]string{endpoints}),
		producer.WithRetry(2),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: config.RocketMQ.Common.AccessKey,
			SecretKey: config.RocketMQ.Common.SecretKey,
		}),
		producer.WithNamespace(config.RocketMQ.Common.InstanceID),
	}

	p, err := rocketmq.NewProducer(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create error producer: %v", err)
	}

	err = p.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start error producer: %v", err)
	}

	return &Producer{
		client: p,
		config: config,
	}, nil
}

// handleMessageWithErrorHandling 带错误处理的消息处理函数
func (c *Consumer) handleMessageWithErrorHandling(msg *primitive.MessageExt, handler func(*types.TaskMessage) error) error {
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

		// 构建JSON解析错误结果
		errorResult := types.BuildErrorResult(types.CodeBadRequest, fmt.Sprintf("JSON解析失败: %v", err), &taskMsg)
		_ = c.sendErrorResult(errorResult)
		return err
	}

	// 验证必需字段
	if err := c.validateTaskMessage(&taskMsg); err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"msgId": msg.MsgId,
			"task":  taskMsg,
			"error": err.Error(),
		}).Error("Invalid task message")

		// 构建验证错误结果
		errorResult := types.BuildValidationErrorResult(err.Error(), &taskMsg)
		_ = c.sendErrorResult(errorResult)
		return err
	}

	// 正常处理任务
	return handler(&taskMsg)
}

// sendErrorResult 发送错误结果到MQ
func (c *Consumer) sendErrorResult(result *types.ResultMessage) error {
	if c.errorProducer == nil {
		logger.Logger.WithFields(logrus.Fields{
			"code":    result.Code,
			"message": result.Message,
		}).Error("Error producer is nil, cannot send error result")
		return fmt.Errorf("error producer is nil")
	}

	return c.errorProducer.SendResult(result)
}

// handleMessage 旧的函数保留以兼容
func (c *Consumer) handleMessage(msg *primitive.MessageExt, handler func(*types.TaskMessage) error) error {
	return c.handleMessageWithErrorHandling(msg, handler)
}

// validateTaskMessage 验证任务消息
func (c *Consumer) validateTaskMessage(msg *types.TaskMessage) error {
	if msg.RequestID == "" {
		return fmt.Errorf("requestId field is required")
	}
	if msg.TenantID == "" {
		return fmt.Errorf("tenantId field is required")
	}
	if msg.Type != 1 && msg.Type != 2 {
		return fmt.Errorf("type must be 1 (company) or 2 (person)")
	}
	// 使用 TaskMessage 的 Validate 方法进行条件性必填验证
	return msg.Validate()
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
	if c.errorProducer != nil {
		c.errorProducer.Shutdown()
	}
	return c.client.Shutdown()
}
