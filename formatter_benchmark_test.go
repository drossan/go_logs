package go_logs

import (
	"encoding/json"
	"testing"
	"time"
)

// BenchmarkTextFormatter_SimpleEntry benchmarks formatting a simple entry without fields.
func BenchmarkTextFormatter_SimpleEntry(b *testing.B) {
	formatter := NewTextFormatter()
	entry := &Entry{
		Level:     InfoLevel,
		Message:   "simple message",
		Timestamp: time.Now(),
		Fields:    []Field{},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

// BenchmarkTextFormatter_WithFields benchmarks formatting an entry with multiple fields.
func BenchmarkTextFormatter_WithFields(b *testing.B) {
	formatter := NewTextFormatter()
	entry := &Entry{
		Level:     InfoLevel,
		Message:   "database connection",
		Timestamp: time.Now(),
		Fields: []Field{
			String("host", "localhost"),
			Int("port", 5432),
			String("user", "admin"),
			Bool("ssl", true),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

// BenchmarkTextFormatter_WithFields benchmarks formatting an entry with many fields.
func BenchmarkTextFormatter_WithManyFields(b *testing.B) {
	formatter := NewTextFormatter()
	entry := &Entry{
		Level:     ErrorLevel,
		Message:   "request processing error",
		Timestamp: time.Now(),
		Fields: []Field{
			String("request_id", "abc-123-def-456"),
			String("user_id", "user-789"),
			String("ip", "192.168.1.100"),
			Int("status_code", 500),
			Int("response_time_ms", 1234),
			String("method", "POST"),
			String("endpoint", "/api/v1/users"),
			String("error_type", "DatabaseError"),
			String("error_message", "connection timeout"),
			Bool("retry", false),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

// BenchmarkJSONFormatter_SimpleEntry benchmarks JSON formatting a simple entry.
func BenchmarkJSONFormatter_SimpleEntry(b *testing.B) {
	formatter := NewJSONFormatter()
	entry := &Entry{
		Level:     InfoLevel,
		Message:   "simple message",
		Timestamp: time.Now(),
		Fields:    []Field{},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

// BenchmarkJSONFormatter_WithFields benchmarks JSON formatting an entry with multiple fields.
func BenchmarkJSONFormatter_WithFields(b *testing.B) {
	formatter := NewJSONFormatter()
	entry := &Entry{
		Level:     InfoLevel,
		Message:   "database connection",
		Timestamp: time.Now(),
		Fields: []Field{
			String("host", "localhost"),
			Int("port", 5432),
			String("user", "admin"),
			Bool("ssl", true),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

// BenchmarkJSONFormatter_WithManyFields benchmarks JSON formatting with many fields.
func BenchmarkJSONFormatter_WithManyFields(b *testing.B) {
	formatter := NewJSONFormatter()
	entry := &Entry{
		Level:     ErrorLevel,
		Message:   "request processing error",
		Timestamp: time.Now(),
		Fields: []Field{
			String("request_id", "abc-123-def-456"),
			String("user_id", "user-789"),
			String("ip", "192.168.1.100"),
			Int("status_code", 500),
			Int("response_time_ms", 1234),
			String("method", "POST"),
			String("endpoint", "/api/v1/users"),
			String("error_type", "DatabaseError"),
			String("error_message", "connection timeout"),
			Bool("retry", false),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

// BenchmarkFieldCreation benchmarks creating Field structs.
func BenchmarkFieldCreation(b *testing.B) {
	b.Run("String", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = String("key", "value")
		}
	})

	b.Run("Int", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Int("key", 42)
		}
	})

	b.Run("Float64", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Float64("key", 3.14)
		}
	})

	b.Run("Bool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Bool("key", true)
		}
	})

	b.Run("Err", func(b *testing.B) {
		err := &benchTestError{msg: "test error"}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Err(err)
		}
	})

	b.Run("Multiple", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fields := []Field{
				String("key1", "value1"),
				Int("key2", 42),
				Float64("key3", 3.14),
				Bool("key4", true),
			}
			_ = fields
		}
	})
}

// BenchmarkEntryString benchmarks the Entry.String() method.
func BenchmarkEntryString(b *testing.B) {
	entry := &Entry{
		Level:     InfoLevel,
		Message:   "test message",
		Timestamp: time.Now(),
		Fields: []Field{
			String("key1", "value1"),
			Int("key2", 42),
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = entry.String()
	}
}

// BenchmarkLevelShouldLog benchmarks the Level.shouldLog() fast-path filtering.
func BenchmarkLevelShouldLog(b *testing.B) {
	threshold := InfoLevel

	b.Run("Passing", func(b *testing.B) {
		level := ErrorLevel
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = level.shouldLog(threshold)
		}
	})

	b.Run("Filtered", func(b *testing.B) {
		level := DebugLevel
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = level.shouldLog(threshold)
		}
	})
}

// BenchmarkJSONMarshalStdlib benchmarks standard JSON marshaling for comparison.
func BenchmarkJSONMarshalStdlib(b *testing.B) {
	entry := JSONLogEntry{
		Level:     "INFO",
		Message:   "test message",
		Timestamp: time.Now().Format(time.RFC3339),
		Fields: map[string]interface{}{
			"host": "localhost",
			"port": 5432,
			"user": "admin",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(entry)
	}
}

// Helper types for benchmarks
type benchTestError struct {
	msg string
}

func (e *benchTestError) Error() string {
	return e.msg
}
