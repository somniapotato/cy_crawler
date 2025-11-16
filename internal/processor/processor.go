package processor

import (
	"cy_crawler/internal/logger"
	"cy_crawler/internal/types"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

type Processor struct {
	pythonScriptPath string
}

// NewProcessor 创建新的处理器
func NewProcessor(pythonScriptPath string) *Processor {
	return &Processor{
		pythonScriptPath: pythonScriptPath,
	}
}

// ProcessTask 处理任务
func (p *Processor) ProcessTask(task *types.TaskMessage) (*types.ResultMessage, error) {
	logger.Logger.WithFields(logrus.Fields{
		"type":    task.Type,
		"name":    task.Name,
		"url":     task.URL,
		"email":   task.Email,
		"country": task.Country,
	}).Info("Processing task")

	// 构建命令行参数
	args := []string{
		p.pythonScriptPath,
		"--type", task.Type,
		"--name", task.Name,
		"--url", task.URL,
	}

	if task.Email != "" {
		args = append(args, "--email", task.Email)
	}
	if task.Country != "" {
		args = append(args, "--country", task.Country)
	}

	// 执行Python脚本
	cmd := exec.Command("python", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"task":   task,
			"error":  err.Error(),
			"output": string(output),
		}).Error("Failed to execute Python script")

		return &types.ResultMessage{
			Success: false,
			Error:   fmt.Sprintf("Python script execution failed: %v", err),
		}, err
	}

	// 清理输出：移除可能的多余字符
	cleanedOutput := cleanPythonOutput(output)

	// 解析Python脚本输出
	var resultData interface{}
	if err := json.Unmarshal([]byte(cleanedOutput), &resultData); err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"task":          task,
			"rawOutput":     string(output),
			"cleanedOutput": cleanedOutput,
			"error":         err.Error(),
		}).Error("Failed to parse Python script output")

		return &types.ResultMessage{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse script output: %v. Raw output: %s", err, string(output)),
		}, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"task":   task,
		"result": resultData,
	}).Info("Successfully processed task")

	return &types.ResultMessage{
		Success: true,
		Data:    resultData,
	}, nil
}

// cleanPythonOutput 清理Python脚本输出
func cleanPythonOutput(output []byte) string {
	str := string(output)

	// 移除BOM头
	if len(str) >= 3 && str[0] == 0xEF && str[1] == 0xBB && str[2] == 0xBF {
		str = str[3:]
	}

	// 移除首尾空白字符
	str = strings.TrimSpace(str)

	// 尝试找到JSON开始和结束位置
	start := strings.Index(str, "{")
	end := strings.LastIndex(str, "}")

	if start != -1 && end != -1 && end > start {
		str = str[start : end+1]
	}

	// 移除可能的Python错误前缀
	lines := strings.Split(str, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			return line
		}
	}

	return str
}

// ValidatePythonEnvironment 验证Python环境
func (p *Processor) ValidatePythonEnvironment() error {
	// 检查Python是否可用
	if _, err := exec.LookPath("python"); err != nil {
		return fmt.Errorf("python not found in PATH: %v", err)
	}

	// 检查脚本文件是否存在
	cmd := exec.Command("python", "-c", "import sys; print(sys.version)")
	if _, err := cmd.Output(); err != nil {
		return fmt.Errorf("python is not working properly: %v", err)
	}

	return nil
}
