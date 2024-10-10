package logger

import (
	"sync"

	"go.uber.org/zap"
)

var (
	instance Logger    // Global instance of Logger interface
	once     sync.Once // Ensures logger is initialized only once
)

// Init initializes the global logger. This function should be called once, ideally during the app startup.
func Init(zapLogger *zap.Logger) {
	once.Do(func() {
		instance = NewZapLogger(zapLogger) // Assign the zapLogger as the global logger
	})
}

// Get returns the global logger instance.
func Get() Logger {
	if instance == nil {
		panic("Logger is not initialized. Call logger.Init() before using the logger.")
	}
	return instance
}
