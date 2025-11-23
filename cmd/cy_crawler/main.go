package main

import (
	"cy_crawler/internal/config"
	"cy_crawler/internal/logger"
	"cy_crawler/internal/mq"
	"cy_crawler/internal/processor"
	"cy_crawler/internal/types"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	// 定义命令行参数
	configPath := flag.String("config", "", "配置文件路径（可选，留空使用默认配置）")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
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

	// 消息处理函数 - 使用defer确保所有错误都返回到MQ
	messageHandler := func(task *types.TaskMessage) error {
		var result *types.ResultMessage
		var err error

		// 使用defer来确保任何错误都会被发送回MQ
		defer func() {
			if err != nil && result != nil {
				// 如果有错误且result不为空，直接发送
				logger.Logger.WithFields(logrus.Fields{
					"requestId": task.RequestID,
					"task":      task,
					"code":      result.Code,
					"message":   result.Message,
					"error":     err.Error(),
				}).Error("Task processing failed")
				_ = producer.SendResult(result)
			} else if err != nil && result == nil {
				// 如果有错误但没有result，构建默认错误结果
				errorResult := types.BuildErrorResult(types.CodeInternalServerError, err.Error(), task)
				logger.Logger.WithFields(logrus.Fields{
					"requestId": task.RequestID,
					"task":      task,
					"code":      errorResult.Code,
					"message":   errorResult.Message,
					"error":     err.Error(),
				}).Error("Task processing failed")
				_ = producer.SendResult(errorResult)
			} else if result != nil {
				// 成功处理，发送结果
				logger.Logger.WithFields(logrus.Fields{
					"requestId": task.RequestID,
					"task":      task,
					"code":      result.Code,
					"message":   result.Message,
				}).Info("Task processed successfully")
				_ = producer.SendResult(result)
			}
		}()

		// 正常处理任务
		result, err = proc.ProcessTask(task)
		return err
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
