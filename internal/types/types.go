package types

// TaskMessage 从MQ接收的任务消息
type TaskMessage struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Email   string `json:"email"`
	Country string `json:"country"`
}

// ResultMessage 处理结果消息
type ResultMessage struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	TaskID  string      `json:"task_id,omitempty"`
}

// Config 应用配置
type Config struct {
	RocketMQ struct {
		NameServer    string `toml:"name_server"`
		ConsumerGroup string `toml:"consumer_group"`
		ProducerGroup string `toml:"producer_group"`
		ConsumerTopic string `toml:"consumer_topic"`
		ProducerTopic string `toml:"producer_topic"`
	} `toml:"rocketmq"`

	Log struct {
		Level      string `toml:"level"`
		FilePath   string `toml:"file_path"`
		MaxSize    int    `toml:"max_size"`
		MaxBackups int    `toml:"max_backups"`
		MaxAge     int    `toml:"max_age"`
	} `toml:"log"`

	Application struct {
		PythonScriptPath  string `toml:"python_script_path"`
		HeartbeatInterval int    `toml:"heartbeat_interval"`
	} `toml:"application"`
}
