package types

// TaskMessage 从MQ接收的任务消息（新格式）
type TaskMessage struct {
	RequestID          string `json:"requestId"`
	RequestTime        string `json:"requestTime"`
	TenantID           string `json:"tenantId"`
	CompanyName        string `json:"companyName"`
	CompanyWebsite     string `json:"companyWebsite"`
	ContactPersonName  string `json:"contactPersonName"`
	EmailAddress       string `json:"emailAddress"`
	Type               int    `json:"type"` // 1: 公司, 2: 个人
	Location           string `json:"location"`
	Position           string `json:"position"`
	ImportExperience   string `json:"importExperience"`
	IndustryExperience string `json:"industryExperience"`
}

// ResultMessage 处理结果消息（新格式）
type ResultMessage struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data"`
	Params  *TaskMessage `json:"params"`
}

// PythonResult Python脚本返回的数据结构
type PythonResult struct {
	Sources map[string]interface{} `json:"sources"`
}

// Config 应用配置
type Config struct {
	RocketMQ struct {
		Common struct {
			Endpoints  string `toml:"endpoints"`
			AccessKey  string `toml:"access_key"`
			SecretKey  string `toml:"secret_key"`
			InstanceID string `toml:"instance_id"`
		} `toml:"common"`
		BGCheck struct {
			Consumer struct {
				Topic string `toml:"topic"`
				Group string `toml:"group"`
				Tag   string `toml:"tag"`
			} `toml:"consumer"`
			Producer struct {
				Topic string `toml:"topic"`
			} `toml:"producer"`
		} `toml:"bgCheck"`
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
