package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/CoffeeSwt/bilibili-tts-chat/config"
)

// LogLevel 日志级别类型
type LogLevel string

// 日志级别常量
const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// ANSI 颜色代码常量
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorWhite  = "\033[97m"
)

// Logger 日志记录器结构体
type Logger struct {
	output io.Writer
}

// 默认日志记录器实例
var defaultLogger = &Logger{
	output: os.Stdout,
}

// colorizeLevel 为日志级别添加颜色
func colorizeLevel(level LogLevel) string {
	switch level {
	case DEBUG:
		return ColorCyan + string(level) + ColorReset
	case INFO:
		return ColorBlue + string(level) + ColorReset
	case WARN:
		return ColorYellow + string(level) + ColorReset
	case ERROR:
		return ColorRed + string(level) + ColorReset
	default:
		return string(level)
	}
}

// formatLogForConsole 格式化控制台日志（带颜色）
func formatLogForConsole(level LogLevel, timestamp, location, message string, keyvals ...interface{}) string {
	coloredLevel := colorizeLevel(level)
	logEntry := fmt.Sprintf("[%s] [%s] %s %s", timestamp, coloredLevel, location, message)

	// 处理键值对参数
	if len(keyvals) > 0 {
		var pairs []string
		for i := 0; i < len(keyvals); i += 2 {
			if i+1 < len(keyvals) {
				key := fmt.Sprintf("%v", keyvals[i])
				value := fmt.Sprintf("%v", keyvals[i+1])
				pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
			} else {
				// 如果是奇数个参数，最后一个作为单独的值
				pairs = append(pairs, fmt.Sprintf("extra=%v", keyvals[i]))
			}
		}
		if len(pairs) > 0 {
			logEntry += " | " + strings.Join(pairs, " ")
		}
	}

	return logEntry
}

// formatLogForFile 格式化文件日志（不带颜色）
func formatLogForFile(level LogLevel, timestamp, location, message string, keyvals ...interface{}) string {
	logEntry := fmt.Sprintf("[%s] [%s] %s %s", timestamp, string(level), location, message)

	// 处理键值对参数
	if len(keyvals) > 0 {
		var pairs []string
		for i := 0; i < len(keyvals); i += 2 {
			if i+1 < len(keyvals) {
				key := fmt.Sprintf("%v", keyvals[i])
				value := fmt.Sprintf("%v", keyvals[i+1])
				pairs = append(pairs, fmt.Sprintf("%s=%s", key, value))
			} else {
				// 如果是奇数个参数，最后一个作为单独的值
				pairs = append(pairs, fmt.Sprintf("extra=%v", keyvals[i]))
			}
		}
		if len(pairs) > 0 {
			logEntry += " | " + strings.Join(pairs, " ")
		}
	}

	return logEntry
}

// Log 主要的日志函数，支持日志级别、消息和键值对参数
// 用法: Log(INFO, "用户登录", "user_id", 123, "ip", "192.168.1.1")
func Log(level LogLevel, message string, keyvals ...interface{}) {
	if level == DEBUG && config.GetMode() == "release" {
		return
	}
	defaultLogger.logWithCaller(level, message, 3, keyvals...)
}

// logWithCaller 内部日志实现，支持指定调用者层级
func (l *Logger) logWithCaller(level LogLevel, message string, callerSkip int, keyvals ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 获取调用者信息（文件名和行号）
	_, file, line, ok := runtime.Caller(callerSkip)
	var location string
	if ok {
		// 只保留文件名，不包含完整路径
		filename := filepath.Base(file)
		location = fmt.Sprintf("[%s:%d]", filename, line)
	} else {
		location = "[unknown:0]"
	}

	// 检查输出目标是否为 FileWriter
	if fw, ok := l.output.(*FileWriter); ok {
		// 如果是 FileWriter，使用自定义的写入方法
		fw.WriteLog(level, timestamp, location, message, keyvals...)
	} else {
		// 如果是其他输出（如控制台），使用带颜色的格式
		logEntry := formatLogForConsole(level, timestamp, location, message, keyvals...)
		fmt.Fprintln(l.output, logEntry)
	}
}

// 便捷方法
func Debug(message string, keyvals ...interface{}) {
	Log(DEBUG, message, keyvals...)
}

func Info(message string, keyvals ...interface{}) {
	Log(INFO, message, keyvals...)
}

func Warn(message string, keyvals ...interface{}) {
	Log(WARN, message, keyvals...)
}

func Error(message string, keyvals ...interface{}) {
	Log(ERROR, message, keyvals...)
}

// SetOutput 设置日志输出目标
func SetOutput(writer io.Writer) {
	defaultLogger.output = writer
}
