package logger

import "go.uber.org/zap"

// Logger interface to abstract logging
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	Named(name string) Logger
}
