package logger

import (
	"kong-gateway/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// Init initializes the global logger
func Init(cfg *config.LogConfig) error {
	var err error

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Determine encoding
	encoding := cfg.Format
	if cfg.Format == "console" {
		encoding = "console"
	}

	// Build config
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         encoding,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	Log, err = zapConfig.Build()
	if err != nil {
		return err
	}

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(msg, fields...)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Debug(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(msg, fields...)
	}
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(msg, fields...)
	}
}

// LogField creates a zap field for logging
func LogField(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	return Log
}
