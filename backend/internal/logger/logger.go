package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger 日志记录器
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var defaultLogger *Logger

// Init 初始化日志系统
func Init(logLevel LogLevel, logFile string) error {
	var output *os.File
	var err error

	if logFile != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %v", err)
		}

		output, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("打开日志文件失败: %v", err)
		}
	} else {
		output = os.Stdout
	}

	defaultLogger = &Logger{
		level:  logLevel,
		logger: log.New(output, "", 0),
	}

	return nil
}

// GetLogger 获取默认日志记录器
func GetLogger() *Logger {
	if defaultLogger == nil {
		// 如果没有初始化，使用默认配置
		Init(INFO, "")
	}
	return defaultLogger
}

// formatMessage 格式化日志消息
func (l *Logger) formatMessage(level LogLevel, message string) string {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, file, line, _ := runtime.Caller(3)
	filename := filepath.Base(file)
	return fmt.Sprintf("[%s] %s %s:%d - %s", now, levelNames[level], filename, line, message)
}

// log 记录日志
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)
	formatted := l.formatMessage(level, message)
	l.logger.Println(formatted)

	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal 致命错误日志
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// 全局日志函数
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}

// LogRequest 记录请求日志
func LogRequest(ctx *gin.Context, message string, args ...interface{}) {
	requestID := ctx.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	userAgent := ctx.GetHeader("User-Agent")
	clientIP := ctx.ClientIP()
	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	logMessage := fmt.Sprintf("[%s] %s %s %s %s - %s", 
		requestID, clientIP, method, path, userAgent, fmt.Sprintf(message, args...))
	
	GetLogger().Info(logMessage)
}

// LogError 记录错误日志
func LogError(ctx *gin.Context, err error, message string, args ...interface{}) {
	requestID := ctx.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	method := ctx.Request.Method
	path := ctx.Request.URL.Path

	logMessage := fmt.Sprintf("[%s] %s %s - %s - Error: %v", 
		requestID, method, path, fmt.Sprintf(message, args...), err)
	
	GetLogger().Error(logMessage)
}

// LogFileOperation 记录文件操作日志
func LogFileOperation(ctx *gin.Context, operation, filename, path string, size int64) {
	requestID := ctx.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	clientIP := ctx.ClientIP()
	
	logMessage := fmt.Sprintf("[%s] %s - %s: %s (path: %s, size: %d bytes)", 
		requestID, clientIP, operation, filename, path, size)
	
	GetLogger().Info(logMessage)
}

// LogSecurityEvent 记录安全事件日志
func LogSecurityEvent(ctx *gin.Context, event, details string) {
	requestID := ctx.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "unknown"
	}

	clientIP := ctx.ClientIP()
	userAgent := ctx.GetHeader("User-Agent")
	
	logMessage := fmt.Sprintf("[SECURITY] [%s] %s - %s: %s (UA: %s)", 
		requestID, clientIP, event, details, userAgent)
	
	GetLogger().Warn(logMessage)
}