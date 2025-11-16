package mq

import (
	"context"
	"cy_crawler/internal/logger"
	"cy_crawler/internal/types"
	"encoding/json"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	client rocketmq.PushConsumer
	config *types.Config
}

// NewConsumer 创建新的消费者
func NewConsumer(config *types.Config, messageHandler func(*types.TaskMessage) error) (*Consumer, error) {
	c, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName(config.RocketMQ.ConsumerGroup),
		consumer.WithNameServer([]string{config.RocketMQ.NameServer}),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromLastOffset), // 明确从最新位置开始
		consumer.WithConsumerOrder(true),                              // 顺序消费
	)
	if err != nil {
		return nil, err
	}

	consumerInst := &Consumer{
		client: c,
		config: config,
	}

	// 注册消息处理函数 - 使用正确的selector
	selector := consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "*",
	}

	err = c.Subscribe(config.RocketMQ.ConsumerTopic, selector,
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, msg := range msgs {
				if err := consumerInst.handleMessage(msg, messageHandler); err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"topic": msg.Topic,
						"msgId": msg.MsgId,
						"error": err.Error(),
					}).Error("Failed to handle message")
					// 继续处理其他消息，不返回错误
				}
			}
			return consumer.ConsumeSuccess, nil
		})

	if err != nil {
		return nil, err
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
	if msg.Type == "" {
		return fmt.Errorf("type field is required")
	}
	if msg.Name == "" {
		return fmt.Errorf("name field is required")
	}
	if msg.URL == "" {
		return fmt.Errorf("url field is required")
	}
	return nil
}

// Start 启动消费者
func (c *Consumer) Start() error {
	return c.client.Start()
}

// Shutdown 关闭消费者
func (c *Consumer) Shutdown() error {
	return c.client.Shutdown()
}
