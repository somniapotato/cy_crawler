package config

import (
	"cy_crawler/internal/types"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*types.Config, error) {
	var config types.Config

	if configPath == "" {
		// 尝试默认路径
		defaultPaths := []string{
			"./configs/config.toml",
			"/etc/cy_crawler/config.toml",
		}

		for _, path := range defaultPaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}

		if configPath == "" {
			return nil, os.ErrNotExist
		}
	}

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, err
	}

	// 确保日志目录存在
	logDir := filepath.Dir(config.Log.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// 处理环境变量覆盖 - 只覆盖access_key和secret_key
	overrideConfigWithEnv(&config)

	return &config, nil
}

// overrideConfigWithEnv 使用环境变量覆盖配置
func overrideConfigWithEnv(config *types.Config) {
	// 优先从环境变量读取 access_key
	if envAccessKey := os.Getenv("ROCKETMQ_ACCESS_KEY"); envAccessKey != "" {
		config.RocketMQ.Common.AccessKey = envAccessKey
	}

	// 优先从环境变量读取 secret_key
	if envSecretKey := os.Getenv("ROCKETMQ_SECRET_KEY"); envSecretKey != "" {
		config.RocketMQ.Common.SecretKey = envSecretKey
	}
}
