package logger

import (
	"cy_crawler/internal/types"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// InitLogger 初始化日志系统
func InitLogger(config *types.Config) error {
	Logger = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Log.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// 创建日志文件
	file, err := os.OpenFile(config.Log.FilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	Logger.SetOutput(file)

	// 设置日志格式
	Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	return nil
}

// StartHeartbeatLogger 启动心跳日志
func StartHeartbeatLogger(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		Logger.WithFields(logrus.Fields{
			"component": "heartbeat",
			"timestamp": time.Now().Unix(),
		}).Info("Application heartbeat")
	}
}
