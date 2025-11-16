package main

import (
	"cy_crawler/internal/config"
	"cy_crawler/internal/logger"
	"cy_crawler/internal/mq"
	"cy_crawler/internal/processor"
	"cy_crawler/internal/types"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// 初始化日志
	if err := logger.InitLogger(cfg); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// 启动心跳日志
	go logger.StartHeartbeatLogger(cfg.Application.HeartbeatInterval)

	// 初始化处理器
	proc := processor.NewProcessor(cfg.Application.PythonScriptPath)

	// 验证Python环境
	if err := proc.ValidatePythonEnvironment(); err != nil {
		logger.Logger.WithError(err).Fatal("Python environment validation failed")
	}

	// 初始化生产者
	producer, err := mq.NewProducer(cfg)
	if err != nil {
		logger.Logger.WithError(err).Fatal("Failed to create producer")
	}
	defer producer.Shutdown()

	// 消息处理函数
	messageHandler := func(task *types.TaskMessage) error {
		result, err := proc.ProcessTask(task)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"task":  task,
				"error": err.Error(),
			}).Error("Task processing failed")

			// 发送失败结果
			errorResult := &types.ResultMessage{
				Success: false,
				Error:   err.Error(),
			}
			return producer.SendResult(errorResult)
		}

		// 发送成功结果
		return producer.SendResult(result)
	}

	// 初始化消费者
	consumer, err := mq.NewConsumer(cfg, messageHandler)
	if err != nil {
		logger.Logger.WithError(err).Fatal("Failed to create consumer")
	}

	// 启动消费者
	if err := consumer.Start(); err != nil {
		logger.Logger.WithError(err).Fatal("Failed to start consumer")
	}
	defer consumer.Shutdown()

	logger.Logger.Info("CyCrawler application started successfully")

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Logger.Info("CyCrawler application shutting down")
}
