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
//   - Child loggers (With) sharing the parent's pipeline; Close is a no-op on them
//
// Example:
//
//	syncLogger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))
//	asyncLogger := async.Wrap(syncLogger, 1000) // buffer size 1000
//	defer asyncLogger.Close() // drains pending entries and stops the worker
//
//	// Non-blocking logging
//	asyncLogger.Info("Server started", go_logs.Int("port", 8080))
package async

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	go_logs "github.com/drossan/go_logs/v3"
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

// core is the pipeline shared by a root async Logger and every child created
// from it with With(). Sharing it by pointer is what keeps Sync() and Close()
// correct across the logger tree: there is a single buffer, a single worker and a
// single pending counter.
type core struct {
	// syncLogger is the underlying synchronous logger
	syncLogger go_logs.Logger

	// buffer is the channel for log entries
	buffer chan *logEntry

	// pending tracks number of enqueued entries not yet written by the worker
	pending atomic.Int64

	// shutdown indicates the logger is shutting down; new entries are discarded
	shutdown atomic.Bool

	// config holds the configuration
	config Config

	// done signals the worker to drain and stop
	done chan struct{}

	// workerDone is closed by the worker when it returns
	workerDone chan struct{}

	// closeOnce makes Close idempotent
	closeOnce sync.Once
}

// Logger wraps a sync logger with async capabilities.
// It implements the go_logs.Logger interface.
//
// The logger returned by Wrap/WrapWithConfig is the root. Children created with
// With() share the root's buffer, worker and pending counter, so Sync() on any
// of them waits for every entry of the tree, and only differ in their fields.
type Logger struct {
	// core is the pipeline shared with the root and all children
	core *core

	// fields are prepended to every entry logged through this logger
	fields []go_logs.Field

	// isRoot is true only for the logger returned by Wrap/WrapWithConfig
	isRoot bool
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
//
// Wrap starts a worker goroutine. Close() on the returned (root) logger is the
// only way to stop it: Sync() waits for pending entries but leaves the worker
// running, so a logger that is only ever synced keeps its goroutine alive for
// the lifetime of the process.
func Wrap(syncLogger go_logs.Logger, bufferSize int) *Logger {
	return WrapWithConfig(syncLogger, Config{
		BufferSize:      bufferSize,
		ShutdownTimeout: 5 * time.Second,
	})
}

// WrapWithConfig creates an async logger wrapper with custom configuration.
// A non-positive BufferSize defaults to 1000 and a non-positive ShutdownTimeout
// to 5 seconds. As with Wrap, call Close() on the returned logger to stop the
// worker goroutine.
func WrapWithConfig(syncLogger go_logs.Logger, config Config) *Logger {
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}
	if config.ShutdownTimeout <= 0 {
		config.ShutdownTimeout = 5 * time.Second
	}

	c := &core{
		syncLogger: syncLogger,
		buffer:     make(chan *logEntry, config.BufferSize),
		config:     config,
		done:       make(chan struct{}),
		workerDone: make(chan struct{}),
	}

	// Start the worker goroutine
	go c.worker()

	return &Logger{core: c, isRoot: true}
}

// worker processes log entries from the buffer until done is closed, then
// drains what is left and returns.
func (c *core) worker() {
	defer close(c.workerDone)
	for {
		select {
		case entry := <-c.buffer:
			c.processEntry(entry)
		case <-c.done:
			for {
				select {
				case entry := <-c.buffer:
					c.processEntry(entry)
				default:
					return
				}
			}
		}
	}
}

// processEntry writes a single log entry and marks it as no longer pending.
func (c *core) processEntry(entry *logEntry) {
	defer c.pending.Add(-1)

	// Combine logger's fields with entry's fields
	allFields := entry.fields
	if len(entry.loggerFields) > 0 {
		allFields = combineFields(entry.loggerFields, entry.fields)
	}

	if entry.ctx != nil {
		c.syncLogger.LogCtx(entry.ctx, entry.level, entry.msg, allFields...)
	} else {
		c.syncLogger.Log(entry.level, entry.msg, allFields...)
	}
}

// combineFields returns a new slice with base followed by extra; it never
// appends to base, so loggers never share field memory.
func combineFields(base, extra []go_logs.Field) []go_logs.Field {
	combined := make([]go_logs.Field, 0, len(base)+len(extra))
	combined = append(combined, base...)
	return append(combined, extra...)
}

// Log implements go_logs.Logger.Log
func (l *Logger) Log(level go_logs.Level, msg string, fields ...go_logs.Field) {
	l.core.enqueue(&logEntry{level: level, msg: msg, fields: fields, loggerFields: l.fields})
}

// LogCtx implements go_logs.Logger.LogCtx
func (l *Logger) LogCtx(ctx context.Context, level go_logs.Level, msg string, fields ...go_logs.Field) {
	l.core.enqueue(&logEntry{level: level, msg: msg, ctx: ctx, fields: fields, loggerFields: l.fields})
}

// enqueue adds a log entry to the buffer, or drops it (counting it in the
// metrics) when the buffer is full. Entries are discarded after Close.
func (c *core) enqueue(entry *logEntry) {
	if c.shutdown.Load() {
		return
	}

	// Check level first (fast-path)
	if !entry.level.ShouldLog(c.syncLogger.GetLevel()) {
		return
	}

	c.pending.Add(1)

	select {
	case c.buffer <- entry:
		// Successfully enqueued
	default:
		// Buffer full, drop the entry
		c.pending.Add(-1)
		// Increment dropped counter in metrics
		if mg, ok := c.syncLogger.(interface{ GetMetrics() *go_logs.Metrics }); ok {
			mg.GetMetrics().IncrementDropped()
		}
	}
}

// waitFor polls until pending reaches zero, the worker has stopped or
// ShutdownTimeout expires, whichever happens first. It never returns an error.
func (c *core) waitFor() {
	timeout := time.NewTimer(c.config.ShutdownTimeout)
	defer timeout.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		if c.pending.Load() == 0 {
			return
		}
		select {
		case <-c.workerDone:
			// Nothing left will be written: entries that raced with Close are dropped.
			return
		case <-timeout.C:
			return
		case <-ticker.C:
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

// Fatal implements go_logs.Logger.Fatal.
// It first waits (up to ShutdownTimeout) for pending entries to be written, so
// earlier messages are not lost, and then calls Fatal on the underlying sync
// logger with this logger's fields followed by fields.
func (l *Logger) Fatal(msg string, fields ...go_logs.Field) {
	l.Sync()
	l.core.syncLogger.Fatal(msg, combineFields(l.fields, fields)...)
}

// With implements go_logs.Logger.With.
// The child shares the parent's pipeline (buffer, worker, pending counter) and
// gets its own copy of the parent's fields plus fields. Close() on the child is
// a no-op.
func (l *Logger) With(fields ...go_logs.Field) go_logs.Logger {
	return &Logger{
		core:   l.core,
		fields: combineFields(l.fields, fields),
		isRoot: false,
	}
}

// SetLevel implements go_logs.Logger.SetLevel
func (l *Logger) SetLevel(level go_logs.Level) {
	l.core.syncLogger.SetLevel(level)
}

// GetLevel implements go_logs.Logger.GetLevel
func (l *Logger) GetLevel() go_logs.Level {
	return l.core.syncLogger.GetLevel()
}

// Sync implements go_logs.Logger.Sync.
// It waits until every entry enqueued by this logger's tree (root and children)
// has been written, or until ShutdownTimeout expires; the timeout is not
// reported as an error. Sync does not stop the worker goroutine: use Close()
// on the root logger for that.
func (l *Logger) Sync() error {
	l.core.waitFor()
	return nil
}

// Close gracefully shuts down the async logger.
// It stops accepting entries, lets the worker drain the buffer and waits for it
// up to ShutdownTimeout. It always returns nil.
//
// Close only has an effect on the root logger returned by Wrap/WrapWithConfig;
// on a child created with With() it is a no-op. Calling Close on the root more
// than once is safe: later calls just wait again. Entries logged after Close
// are discarded.
func (l *Logger) Close() error {
	if !l.isRoot {
		return nil
	}
	l.core.closeOnce.Do(func() {
		l.core.shutdown.Store(true)
		close(l.core.done)
	})
	l.core.waitFor()
	return nil
}

// GetMetrics returns the metrics from the underlying logger.
func (l *Logger) GetMetrics() *go_logs.Metrics {
	if mg, ok := l.core.syncLogger.(interface{ GetMetrics() *go_logs.Metrics }); ok {
		return mg.GetMetrics()
	}
	return nil
}
