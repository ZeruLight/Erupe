package logger

import (
	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new ZapLogger instance
func NewZapLogger(zapLogger *zap.Logger) *ZapLogger {
	return &ZapLogger{logger: zapLogger}
}

// Implement the Logger interface methods
func (z *ZapLogger) Info(msg string, fields ...zap.Field) {
	z.logger.Info(msg, fields...)
}

func (z *ZapLogger) Warn(msg string, fields ...zap.Field) {
	z.logger.Warn(msg, fields...)
}

func (z *ZapLogger) Error(msg string, fields ...zap.Field) {
	z.logger.Error(msg, fields...)
}
func (z *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	z.logger.Fatal(msg, fields...)
}

func (z *ZapLogger) Debug(msg string, fields ...zap.Field) {
	z.logger.Debug(msg, fields...)
}
func (z *ZapLogger) Named(name string) Logger {
	return &ZapLogger{logger: z.logger.Named(name)}
}
