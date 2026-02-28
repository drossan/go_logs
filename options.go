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
