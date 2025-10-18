package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorCode 错误代码
type ErrorCode int

const (
	// 通用错误
	ErrInvalidRequest ErrorCode = 40000 // 无效请求
	ErrUnauthorized   ErrorCode = 40100 // 未授权
	ErrForbidden      ErrorCode = 40300 // 禁止访问
	ErrNotFound       ErrorCode = 40400 // 资源不存在
	ErrConflict       ErrorCode = 40900 // 资源冲突
	ErrInternalServer ErrorCode = 50000 // 服务器内部错误

	// 文件相关错误
	ErrFileInvalid     ErrorCode = 41001 // 无效文件
	ErrFileTooBig      ErrorCode = 41002 // 文件过大
	ErrFileTypeInvalid ErrorCode = 41003 // 文件类型不支持
	ErrFileNameInvalid ErrorCode = 41004 // 文件名不合法
	ErrPathInvalid     ErrorCode = 41005 // 路径不合法
	ErrFileNotFound    ErrorCode = 41006 // 文件不存在
	ErrFileExists      ErrorCode = 41007 // 文件已存在

	// 目录相关错误
	ErrDirInvalid   ErrorCode = 42001 // 无效目录
	ErrDirNotFound  ErrorCode = 42002 // 目录不存在
	ErrDirExists    ErrorCode = 42003 // 目录已存在
	ErrDirNotEmpty  ErrorCode = 42004 // 目录不为空

	// 存储相关错误
	ErrStorageFull    ErrorCode = 43001 // 存储空间不足
	ErrStorageFailure ErrorCode = 43002 // 存储操作失败
)

// ErrorMessage 错误消息映射
var ErrorMessages = map[ErrorCode]string{
	ErrInvalidRequest: "无效的请求参数",
	ErrUnauthorized:   "未授权访问",
	ErrForbidden:      "禁止访问",
	ErrNotFound:       "资源不存在",
	ErrConflict:       "资源冲突",
	ErrInternalServer: "服务器内部错误",

	ErrFileInvalid:     "无效的文件",
	ErrFileTooBig:      "文件大小超过限制",
	ErrFileTypeInvalid: "不支持的文件类型",
	ErrFileNameInvalid: "文件名不合法",
	ErrPathInvalid:     "路径不合法",
	ErrFileNotFound:    "文件不存在",
	ErrFileExists:      "文件已存在",

	ErrDirInvalid:   "无效的目录",
	ErrDirNotFound:  "目录不存在",
	ErrDirExists:    "目录已存在",
	ErrDirNotEmpty:  "目录不为空",

	ErrStorageFull:    "存储空间不足",
	ErrStorageFailure: "存储操作失败",
}

// Response 统一响应结构
type Response struct {
	Success   bool        `json:"success"`
	Code      ErrorCode   `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"requestId,omitempty"`
}

// Success 成功响应
func Success(ctx *gin.Context, data interface{}) {
	response := Response{
		Success:   true,
		Code:      0,
		Message:   "操作成功",
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(ctx),
	}
	ctx.JSON(http.StatusOK, response)
}

// Error 错误响应
func Error(ctx *gin.Context, code ErrorCode, customMessage ...string) {
	message := ErrorMessages[code]
	if len(customMessage) > 0 && customMessage[0] != "" {
		message = customMessage[0]
	}

	httpStatus := getHTTPStatus(code)
	response := Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(ctx),
	}
	ctx.JSON(httpStatus, response)
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(ctx *gin.Context, code ErrorCode, data interface{}, customMessage ...string) {
	message := ErrorMessages[code]
	if len(customMessage) > 0 && customMessage[0] != "" {
		message = customMessage[0]
	}

	httpStatus := getHTTPStatus(code)
	response := Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(ctx),
	}
	ctx.JSON(httpStatus, response)
}

// getHTTPStatus 根据错误代码获取HTTP状态码
func getHTTPStatus(code ErrorCode) int {
	switch {
	case code >= 40000 && code < 41000:
		return http.StatusBadRequest
	case code >= 41000 && code < 42000:
		return http.StatusBadRequest
	case code >= 42000 && code < 43000:
		return http.StatusBadRequest
	case code >= 43000 && code < 44000:
		return http.StatusInternalServerError
	case code == ErrUnauthorized:
		return http.StatusUnauthorized
	case code == ErrForbidden:
		return http.StatusForbidden
	case code == ErrNotFound:
		return http.StatusNotFound
	case code == ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// getRequestID 获取请求ID
func getRequestID(ctx *gin.Context) string {
	if requestID := ctx.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID := ctx.GetString("requestId"); requestID != "" {
		return requestID
	}
	return ""
}