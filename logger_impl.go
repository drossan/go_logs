package go_logs

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// LoggerImpl is the internal implementation of the Logger interface.
// This type is exported to allow the Option pattern to work, but should not
// be used directly by application code.
type LoggerImpl struct {
	// mu protects level changes during concurrent operations
	mu sync.RWMutex

	// level is the minimum log level threshold
	level Level

	// output is where log entries are written
	output io.Writer

	// formatter converts entries to bytes (will be used in Phase 2)
	formatter Formatter

	// hooks are extension points for custom processing (will be used in Phase 4)
	hooks []Hook

	// redactor masks sensitive data (will be used in Phase 6)
	redactor *Redactor

	// parent enables child logger support via With()
	parent *LoggerImpl

	// fields are inherited by child loggers
	fields []Field

	// flags control output formatting (will be used in Phase 2)
	flags int
}

// NewLogger creates a new logger implementation with the given options.
// This is called by go_logs.New() in the public API.
func NewLogger(opts ...Option) (Logger, error) {
	// Create logger with defaults
	l := &LoggerImpl{
		level:     InfoLevel,
		output:    os.Stdout,
		formatter: loadLogFormat(), // Load from LOG_FORMAT env var
		hooks:     []Hook{},
		fields:    []Field{},
		flags:     0,
	}

	// Apply options
	for _, opt := range opts {
		switch o := opt.(type) {
		case *LevelOption:
			l.mu.Lock()
			l.level = o.Level
			l.mu.Unlock()
		case *OutputOption:
			l.mu.Lock()
			l.output = o.Output
			l.mu.Unlock()
		case *FormatterOption:
			l.mu.Lock()
			l.formatter = o.Formatter
			l.mu.Unlock()
		case *HooksOption:
			l.mu.Lock()
			l.hooks = append(l.hooks, o.Hooks...)
			l.mu.Unlock()
		case *RedactorOption:
			l.mu.Lock()
			l.redactor = NewRedactor(o.Keys...)
			l.mu.Unlock()
		case *FlagsOption:
			l.mu.Lock()
			l.flags = o.Flags
			l.mu.Unlock()
		}
	}

	return l, nil
}

// Log implements Logger.Log
func (l *LoggerImpl) Log(level Level, msg string, fields ...Field) {
	// Fast-path: Check level BEFORE any allocations or locking
	// This is the critical performance optimization
	if !level.shouldLog(l.getLevel()) {
		return
	}

	// Create entry
	entry := &Entry{
		Level:     level,
		Message:   msg,
		Fields:    l.combineFields(fields),
		Timestamp: time.Now(),
	}

	// Apply redactor if configured (Phase 6)
	if l.redactor != nil {
		l.redactor.Redact(entry)
	}

	// Execute hooks if configured (Phase 4)
	for _, hook := range l.getHooks() {
		if err := hook.Run(entry); err != nil {
			// Log hook errors to stderr but don't fail the log entry
			fmt.Fprintf(os.Stderr, "Hook error: %v\n", err)
		}
	}

	// Format and write entry
	l.writeEntry(entry)
}

// LogCtx implements Logger.LogCtx
func (l *LoggerImpl) LogCtx(ctx context.Context, level Level, msg string, fields ...Field) {
	// Fast-path filtering
	if !level.shouldLog(l.getLevel()) {
		return
	}

	// Extract fields from context (Phase 3: Real Context)
	// For now, this is a no-op. In Phase 3, we'll extract trace_id, span_id, etc.
	contextFields := l.extractContextFields(ctx)

	// Combine context fields + explicit fields
	allFields := append(contextFields, fields...)

	// Log with combined fields
	l.Log(level, msg, allFields...)
}

// Trace implements Logger.Trace
func (l *LoggerImpl) Trace(msg string, fields ...Field) {
	l.Log(TraceLevel, msg, fields...)
}

// Debug implements Logger.Debug
func (l *LoggerImpl) Debug(msg string, fields ...Field) {
	l.Log(DebugLevel, msg, fields...)
}

// Info implements Logger.Info
func (l *LoggerImpl) Info(msg string, fields ...Field) {
	l.Log(InfoLevel, msg, fields...)
}

// Warn implements Logger.Warn
func (l *LoggerImpl) Warn(msg string, fields ...Field) {
	l.Log(WarnLevel, msg, fields...)
}

// Error implements Logger.Error
func (l *LoggerImpl) Error(msg string, fields ...Field) {
	l.Log(ErrorLevel, msg, fields...)
}

// Fatal implements Logger.Fatal
func (l *LoggerImpl) Fatal(msg string, fields ...Field) {
	l.Log(FatalLevel, msg, fields...)
	// Terminate application after fatal log
	os.Exit(1)
}

// With implements Logger.With
func (l *LoggerImpl) With(fields ...Field) Logger {
	// Create child logger with parent's fields + new fields
	childFields := append(l.fields, fields...)

	return &LoggerImpl{
		level:     l.level,
		output:    l.output,
		formatter: l.formatter,
		hooks:     l.hooks,
		redactor:  l.redactor,
		parent:    l,
		fields:    childFields,
		flags:     l.flags,
	}
}

// SetLevel implements Logger.SetLevel
func (l *LoggerImpl) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel implements Logger.GetLevel
func (l *LoggerImpl) GetLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// Sync implements Logger.Sync
func (l *LoggerImpl) Sync() error {
	// For now, this is a no-op since we're not buffering
	// In Phase 5, this will flush the RotatingFileWriter
	return nil
}

// Helper methods

// getLevel returns the current level with read lock
func (l *LoggerImpl) getLevel() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// getHooks returns a copy of hooks to avoid race conditions
func (l *LoggerImpl) getHooks() []Hook {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.hooks
}

// combineFields merges inherited fields with new fields
func (l *LoggerImpl) combineFields(fields []Field) []Field {
	if len(l.fields) == 0 {
		return fields
	}

	// Create a new slice with parent fields + new fields
	combined := make([]Field, 0, len(l.fields)+len(fields))
	combined = append(combined, l.fields...)
	combined = append(combined, fields...)
	return combined
}

// extractContextFields extracts fields from context for logging
func (l *LoggerImpl) extractContextFields(ctx context.Context) []Field {
	// Extract trace_id, span_id, request_id, user_id from context
	// These fields are automatically included in all LogCtx calls
	return ExtractFieldsFromContext(ctx)
}

// writeEntry formats and writes the log entry using the configured formatter
func (l *LoggerImpl) writeEntry(entry *Entry) {
	// Get formatter (use TextFormatter as default if none configured)
	formatter := l.getFormatter()
	if formatter == nil {
		// Fallback to TextFormatter if none configured
		formatter = NewTextFormatter()
	}

	// Format the entry
	formatted, err := formatter.Format(entry)
	if err != nil {
		// If formatting fails, write error to stderr and use simple string format
		fmt.Fprintf(os.Stderr, "Format error: %v\n", err)
		formatted = []byte(entry.String() + "\n")
	}

	// Write to output
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output.Write(formatted)
}

// getFormatter returns the current formatter with read lock
func (l *LoggerImpl) getFormatter() Formatter {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.formatter
}
