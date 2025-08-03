package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"crypto-info/internal/config"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志接口
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// logrusLogger logrus实现
type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewLogger 创建新的日志实例
func NewLogger(cfg *config.Log) Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// 设置输出
	var output io.Writer
	switch cfg.Output {
	case "file":
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			logger.Errorf("Failed to create log directory: %v", err)
			output = os.Stdout
		} else {
			output = &lumberjack.Logger{
				Filename:   cfg.FilePath,
				MaxSize:    cfg.MaxSize,
				MaxAge:     cfg.MaxAge,
				MaxBackups: cfg.MaxBackups,
				Compress:   cfg.Compress,
			}
		}
	case "stderr":
		output = os.Stderr
	default:
		output = os.Stdout
	}

	logger.SetOutput(output)

	return &logrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
}

// Debug 调试日志
func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

// Debugf 格式化调试日志
func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

// Info 信息日志
func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

// Infof 格式化信息日志
func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

// Warn 警告日志
func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

// Warnf 格式化警告日志
func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

// Error 错误日志
func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

// Errorf 格式化错误日志
func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Fatal 致命错误日志
func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Fatalf 格式化致命错误日志
func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// WithField 添加字段
func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField(key, value),
	}
}

// WithFields 添加多个字段
func (l *logrusLogger) WithFields(fields map[string]interface{}) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(fields),
	}
}

// 全局日志实例
var defaultLogger Logger

// Init 初始化全局日志
func Init(cfg *config.Log) {
	defaultLogger = NewLogger(cfg)
}

// GetLogger 获取全局日志实例
func GetLogger() Logger {
	if defaultLogger == nil {
		// 如果没有初始化，使用默认配置
		defaultLogger = NewLogger(&config.Log{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		})
	}
	return defaultLogger
}

// GetLogrusLogger 获取底层的logrus.Logger实例
func GetLogrusLogger() *logrus.Logger {
	logger := GetLogger()
	if logrusImpl, ok := logger.(*logrusLogger); ok {
		return logrusImpl.logger
	}
	// 如果转换失败，返回一个新的logrus实例
	return logrus.New()
}

// 便捷方法
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

func WithField(key string, value interface{}) Logger {
	return GetLogger().WithField(key, value)
}

func WithFields(fields map[string]interface{}) Logger {
	return GetLogger().WithFields(fields)
}