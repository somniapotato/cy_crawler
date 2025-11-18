package mq

import (
	"context"
	"cy_crawler/internal/logger"
	"cy_crawler/internal/types"
	"encoding/json"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/sirupsen/logrus"
)

type Producer struct {
	client rocketmq.Producer
	config *types.Config
}

// NewProducer 创建支持阿里云的生产者
func NewProducer(config *types.Config) (*Producer, error) {
	// 阿里云 RocketMQ 配置
	endpoints := config.RocketMQ.Common.Endpoints

	// 创建生产者选项
	opts := []producer.Option{
		producer.WithGroupName(config.RocketMQ.BGCheck.Consumer.Group + "_producer"),
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
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	err = p.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start producer: %v", err)
	}

	return &Producer{
		client: p,
		config: config,
	}, nil
}

// SendResult 发送处理结果
func (p *Producer) SendResult(result *types.ResultMessage) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	msg := &primitive.Message{
		Topic: p.config.RocketMQ.BGCheck.Producer.Topic,
		Body:  data,
	}

	ctx := context.Background()
	res, err := p.client.SendSync(ctx, msg)

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"topic": msg.Topic,
			"error": err.Error(),
		}).Error("Failed to send result message")
		return err
	}

	logger.Logger.WithFields(logrus.Fields{
		"topic":     msg.Topic,
		"msgId":     res.MsgID,
		"code":      result.Code,
		"message":   result.Message,
		"requestId": getRequestID(result.Params),
	}).Info("Successfully sent result message")

	return nil
}

// getRequestID 从params中获取requestId
func getRequestID(params *types.TaskMessage) string {
	if params != nil {
		return params.RequestID
	}
	return "unknown"
}

// Shutdown 关闭生产者
func (p *Producer) Shutdown() error {
	return p.client.Shutdown()
}
