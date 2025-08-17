package logger

import (
	"fmt"
	"log"
	"os"

	"go-cloud-disk/conf"
)

const (
	LevelError         = iota // 错误级别
	LevelWarning              // 警告级别
	LevelInformational        // 信息级别
	LevelDebug                // 调试级别
)

// Logger 日志记录器
type Logger struct {
	level int // 日志级别
}

var logger *Logger

// Println 打印带时间戳的日志消息
func (ll *Logger) Println(msg string) {
	log.Println(msg)
}

// Panic 打印严重错误并退出程序
func (ll *Logger) Panic(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	msg := fmt.Sprintf("[Panic] "+format, v...)
	ll.Println(msg)
	os.Exit(0)
}

// Error 打印错误信息
func (ll *Logger) Error(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	msg := fmt.Sprintf("[Error] "+format, v...)
	ll.Println(msg)
}

// Warning 打印警告信息
func (ll *Logger) Warning(format string, v ...interface{}) {
	if LevelWarning > ll.level {
		return
	}
	msg := fmt.Sprintf("[Warning] "+format, v...)
	ll.Println(msg)
}

// Info 打印提示信息
func (ll *Logger) Info(format string, v ...interface{}) {
	if LevelInformational > ll.level {
		return
	}
	msg := fmt.Sprintf("[Info] "+format, v...)
	ll.Println(msg)
}

// Debug 打印调试信息
func (ll *Logger) Debug(format string, v ...interface{}) {
	if LevelDebug > ll.level {
		return
	}
	msg := fmt.Sprintf("[Debug] "+format, v...)
	ll.Println(msg)
}

// BuildLogger 根据级别构建日志记录器
func BuildLogger() {
	level := conf.LogLevel
	intLevel := LevelError
	switch level {
	case "error":
		intLevel = LevelError
	case "warning":
		intLevel = LevelWarning
	case "info":
		intLevel = LevelInformational
	case "debug":
		intLevel = LevelDebug
	}
	l := Logger{
		level: intLevel,
	}
	logger = &l
}

// Log 返回日志记录器实例
func Log() *Logger {
	if logger == nil {
		l := Logger{
			level: LevelDebug,
		}
		logger = &l
	}
	return logger
}
