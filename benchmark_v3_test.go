package go_logs

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"testing"
)

// BenchmarkV3LoggerCreation measures the performance of creating a new logger
func BenchmarkV3LoggerCreation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = New(WithLevel(InfoLevel))
	}
}

// BenchmarkV3LoggerCreationWithOptions measures creating logger with multiple options
func BenchmarkV3LoggerCreationWithOptions(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = New(
			WithLevel(InfoLevel),
			WithFormatter(NewJSONFormatter()),
			WithOutput(io.Discard),
		)
	}
}

// BenchmarkV3LevelFiltering measures the fast-path level filtering (filtered out)
func BenchmarkV3LevelFiltering(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("this should be filtered") // Below InfoLevel
	}
}

// BenchmarkV3LevelFilteringPass measures logging that passes the filter
func BenchmarkV3LevelFilteringPass(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("this should pass")
	}
}

// BenchmarkV3FieldCreation measures creating structured fields
func BenchmarkV3FieldCreation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = String("key", "value")
		_ = Int("count", 42)
		_ = Err(io.EOF)
	}
}

// BenchmarkV3TextFormatter measures TextFormatter performance
func BenchmarkV3TextFormatter(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewTextFormatter()),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			String("host", "localhost"),
			Int("port", 8080),
		)
	}
}

// BenchmarkV3JSONFormatter measures JSONFormatter performance
func BenchmarkV3JSONFormatter(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			String("host", "localhost"),
			Int("port", 8080),
		)
	}
}

// BenchmarkV3JSONFormatterWithFields measures JSON with many fields
func BenchmarkV3JSONFormatterWithFields(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			String("request_id", "abc-123"),
			String("user_id", "user-456"),
			String("method", "GET"),
			String("path", "/api/users"),
			Int("status", 200),
			Int64("duration_ms", 42),
		)
	}
}

// BenchmarkV3WithChildLogger measures child logger creation via With()
func BenchmarkV3WithChildLogger(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		child := logger.With(String("request_id", "abc-123"))
		child.Info("test message")
	}
}

// BenchmarkV3WithChildLoggerReuse measures logging with a reused child logger
func BenchmarkV3WithChildLoggerReuse(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
	)
	child := logger.With(String("request_id", "abc-123"))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		child.Info("test message")
	}
}

// BenchmarkV3ContextPropagation measures LogCtx performance
func BenchmarkV3ContextPropagation(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
	)
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogCtx(ctx, InfoLevel, "benchmark message",
			String("key", "value"),
		)
	}
}

// BenchmarkV3ContextWithTraceID measures LogCtx with trace ID
func BenchmarkV3ContextWithTraceID(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
	)
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.LogCtx(ctx, InfoLevel, "benchmark message")
	}
}

// BenchmarkV3WithHook measures logging with a hook
func BenchmarkV3WithHook(b *testing.B) {
	hook := NewFuncHook(func(entry *Entry) error {
		return nil // No-op hook
	})
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
		WithHooks(hook),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

// BenchmarkV3WithMultipleHooks measures logging with multiple hooks
func BenchmarkV3WithMultipleHooks(b *testing.B) {
	hook1 := NewFuncHook(func(entry *Entry) error { return nil })
	hook2 := NewFuncHook(func(entry *Entry) error { return nil })
	hook3 := NewFuncHook(func(entry *Entry) error { return nil })
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
		WithHooks(hook1, hook2, hook3),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

// BenchmarkV3ErrorWithCaller measures Error logging with auto-caller (default)
func BenchmarkV3ErrorWithCaller(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("error message", Err(io.EOF))
	}
}

// BenchmarkV3InfoWithoutCaller measures Info logging without caller (default)
func BenchmarkV3InfoWithoutCaller(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("info message")
	}
}

// BenchmarkV3InfoWithCallerEnabled measures Info with caller explicitly enabled
func BenchmarkV3InfoWithCallerEnabled(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(io.Discard),
		WithCaller(true),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("info message")
	}
}

// BenchmarkV3ConcurrentLogging measures concurrent logging performance
func BenchmarkV3ConcurrentLogging(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(io.Discard),
	)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("concurrent message",
				String("goroutine", "test"),
				Int("iteration", 1),
			)
		}
	})
}

// BenchmarkV3Metrics measures logging with metrics collection
func BenchmarkV3Metrics(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithOutput(io.Discard),
	)
	// Type assert to access GetMetrics
	if impl, ok := logger.(*LoggerImpl); ok {
		metrics := impl.GetMetrics()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("benchmark message")
			_ = metrics.Total()
		}
	}
}

// BenchmarkV3RotatingFileWriter measures rotating file writer performance
func BenchmarkV3RotatingFileWriter(b *testing.B) {
	// Use a buffer to avoid actual file I/O
	var buf bytes.Buffer
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(&buf),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			String("key", "value"),
			Int("count", i),
		)
	}
}

// BenchmarkLoggerToRotatingFile measures the end-to-end cost of one entry to a
// real file: Logger -> JSONFormatter -> RotatingFileWriter (per-entry Flush,
// no fsync).
func BenchmarkLoggerToRotatingFile(b *testing.B) {
	w, err := NewRotatingFileWriter(filepath.Join(b.TempDir(), "bench.log"), 1000, 1)
	if err != nil {
		b.Fatalf("NewRotatingFileWriter() error = %v", err)
	}
	defer w.Close()
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(w),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			String("key", "value"),
			Int("count", i),
		)
	}
}

// BenchmarkV3Redaction measures logging with redaction enabled
func BenchmarkV3Redaction(b *testing.B) {
	logger, _ := New(
		WithLevel(InfoLevel),
		WithFormatter(NewJSONFormatter()),
		WithOutput(io.Discard),
		WithCommonRedaction(),
	)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			String("password", "secret123"),
			String("token", "abc-xyz"),
			String("public", "data"),
		)
	}
}
