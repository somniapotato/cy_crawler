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

	return &config, nil
}
