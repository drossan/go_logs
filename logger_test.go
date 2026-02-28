package go_logs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

// TestNewLogger verifies New() creates a valid logger
func TestNewLogger(t *testing.T) {
	logger, err := New()

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if logger == nil {
		t.Fatal("New() returned nil logger")
	}
}

// TestNewLoggerWithOptions verifies New() with options
func TestNewLoggerWithOptions(t *testing.T) {
	buf := &bytes.Buffer{}

	logger, err := New(
		WithLevel(DebugLevel),
		WithOutput(buf),
	)

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if logger == nil {
		t.Fatal("New() returned nil logger")
	}

	// Verify level was set
	if logger.GetLevel() != DebugLevel {
		t.Errorf("GetLevel() = %v, want DebugLevel", logger.GetLevel())
	}
}

// TestLogger_Log verifies basic Log() functionality
func TestLogger_Log(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(WithOutput(buf))

	logger.Log(InfoLevel, "test message", String("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Error("Log() output should contain message")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("Log() output should contain fields")
	}
}

// TestLogger_LevelMethods verifies convenience methods
func TestLogger_LevelMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithLevel(TraceLevel), // Log everything
	)

	tests := []struct {
		name          string
		method        func(string, ...Field)
		shouldContain string
	}{
		{"Trace", func(msg string, fields ...Field) { logger.Trace(msg) }, "TRACE"},
		{"Debug", func(msg string, fields ...Field) { logger.Debug(msg) }, "DEBUG"},
		{"Info", func(msg string, fields ...Field) { logger.Info(msg) }, "INFO"},
		{"Warn", func(msg string, fields ...Field) { logger.Warn(msg) }, "WARN"},
		{"Error", func(msg string, fields ...Field) { logger.Error(msg) }, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.method("test message")
			output := buf.String()
			if !strings.Contains(output, tt.shouldContain) {
				t.Errorf("%s() output should contain %s", tt.name, tt.shouldContain)
			}
		})
	}
}

// TestLogger_LevelFiltering verifies fast-path filtering
func TestLogger_LevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithLevel(WarnLevel), // Only log Warn and above
	)

	// These should be filtered out
	logger.Debug("debug message")
	logger.Info("info message")

	// These should be logged
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()

	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be logged")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should be logged")
	}
}

// TestLogger_SetLevel verifies SetLevel() works at runtime
func TestLogger_SetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithLevel(ErrorLevel),
	)

	// Should be filtered
	logger.Info("initial info")

	// Change level
	logger.SetLevel(InfoLevel)

	// Should now be logged
	logger.Info("new info")

	output := buf.String()

	if strings.Contains(output, "initial info") {
		t.Error("Initial info should be filtered")
	}
	if !strings.Contains(output, "new info") {
		t.Error("New info should be logged after SetLevel()")
	}
}

// TestLogger_GetLevel verifies GetLevel() returns correct level
func TestLogger_GetLevel(t *testing.T) {
	tests := []struct {
		name  string
		level Level
	}{
		{"Trace", TraceLevel},
		{"Debug", DebugLevel},
		{"Info", InfoLevel},
		{"Warn", WarnLevel},
		{"Error", ErrorLevel},
		{"Fatal", FatalLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := New(WithLevel(tt.level))
			if logger.GetLevel() != tt.level {
				t.Errorf("GetLevel() = %v, want %v", logger.GetLevel(), tt.level)
			}
		})
	}
}

// TestLogger_WithChildLogger verifies With() creates child loggers
func TestLogger_WithChildLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	parent, _ := New(
		WithOutput(buf),
		WithLevel(TraceLevel),
	)

	child := parent.With(
		String("service", "api"),
		String("version", "1.0"),
	)

	// Parent log should NOT have service field
	parent.Info("parent message")

	// Child log should have service field
	child.Info("child message", String("endpoint", "/users"))

	output := buf.String()

	// Verify parent doesn't have inherited fields
	parentLines := strings.Split(output, "\n")
	if strings.Contains(parentLines[0], "service=api") {
		t.Error("Parent log should not have child fields")
	}

	// Verify child has inherited fields
	if !strings.Contains(output, "service=api") {
		t.Error("Child log should have inherited service field")
	}
	if !strings.Contains(output, "version=1.0") {
		t.Error("Child log should have inherited version field")
	}
	if !strings.Contains(output, "endpoint=/users") {
		t.Error("Child log should have its own field")
	}
}

// TestLogger_LogCtx verifies LogCtx() works with context
func TestLogger_LogCtx(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithLevel(InfoLevel),
	)

	ctx := context.Background()
	logger.LogCtx(ctx, InfoLevel, "context message", String("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "context message") {
		t.Error("LogCtx() should log message")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("LogCtx() should include fields")
	}
}

// TestLogger_StructuredFields verifies all field types work
func TestLogger_StructuredFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(WithOutput(buf), WithLevel(InfoLevel))

	err := errors.New("test error")

	logger.Info("structured message",
		String("str", "value"),
		Int("int", 42),
		Int64("int64", int64(1234567890)),
		Float64("float", 3.14),
		Bool("bool", true),
		Err(err),
	)

	output := buf.String()

	if !strings.Contains(output, "str=value") {
		t.Error("Should contain string field")
	}
	if !strings.Contains(output, "int=42") {
		t.Error("Should contain int field")
	}
	if !strings.Contains(output, "int64=1234567890") {
		t.Error("Should contain int64 field")
	}
	if !strings.Contains(output, "float=3.14") {
		t.Error("Should contain float64 field")
	}
	if !strings.Contains(output, "bool=true") {
		t.Error("Should contain bool field")
	}
	if !strings.Contains(output, "error=test error") {
		t.Error("Should contain error field")
	}
}

// TestLogger_DefaultLevel verifies default level is Info
func TestLogger_DefaultLevel(t *testing.T) {
	logger, _ := New()
	if logger.GetLevel() != InfoLevel {
		t.Errorf("Default level should be Info, got %v", logger.GetLevel())
	}
}

// TestLogger_WithRedaction verifies redactor works
func TestLogger_WithRedaction(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithRedactor("password", "token"),
		WithLevel(InfoLevel),
	)

	logger.Info("user login",
		String("username", "john"),
		String("password", "secret123"),
		String("token", "abc-xyz"),
	)

	output := buf.String()

	// Should contain username
	if !strings.Contains(output, "username=john") {
		t.Error("Should contain non-redacted field")
	}

	// Should NOT contain actual password or token values
	if strings.Contains(output, "secret123") {
		t.Error("Password should be redacted")
	}
	if strings.Contains(output, "abc-xyz") {
		t.Error("Token should be redacted")
	}

	// Should contain masked values
	if !strings.Contains(output, "password=***") {
		t.Error("Should contain masked password")
	}
	if !strings.Contains(output, "token=***") {
		t.Error("Should contain masked token")
	}
}

// TestLogger_ConcurrentAccess verifies thread-safety
func TestLogger_ConcurrentAccess(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithLevel(InfoLevel),
	)

	// Log from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("goroutine message", Int("id", id))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	output := buf.String()

	// Should have 10 log lines
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 10 {
		t.Errorf("Expected 10 log lines, got %d", len(lines))
	}
}

// BenchmarkLogger_Log_NoFields benchmarks logging without fields
func BenchmarkLogger_Log_NoFields(b *testing.B) {
	logger, _ := New(WithOutput(io.Discard))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message")
	}
}

// BenchmarkLogger_Log_WithFields benchmarks logging with fields
func BenchmarkLogger_Log_WithFields(b *testing.B) {
	logger, _ := New(WithOutput(io.Discard))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test message",
			String("key1", "value1"),
			Int("key2", 42),
			Float64("key3", 3.14),
		)
	}
}

// BenchmarkLogger_FastPathFiltering benchmarks level filtering performance
func BenchmarkLogger_FastPathFiltering(b *testing.B) {
	logger, _ := New(
		WithOutput(io.Discard),
		WithLevel(ErrorLevel), // Filter out Info messages
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("filtered message") // Should be filtered immediately
	}
}

// BenchmarkLogger_ChildLogger benchmarks child logger creation and logging
func BenchmarkLogger_ChildLogger(b *testing.B) {
	parent, _ := New(WithOutput(io.Discard))
	child := parent.With(String("service", "test"))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		child.Info("test message")
	}
}

// TestLogger_Sync verifies Sync() can be called without error
func TestLogger_Sync(t *testing.T) {
	logger, _ := New()
	err := logger.Sync()
	if err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

// TestLogger_Fatal verifies Fatal() terminates the program
// Note: This test cannot actually test os.Exit(1) as it would terminate the test runner
func TestLogger_Fatal(t *testing.T) {
	// We can't actually call Fatal() as it would exit
	// Just verify the Log implementation is correct by testing the entry creation
	if !FatalLevel.shouldLog(InfoLevel) {
		t.Error("FatalLevel should log at InfoLevel threshold")
	}
}

// TestLogger_OutputToStdout verifies default output is stdout
func TestLogger_OutputToStdout(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger, _ := New() // No WithOutput option
	logger.Info("to stdout")

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read from pipe
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "to stdout") {
		t.Error("Default logger should write to stdout")
	}
}

// TestLogger_MultipleLoggers verifies multiple loggers work independently
func TestLogger_MultipleLoggers(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	logger1, _ := New(WithOutput(buf1), WithLevel(DebugLevel))
	logger2, _ := New(WithOutput(buf2), WithLevel(ErrorLevel))

	logger1.Debug("debug message")
	logger2.Debug("debug message")

	output1 := buf1.String()
	output2 := buf2.String()

	if !strings.Contains(output1, "debug message") {
		t.Error("Logger1 should log debug message")
	}
	if strings.Contains(output2, "debug message") {
		t.Error("Logger2 should filter debug message")
	}
}

// TestLogger_EmptyMessage verifies empty messages are handled
func TestLogger_EmptyMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(WithOutput(buf))

	logger.Info("") // Empty message

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Should log even with empty message")
	}
}

// TestLogger_NoFields verifies logging without fields works
func TestLogger_NoFields(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(WithOutput(buf))

	logger.Info("message without fields")

	output := buf.String()
	if !strings.Contains(output, "message without fields") {
		t.Error("Should log message without fields")
	}
}
