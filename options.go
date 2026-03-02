package go_logs

import (
	"io"
)

// IMPORTANT: The types and functions in this file are placeholders for Phase 1.
// They will be fully implemented in Phase 2 (Output Layer) and later phases.

// Option defines a functional option for configuring a Logger.
type Option interface{}

// LevelOption is the actual implementation for WithLevel option
type LevelOption struct {
	Level Level
}

// WithLevel sets the minimum log level threshold.
func WithLevel(level Level) Option {
	return &LevelOption{Level: level}
}

// OutputOption is the actual implementation for WithOutput option
type OutputOption struct {
	Output io.Writer
}

// WithOutput sets the output destination for log entries.
func WithOutput(w io.Writer) Option {
	return &OutputOption{Output: w}
}

// FormatterOption is the actual implementation for WithFormatter option
type FormatterOption struct {
	Formatter Formatter
}

// WithFormatter sets the formatter for log entries.
func WithFormatter(f Formatter) Option {
	return &FormatterOption{Formatter: f}
}

// HooksOption is the actual implementation for WithHooks option
type HooksOption struct {
	Hooks []Hook
}

// WithHooks adds hooks to the logger.
func WithHooks(hooks ...Hook) Option {
	return &HooksOption{Hooks: hooks}
}

// RedactorOption is the actual implementation for WithRedactor option
type RedactorOption struct {
	Keys []string
}

// WithRedactor enables redaction of sensitive fields.
func WithRedactor(keys ...string) Option {
	return &RedactorOption{Keys: keys}
}

// FlagsOption is the actual implementation for WithOutputFlags option
type FlagsOption struct {
	Flags int
}

// WithOutputFlags configures output formatting flags.
func WithOutputFlags(flags int) Option {
	return &FlagsOption{Flags: flags}
}

// Placeholder interfaces that will be implemented in later phases:

// Hook defines an extension point for processing log entries.
// Will be implemented in Phase 4 (Extensibility).
type Hook interface {
	Run(entry *Entry) error
}

// Redactor masks sensitive data in log entries.
// Will be implemented in Phase 6 (Security).
type Redactor struct {
	sensitiveKeys map[string]bool
	maskValue     string
}

// NewRedactor creates a new redactor with the given sensitive keys.
// Will be implemented in Phase 6 (Security).
func NewRedactor(keys ...string) *Redactor {
	sensitiveKeys := make(map[string]bool)
	for _, key := range keys {
		sensitiveKeys[key] = true
	}
	return &Redactor{
		sensitiveKeys: sensitiveKeys,
		maskValue:     "***",
	}
}

// Redact modifies the entry by masking sensitive field values.
// Will be implemented in Phase 6 (Security).
func (r *Redactor) Redact(entry *Entry) {
	for i := range entry.Fields {
		if r.sensitiveKeys[entry.Fields[i].Key()] {
			entry.Fields[i] = Field{
				key:       entry.Fields[i].Key(),
				valueType: entry.Fields[i].Type(),
				value:     r.maskValue,
			}
		}
	}
}

// CommonSensitiveKeys returns a list of commonly sensitive field names.
func CommonSensitiveKeys() []string {
	return []string{
		"password", "passwd", "pwd",
		"token", "api_key", "apikey", "api-key",
		"secret", "authorization", "auth",
		"cookie", "session",
		"credit_card", "ssn", "social_security",
	}
}

// WithCommonRedaction enables redaction of common sensitive fields.
func WithCommonRedaction() Option {
	return WithRedactor(CommonSensitiveKeys()...)
}

// CallerOption is the actual implementation for WithCaller option
type CallerOption struct {
	Enabled bool
}

// WithCaller enables or disables caller information in log entries.
// When enabled, each log entry includes file name, line number, and function name.
//
// Example output with caller enabled:
//
//	[2026/02/28 17:30:00] INFO main.go:42 myFunction connection established
//
// Performance note: Enabling caller adds ~100-200ns per log call.
func WithCaller(enabled bool) Option {
	return &CallerOption{Enabled: enabled}
}

// StackTraceOption is the actual implementation for WithStackTrace option
type StackTraceOption struct {
	Enabled bool
}

// WithStackTrace enables or disables stack trace capture in log entries.
// Stack traces are captured for Error level and above by default.
//
// Use WithStackTraceLevel to customize the minimum level for stack traces.
func WithStackTrace(enabled bool) Option {
	return &StackTraceOption{Enabled: enabled}
}

// StackTraceLevelOption is the actual implementation for WithStackTraceLevel option
type StackTraceLevelOption struct {
	Level Level
}

// WithStackTraceLevel sets the minimum level for automatic stack trace capture.
// Stack traces will only be captured for log entries at or above this level.
//
// Default: ErrorLevel (stack traces for Error and Fatal)
// Common values:
//   - WarnLevel: Capture for Warn, Error, Fatal
//   - ErrorLevel: Capture for Error, Fatal (default)
//   - FatalLevel: Capture only for Fatal
func WithStackTraceLevel(level Level) Option {
	return &StackTraceLevelOption{Level: level}
}

// CallerSkipOption is the actual implementation for WithCallerSkip option
type CallerSkipOption struct {
	Skip int
}

// WithCallerSkip sets the number of stack frames to skip when capturing caller info.
// This is useful when wrapping the logger with additional helper functions.
//
// Default: 2 (skips GetCaller and Log methods)
// Increase this value if you have additional wrapper functions.
func WithCallerSkip(skip int) Option {
	return &CallerSkipOption{Skip: skip}
}

// CallerLevelOption is the actual implementation for WithCallerLevel option
type CallerLevelOption struct {
	Level Level
}

// WithCallerLevel sets the minimum level for automatic caller info capture.
// Caller info (file:line function) will be captured for log entries at or above this level.
//
// Default: ErrorLevel (caller info for Error and Fatal)
// Common values:
//   - WarnLevel: Capture for Warn, Error, Fatal
//   - ErrorLevel: Capture for Error, Fatal (default)
//   - FatalLevel: Capture only for Fatal
//   - SilentLevel: Disable automatic caller capture (use WithCaller(true) for all levels)
//
// Note: WithCaller(true) enables caller for ALL levels, overriding this setting.
func WithCallerLevel(level Level) Option {
	return &CallerLevelOption{Level: level}
}

// MultiOutputOption is the actual implementation for WithMultiOutput option
type MultiOutputOption struct {
	Writers []io.Writer
}

// WithMultiOutput enables output to multiple writers simultaneously.
// This is useful for logging to both file and console.
//
// Example:
//
//	file, _ := go_logs.NewRotatingFileWriter("app.log", 100, 5)
//	logger, _ := go_logs.New(
//	    go_logs.WithMultiOutput(file, os.Stdout),
//	)
func WithMultiOutput(writers ...io.Writer) Option {
	return &MultiOutputOption{Writers: writers}
}

// RotatingFileEnhancedOption is the actual implementation for WithRotatingFileEnhanced option
type RotatingFileEnhancedOption struct {
	Config RotatingFileConfig
}

// WithRotatingFileEnhanced creates a rotating file writer with full configuration.
// This supports time-based rotation and compression.
//
// Example:
//
//	logger, _ := go_logs.New(
//	    go_logs.WithRotatingFileEnhanced(go_logs.RotatingFileConfig{
//	        Filename:     "/var/log/app.log",
//	        MaxSizeMB:    100,
//	        MaxBackups:   5,
//	        RotationType: go_logs.RotateDaily,
//	        Compress:     true,
//	        MaxAge:       30, // Keep 30 days
//	    }),
//	)
func WithRotatingFileEnhanced(config RotatingFileConfig) Option {
	return &RotatingFileEnhancedOption{Config: config}
}

// RotatingFileOption is the actual implementation for WithRotatingFile option
type RotatingFileOption struct {
	Filename   string
	MaxSizeMB  int
	MaxBackups int
}

// WithRotatingFile creates a simple rotating file writer.
// This is a convenience wrapper around WithRotatingFileEnhanced with sensible defaults.
//
// Parameters:
//   - filename: Path to the log file
//   - maxSizeMB: Maximum size in megabytes before rotation
//   - maxBackups: Maximum number of backup files to keep
//
// Example:
//
//	logger, _ := go_logs.New(
//	    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
//	)
func WithRotatingFile(filename string, maxSizeMB int, maxBackups int) Option {
	return &RotatingFileOption{
		Filename:   filename,
		MaxSizeMB:  maxSizeMB,
		MaxBackups: maxBackups,
	}
}
