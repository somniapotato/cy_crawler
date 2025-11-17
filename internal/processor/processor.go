package processor

import (
	"bytes"
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
		"requestId":   task.RequestID,
		"tenantId":    task.TenantID,
		"companyName": task.CompanyName,
		"type":        task.Type,
		"location":    task.Location,
	}).Info("Processing task")

	// 根据type决定name参数
	nameParam := task.CompanyName
	if task.Type == 2 { // 个人
		nameParam = task.ContactPersonName
	}

	// 构建命令行参数
	args := []string{
		p.pythonScriptPath,
		"--type", getTypeString(task.Type),
		"--name", nameParam,
		"--url", task.CompanyWebsite,
	}

	if task.EmailAddress != "" {
		args = append(args, "--email", task.EmailAddress)
	}
	if task.Location != "" {
		args = append(args, "--country", task.Location)
	}

	// 执行Python脚本
	cmd := exec.Command("python", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.Bytes()
	errorOutput := stderr.Bytes()

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestId": task.RequestID,
			"error":     err.Error(),
			"stdout":    string(output),
			"stderr":    string(errorOutput),
		}).Error("Failed to execute Python script")

		return &types.ResultMessage{
			Code:    500,
			Message: fmt.Sprintf("Python script execution failed: %v", err),
			Data:    nil,
			Params:  task,
		}, err
	}

	// 如果有stderr输出但命令成功，记录警告
	if len(errorOutput) > 0 {
		logger.Logger.WithFields(logrus.Fields{
			"requestId": task.RequestID,
			"stderr":    string(errorOutput),
		}).Warn("Python script produced stderr output")
	}

	// 清理输出
	cleanedOutput := cleanPythonOutput(output)

	// 解析Python脚本输出
	var pythonResult types.PythonResult
	if err := json.Unmarshal([]byte(cleanedOutput), &pythonResult); err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestId":     task.RequestID,
			"rawOutput":     string(output),
			"cleanedOutput": cleanedOutput,
			"error":         err.Error(),
		}).Error("Failed to parse Python script output")

		return &types.ResultMessage{
			Code:    500,
			Message: fmt.Sprintf("Failed to parse script output: %v", err),
			Data:    nil,
			Params:  task,
		}, err
	}

	// 组装最终结果
	finalData := []types.FinalResult{
		{
			Sources: pythonResult.Sources,
		},
	}

	logger.Logger.WithFields(logrus.Fields{
		"requestId": task.RequestID,
		"sources":   len(pythonResult.Sources),
	}).Info("Successfully processed task")

	return &types.ResultMessage{
		Code:    200,
		Message: "success",
		Data:    finalData,
		Params:  task,
	}, nil
}

// getTypeString 将type数字转换为字符串
func getTypeString(typeNum int) string {
	if typeNum == 1 {
		return "company"
	} else if typeNum == 2 {
		return "person"
	}
	return "unknown"
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
