package types

import "fmt"

// ErrorCode 错误码常量
const (
	// 系统错误 (500系列)
	CodeInternalServerError = 500
	CodeServiceUnavailable  = 503
	CodeTimeout             = 504

	// 输入验证错误 (400系列)
	CodeBadRequest        = 400
	CodeValidationError   = 422
	CodeParameterRequired = 421

	// 外部服务错误 (502-504系列)
	CodeBadGateway           = 502
	CodeExternalServiceError = 505
)

// ErrorMessage 错误消息映射
var ErrorMessage = map[int]string{
	CodeInternalServerError:  "内部服务器错误",
	CodeServiceUnavailable:   "服务不可用",
	CodeTimeout:              "请求超时",
	CodeBadRequest:           "请求参数错误",
	CodeValidationError:      "输入验证失败",
	CodeParameterRequired:    "必需参数缺失",
	CodeBadGateway:           "网关错误",
	CodeExternalServiceError: "外部服务错误",
}

// GetErrorMessage 获取错误码对应的错误消息
func GetErrorMessage(code int) string {
	if msg, exists := ErrorMessage[code]; exists {
		return msg
	}
	return "未知错误"
}

// BuildErrorResult 构建错误结果消息
func BuildErrorResult(code int, message string, params *TaskMessage) *ResultMessage {
	// 如果没有提供具体消息，使用默认消息
	if message == "" {
		message = GetErrorMessage(code)
	}

	return &ResultMessage{
		Code:    code,
		Message: message,
		Data:    []interface{}{},
		Params:  params,
	}
}

// BuildPythonErrorResult 构建Python脚本错误结果
func BuildPythonErrorResult(err error, params *TaskMessage) *ResultMessage {
	message := fmt.Sprintf("Python脚本执行失败: %v", err)
	return BuildErrorResult(CodeInternalServerError, message, params)
}

// BuildValidationErrorResult 构建验证错误结果
func BuildValidationErrorResult(message string, params *TaskMessage) *ResultMessage {
	return BuildErrorResult(CodeValidationError, message, params)
}
