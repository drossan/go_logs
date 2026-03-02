// Package async provides asynchronous logging capabilities for go_logs.
//
// Async logging writes log entries in a separate goroutine, providing non-blocking
// logging for high-throughput applications. This is especially useful when logging
// to slow outputs like remote services or network filesystems.
//
// Features:
//   - Non-blocking log writes
//   - Buffered channel for high throughput
//   - Graceful shutdown with timeout
//   - Drop-on-overflow behavior (configurable)
//   - Shared metrics with underlying logger
//
// Example:
//
//	syncLogger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
//	asyncLogger := async.Wrap(syncLogger, 1000) // buffer size 1000
//	defer asyncLogger.Sync()
//
//	// Non-blocking logging
//	asyncLogger.Info("Server started", go_logs.Int("port", 8080))
package async

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	go_logs "github.com/drossan/go_logs"
)

// Config holds configuration for the async logger.
type Config struct {
	// BufferSize is the number of log entries to buffer.
	// When full, new entries are dropped.
	BufferSize int

	// ShutdownTimeout is the maximum time to wait for pending logs during shutdown.
	// If timeout is reached, remaining logs are dropped.
	ShutdownTimeout time.Duration
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		BufferSize:      1000,
		ShutdownTimeout: 5 * time.Second,
	}
}

// Logger wraps a sync logger with async capabilities.
// It implements the go_logs.Logger interface.
type Logger struct {
	// syncLogger is the underlying synchronous logger
	syncLogger go_logs.Logger

	// buffer is the channel for log entries
	buffer chan *logEntry

	// pending tracks number of pending log operations
	pending atomic.Int64

	// mu protects shutdown state
	mu sync.RWMutex

	// shutdown indicates the logger is shutting down
	shutdown atomic.Bool

	// config holds the configuration
	config Config

	// done channel signals the worker to stop
	done chan struct{}

	// fields for child loggers
	fields []go_logs.Field
}

// logEntry represents a log entry in the buffer
type logEntry struct {
	level  go_logs.Level
	msg    string
	ctx    context.Context
	fields []go_logs.Field
	// loggerFields are the fields from the async logger (for child loggers)
	loggerFields []go_logs.Field
}

// Wrap creates an async logger wrapper around a sync logger.
// The buffer size determines how many log entries can be queued before dropping.
func Wrap(syncLogger go_logs.Logger, bufferSize int) *Logger {
	return WrapWithConfig(syncLogger, Config{
		BufferSize:      bufferSize,
		ShutdownTimeout: 5 * time.Second,
	})
}

// WrapWithConfig creates an async logger wrapper with custom configuration.
func WrapWithConfig(syncLogger go_logs.Logger, config Config) *Logger {
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}
	if config.ShutdownTimeout <= 0 {
		config.ShutdownTimeout = 5 * time.Second
	}

	l := &Logger{
		syncLogger: syncLogger,
		buffer:     make(chan *logEntry, config.BufferSize),
		config:     config,
		done:       make(chan struct{}),
		fields:     nil,
	}

	// Start the worker goroutine
	go l.worker()

	return l
}

// worker processes log entries from the buffer
func (l *Logger) worker() {
	for {
		select {
		case entry := <-l.buffer:
			l.processEntry(entry)
			l.pending.Add(-1)
		case <-l.done:
			// Drain remaining entries
			for {
				select {
				case entry := <-l.buffer:
					l.processEntry(entry)
					l.pending.Add(-1)
				default:
					return
				}
			}
		}
	}
}

// processEntry processes a single log entry
func (l *Logger) processEntry(entry *logEntry) {
	// Combine logger's fields with entry's fields
	allFields := entry.fields
	if len(entry.loggerFields) > 0 {
		combined := make([]go_logs.Field, 0, len(entry.loggerFields)+len(entry.fields))
		combined = append(combined, entry.loggerFields...)
		combined = append(combined, entry.fields...)
		allFields = combined
	}

	if entry.ctx != nil {
		l.syncLogger.LogCtx(entry.ctx, entry.level, entry.msg, allFields...)
	} else {
		l.syncLogger.Log(entry.level, entry.msg, allFields...)
	}
}

// Log implements go_logs.Logger.Log
func (l *Logger) Log(level go_logs.Level, msg string, fields ...go_logs.Field) {
	l.enqueue(&logEntry{level: level, msg: msg, fields: fields, loggerFields: l.fields})
}

// LogCtx implements go_logs.Logger.LogCtx
func (l *Logger) LogCtx(ctx context.Context, level go_logs.Level, msg string, fields ...go_logs.Field) {
	l.enqueue(&logEntry{level: level, msg: msg, ctx: ctx, fields: fields, loggerFields: l.fields})
}

// enqueue adds a log entry to the buffer
func (l *Logger) enqueue(entry *logEntry) {
	if l.shutdown.Load() {
		return
	}

	// Check level first (fast-path)
	if !entry.level.ShouldLog(l.syncLogger.GetLevel()) {
		return
	}

	l.pending.Add(1)

	select {
	case l.buffer <- entry:
		// Successfully enqueued
	default:
		// Buffer full, drop the entry
		l.pending.Add(-1)
		// Increment dropped counter in metrics
		if mg, ok := l.syncLogger.(interface{ GetMetrics() *go_logs.Metrics }); ok {
			mg.GetMetrics().IncrementDropped()
		}
	}
}

// Trace implements go_logs.Logger.Trace
func (l *Logger) Trace(msg string, fields ...go_logs.Field) {
	l.Log(go_logs.TraceLevel, msg, fields...)
}

// Debug implements go_logs.Logger.Debug
func (l *Logger) Debug(msg string, fields ...go_logs.Field) {
	l.Log(go_logs.DebugLevel, msg, fields...)
}

// Info implements go_logs.Logger.Info
func (l *Logger) Info(msg string, fields ...go_logs.Field) {
	l.Log(go_logs.InfoLevel, msg, fields...)
}

// Warn implements go_logs.Logger.Warn
func (l *Logger) Warn(msg string, fields ...go_logs.Field) {
	l.Log(go_logs.WarnLevel, msg, fields...)
}

// Error implements go_logs.Logger.Error
func (l *Logger) Error(msg string, fields ...go_logs.Field) {
	l.Log(go_logs.ErrorLevel, msg, fields...)
}

// Fatal implements go_logs.Logger.Fatal
func (l *Logger) Fatal(msg string, fields ...go_logs.Field) {
	// Fatal must be synchronous to ensure message is written before exit
	l.syncLogger.Fatal(msg, fields...)
}

// With implements go_logs.Logger.With
func (l *Logger) With(fields ...go_logs.Field) go_logs.Logger {
	// Combine existing fields with new fields
	combined := make([]go_logs.Field, 0, len(l.fields)+len(fields))
	combined = append(combined, l.fields...)
	combined = append(combined, fields...)

	return &Logger{
		syncLogger: l.syncLogger,
		buffer:     l.buffer,
		config:     l.config,
		done:       l.done,
		fields:     combined,
	}
}

// SetLevel implements go_logs.Logger.SetLevel
func (l *Logger) SetLevel(level go_logs.Level) {
	l.syncLogger.SetLevel(level)
}

// GetLevel implements go_logs.Logger.GetLevel
func (l *Logger) GetLevel() go_logs.Level {
	return l.syncLogger.GetLevel()
}

// Sync implements go_logs.Logger.Sync
// It waits for all pending log entries to be written.
func (l *Logger) Sync() error {
	// Wait for all pending entries with a reasonable timeout
	timeout := time.NewTimer(l.config.ShutdownTimeout)
	defer timeout.Stop()

	for {
		if l.pending.Load() == 0 {
			return nil
		}
		select {
		case <-timeout.C:
			return nil // Timeout, but don't error
		case <-time.After(1 * time.Millisecond):
			// Check again
		}
	}
}

// Close gracefully shuts down the async logger.
// It flushes pending logs and stops the worker goroutine.
func (l *Logger) Close() error {
	l.shutdown.Store(true)

	// Signal worker to stop and drain
	close(l.done)

	// Wait for pending entries with timeout
	timeout := time.NewTimer(l.config.ShutdownTimeout)
	defer timeout.Stop()

	for {
		if l.pending.Load() == 0 {
			return nil
		}
		select {
		case <-timeout.C:
			return nil
		case <-time.After(1 * time.Millisecond):
			// Check again
		}
	}
}

// GetMetrics returns the metrics from the underlying logger.
func (l *Logger) GetMetrics() *go_logs.Metrics {
	if mg, ok := l.syncLogger.(interface{ GetMetrics() *go_logs.Metrics }); ok {
		return mg.GetMetrics()
	}
	return nil
}
