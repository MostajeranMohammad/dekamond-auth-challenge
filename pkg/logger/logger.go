package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
}

type zapLogger struct {
	sugared *zap.SugaredLogger
}

func New(logLevel string) (Logger, *zap.Logger) {
	var level zapcore.Level

	switch strings.ToLower(logLevel) {
	case "error":
		level = zapcore.ErrorLevel
	case "warn":
		level = zapcore.WarnLevel
	case "info":
		level = zapcore.InfoLevel
	case "debug":
		level = zapcore.DebugLevel
	default:
		level = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	baseLogger, _ := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(3), // Matches original skip frame count
	)

	return &zapLogger{
		sugared: baseLogger.Sugar(),
	}, baseLogger
}

func convertMessage(level string, message interface{}) string {
	switch msg := message.(type) {
	case error:
		return msg.Error()
	case string:
		return msg
	default:
		return fmt.Sprintf("%s message - %v (type: %T)", level, message, message)
	}
}

func (l *zapLogger) Debug(message interface{}, args ...interface{}) {
	msgStr := convertMessage("debug", message)
	if len(args) == 0 {
		l.sugared.Debug(msgStr)
	} else {
		l.sugared.Debugf(msgStr, args...)
	}
}

func (l *zapLogger) Info(message string, args ...interface{}) {
	if len(args) == 0 {
		l.sugared.Info(message)
	} else {
		l.sugared.Infof(message, args...)
	}
}

func (l *zapLogger) Warn(message string, args ...interface{}) {
	if len(args) == 0 {
		l.sugared.Warn(message)
	} else {
		l.sugared.Warnf(message, args...)
	}
}

func (l *zapLogger) Error(message interface{}, args ...interface{}) {
	msgStr := convertMessage("error", message)
	if len(args) == 0 {
		l.sugared.Error(msgStr)
	} else {
		l.sugared.Errorf(msgStr, args...)
	}
}

func (l *zapLogger) Fatal(message interface{}, args ...interface{}) {
	msgStr := convertMessage("fatal", message)
	if len(args) == 0 {
		l.sugared.Fatal(msgStr)
	} else {
		l.sugared.Fatalf(msgStr, args...)
	}
}
