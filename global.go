package go_logs

import (
	"sync"
)

var (
	// defaultLogger is the global logger instance
	defaultLogger Logger

	// defaultLoggerMu protects the defaultLogger
	defaultLoggerMu sync.RWMutex
)

// Default returns the global default logger.
// If no default logger is set, returns a logger that writes to stdout.
func Default() Logger {
	defaultLoggerMu.RLock()
	defer defaultLoggerMu.RUnlock()

	if defaultLogger != nil {
		return defaultLogger
	}

	// Return a basic stdout logger as fallback
	l, _ := NewLogger(WithLevel(InfoLevel))
	return l
}

// SetDefault sets the global default logger.
// This is typically called once at application startup.
func SetDefault(l Logger) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()
	defaultLogger = l
}

// SetDefaultLevel sets the level of the global default logger.
// If no default logger is set, creates one with the specified level.
func SetDefaultLevel(level Level) {
	defaultLoggerMu.Lock()
	defer defaultLoggerMu.Unlock()

	if defaultLogger != nil {
		defaultLogger.SetLevel(level)
		return
	}

	// Create new default logger with specified level
	l, _ := NewLogger(WithLevel(level))
	defaultLogger = l
}

// Global convenience functions using the default logger

// GlobalInfo logs at Info level using the default logger
func GlobalInfo(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

// GlobalDebug logs at Debug level using the default logger
func GlobalDebug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

// GlobalWarn logs at Warn level using the default logger
func GlobalWarn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

// GlobalError logs at Error level using the default logger
func GlobalError(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}

// GlobalTrace logs at Trace level using the default logger
func GlobalTrace(msg string, fields ...Field) {
	Default().Trace(msg, fields...)
}

// GlobalFatal logs at Fatal level using the default logger and exits
func GlobalFatal(msg string, fields ...Field) {
	Default().Fatal(msg, fields...)
}
