package go_logs

import (
	"fmt"
	"strings"
	"time"
)

// Entry represents a complete log entry with all its components.
// It is created by the logger and passed to formatters and hooks.
type Entry struct {
	// Level is the severity level of this log entry
	Level Level

	// Message is the main log message
	Message string

	// Fields contains structured key-value pairs for this entry
	Fields []Field

	// Timestamp is when this log entry was created
	Timestamp time.Time

	// Caller contains information about the calling function (optional)
	// Only populated when WithCaller(true) is configured
	Caller *CallerInfo

	// StackTrace contains the stack trace (optional)
	// Only populated when WithStackTrace(true) is configured and level meets threshold
	StackTrace []byte
}

// String returns a simple string representation of the entry.
// This is useful for debugging and simple formatters.
func (e *Entry) String() string {
	var builder strings.Builder

	// Timestamp
	if !e.Timestamp.IsZero() {
		builder.WriteString(e.Timestamp.Format("2006/01/02 15:04:05"))
		builder.WriteString(" ")
	}

	// Level
	builder.WriteString("[")
	builder.WriteString(e.Level.String())
	builder.WriteString("] ")

	// Message
	builder.WriteString(e.Message)

	// Fields
	if len(e.Fields) > 0 {
		builder.WriteString(" ")
		for i, field := range e.Fields {
			if i > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(fmt.Sprintf("%s=%v", field.Key(), field.Value()))
		}
	}

	return builder.String()
}

// addFields appends fields to the entry.
// This is used internally by the logger to add context fields.
func (e *Entry) addFields(fields ...Field) {
	e.Fields = append(e.Fields, fields...)
}

// HasField checks if the entry contains a field with the given key.
func (e *Entry) HasField(key string) bool {
	for _, field := range e.Fields {
		if field.Key() == key {
			return true
		}
	}
	return false
}

// GetField returns the field with the given key, or a zero-value Field if not found.
func (e *Entry) GetField(key string) Field {
	for _, field := range e.Fields {
		if field.Key() == key {
			return field
		}
	}
	return Field{}
}

// GetFieldValue returns the value of the field with the given key.
// Returns nil if the field is not found.
func (e *Entry) GetFieldValue(key string) interface{} {
	for _, field := range e.Fields {
		if field.Key() == key {
			return field.Value()
		}
	}
	return nil
}

// Clone creates a deep copy of the entry.
// This is useful for hooks that need to modify the entry without affecting the original.
func (e *Entry) Clone() *Entry {
	cloned := &Entry{
		Level:     e.Level,
		Message:   e.Message,
		Timestamp: e.Timestamp,
	}

	if len(e.Fields) > 0 {
		cloned.Fields = make([]Field, len(e.Fields))
		copy(cloned.Fields, e.Fields)
	}

	// Copy caller info
	if e.Caller != nil {
		cloned.Caller = &CallerInfo{
			File:     e.Caller.File,
			FileFull: e.Caller.FileFull,
			Line:     e.Caller.Line,
			Func:     e.Caller.Func,
			Package:  e.Caller.Package,
		}
	}

	// Copy stack trace
	if e.StackTrace != nil {
		cloned.StackTrace = make([]byte, len(e.StackTrace))
		copy(cloned.StackTrace, e.StackTrace)
	}

	return cloned
}

// WithFields returns a new entry with additional fields appended.
// The original entry is not modified.
func (e *Entry) WithFields(fields ...Field) *Entry {
	cloned := e.Clone()
	cloned.addFields(fields...)
	return cloned
}

// FieldCount returns the number of fields in the entry.
func (e *Entry) FieldCount() int {
	return len(e.Fields)
}

// WithLevel returns a new entry with a different level.
// The original entry is not modified.
func (e *Entry) WithLevel(level Level) *Entry {
	cloned := e.Clone()
	cloned.Level = level
	return cloned
}

// WithMessage returns a new entry with a different message.
// The original entry is not modified.
func (e *Entry) WithMessage(msg string) *Entry {
	cloned := e.Clone()
	cloned.Message = msg
	return cloned
}

// WithTimestamp returns a new entry with a different timestamp.
// The original entry is not modified.
func (e *Entry) WithTimestamp(ts time.Time) *Entry {
	cloned := e.Clone()
	cloned.Timestamp = ts
	return cloned
}

// FormatBasic returns a basic formatted string representation of the entry.
// This is a simpler alternative to String() that's more suitable for production.
func (e *Entry) FormatBasic() string {
	var builder strings.Builder

	builder.WriteString(e.Timestamp.Format(time.RFC3339))
	builder.WriteString(" ")
	builder.WriteString(e.Level.String())
	builder.WriteString(" ")
	builder.WriteString(e.Message)

	if len(e.Fields) > 0 {
		builder.WriteString(" |")
		for _, field := range e.Fields {
			builder.WriteString(" ")
			builder.WriteString(field.Key())
			builder.WriteString("=")
			builder.WriteString(fmt.Sprintf("%v", field.Value()))
		}
	}

	return builder.String()
}

// GetFieldsByPrefix returns all fields whose keys start with the given prefix.
func (e *Entry) GetFieldsByPrefix(prefix string) []Field {
	var result []Field
	for _, field := range e.Fields {
		if strings.HasPrefix(field.Key(), prefix) {
			result = append(result, field)
		}
	}
	return result
}

// RemoveField removes the first field with the given key from the entry.
// Returns true if a field was removed, false otherwise.
func (e *Entry) RemoveField(key string) bool {
	for i, field := range e.Fields {
		if field.Key() == key {
			// Remove by swapping with last element and truncating
			e.Fields[i] = e.Fields[len(e.Fields)-1]
			e.Fields = e.Fields[:len(e.Fields)-1]
			return true
		}
	}
	return false
}

// ReplaceFieldValue replaces the value of the first field with the given key.
// Returns true if a field was found and replaced, false otherwise.
func (e *Entry) ReplaceFieldValue(key string, value interface{}) bool {
	for i := range e.Fields {
		if e.Fields[i].Key() == key {
			// Create a new field with the same key and type but new value
			e.Fields[i] = Field{
				key:       key,
				valueType: e.Fields[i].Type(),
				value:     value,
			}
			return true
		}
	}
	return false
}
