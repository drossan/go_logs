package go_logs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
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

// TestLogger_LogCtx_NilContext verifies LogCtx() does not panic when passed a nil context.
func TestLogger_LogCtx_NilContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithLevel(InfoLevel),
	)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("LogCtx(nil, ...) panicked: %v", r)
		}
	}()

	logger.LogCtx(nil, InfoLevel, "nil context message")

	output := buf.String()
	if !strings.Contains(output, "nil context message") {
		t.Error("LogCtx(nil, ...) should still log the message")
	}
}

// TestLogger_SuccessLevel_VisibleAtDefaultLevel verifies SuccessLevel messages are
// not filtered out when the logger uses the default level (InfoLevel).
func TestLogger_SuccessLevel_VisibleAtDefaultLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(WithOutput(buf))

	if logger.GetLevel() != InfoLevel {
		t.Fatalf("expected default level to be InfoLevel, got %v", logger.GetLevel())
	}

	logger.Log(SuccessLevel, "success message")

	output := buf.String()
	if !strings.Contains(output, "success message") {
		t.Error("SuccessLevel message should be visible at the default log level")
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

// lockedBuffer is a bytes.Buffer safe for concurrent writers. Children of the
// same logger do not share a write mutex, so concurrent tests need it.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// parentWithSpareCapacity returns a logger whose inherited fields slice has
// len 3 and spare capacity, the precondition of the shared-array bug.
func parentWithSpareCapacity(t *testing.T, out io.Writer) Logger {
	t.Helper()
	root, err := New(WithOutput(out), WithLevel(InfoLevel))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	impl := root.(*LoggerImpl)
	impl.fields = append(make([]Field, 0, 8), Int("a", 1), Int("b", 2), Int("c", 3))
	if cap(impl.fields) <= len(impl.fields) {
		t.Fatalf("precondition: fields must have spare capacity, len=%d cap=%d", len(impl.fields), cap(impl.fields))
	}
	return root
}

// lineWith returns the first output line containing msg.
func lineWith(t *testing.T, output, msg string) string {
	t.Helper()
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, msg) {
			return line
		}
	}
	t.Fatalf("no line contains %q in output:\n%s", msg, output)
	return ""
}

func TestLogger_With_SiblingsDoNotShareFields(t *testing.T) {
	buf := &bytes.Buffer{}
	parent := parentWithSpareCapacity(t, buf)

	childA := parent.With(String("req", "AAA"))
	childB := parent.With(String("req", "BBB"))

	childA.Info("handling A")
	childB.Info("handling B")

	lineA := lineWith(t, buf.String(), "handling A")
	lineB := lineWith(t, buf.String(), "handling B")
	if !strings.Contains(lineA, "req=AAA") || strings.Contains(lineA, "BBB") {
		t.Errorf("child A line = %q, want req=AAA and no BBB", lineA)
	}
	if !strings.Contains(lineB, "req=BBB") || strings.Contains(lineB, "AAA") {
		t.Errorf("child B line = %q, want req=BBB and no AAA", lineB)
	}
}

func TestLogger_With_FromParentWithoutFields(t *testing.T) {
	parent, buf := newTestLogger()

	child := parent.With(Int("x", 1))
	child.Info("child msg")
	parent.Info("parent msg")

	if line := lineWith(t, buf.String(), "child msg"); !strings.Contains(line, "x=1") {
		t.Errorf("child line = %q, want x=1", line)
	}
	if line := lineWith(t, buf.String(), "parent msg"); strings.Contains(line, "x=1") {
		t.Errorf("parent line = %q, want no x=1", line)
	}
	if n := len(parent.(*LoggerImpl).fields); n != 0 {
		t.Errorf("parent fields len = %d, want 0", n)
	}
}

func TestLogger_With_NoArgsIsEquivalentToParent(t *testing.T) {
	root, buf := newTestLogger()
	parent := root.With(Int("a", 1))

	child := parent.With()
	parent.Info("from parent")
	child.Info("from child")

	if line := lineWith(t, buf.String(), "from child"); !strings.Contains(line, "a=1") {
		t.Errorf("child line = %q, want a=1", line)
	}
	if line := lineWith(t, buf.String(), "from parent"); !strings.Contains(line, "a=1") {
		t.Errorf("parent line = %q, want a=1", line)
	}
}

// Duplicate keys are not deduplicated (see task design note): both appear.
func TestLogger_With_DuplicateKeyKeepsBoth(t *testing.T) {
	root, buf := newTestLogger()
	parent := root.With(String("env", "prod"))

	parent.With(String("env", "staging")).Info("dup")

	line := lineWith(t, buf.String(), "dup")
	if !strings.Contains(line, "env=prod") || !strings.Contains(line, "env=staging") {
		t.Errorf("line = %q, want both env=prod and env=staging", line)
	}
}

func TestLogger_With_DoesNotAlterParent(t *testing.T) {
	root, buf := newTestLogger()
	parent := root.With(Int("a", 1))

	_ = parent.With(Int("b", 2))
	parent.Info("solo padre")

	line := lineWith(t, buf.String(), "solo padre")
	if !strings.Contains(line, "a=1") || strings.Contains(line, "b=2") {
		t.Errorf("parent line = %q, want a=1 and no b=2", line)
	}
}

func TestLogger_With_ConcurrentChildrenAndSetLevel(t *testing.T) {
	buf := &lockedBuffer{}
	parent := parentWithSpareCapacity(t, buf)

	const n = 50
	stop := make(chan struct{})
	levelDone := make(chan struct{})
	go func() {
		defer close(levelDone)
		for {
			select {
			case <-stop:
				return
			default:
				parent.SetLevel(TraceLevel)
				parent.SetLevel(InfoLevel)
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			parent.With(String("g", fmt.Sprintf("g%02d", id))).Info(fmt.Sprintf("msg-%02d", id))
		}(i)
	}
	wg.Wait()
	close(stop)
	<-levelDone

	output := buf.String()
	for i := 0; i < n; i++ {
		line := lineWith(t, output, fmt.Sprintf("msg-%02d", i))
		if got := strings.Count(line, "g=g"); got != 1 || !strings.Contains(line, fmt.Sprintf("g=g%02d", i)) {
			t.Errorf("line for goroutine %d = %q, want only g=g%02d", i, line, i)
		}
	}
}

func TestLogger_SharedLevel_ParentChangeReachesExistingChild(t *testing.T) {
	parent, buf := newTestLogger(WithLevel(InfoLevel))
	child := parent.With(String("k", "v"))

	parent.SetLevel(DebugLevel)
	child.Debug("child debug")

	if !strings.Contains(buf.String(), "child debug") {
		t.Errorf("child debug message was not emitted after parent.SetLevel(Debug); output=%q", buf.String())
	}
}

func TestLogger_SharedLevel_ChildChangeReachesParent(t *testing.T) {
	parent, buf := newTestLogger(WithLevel(InfoLevel))
	child := parent.With(String("k", "v"))

	child.SetLevel(ErrorLevel)
	parent.Info("parent info")

	if strings.Contains(buf.String(), "parent info") {
		t.Errorf("parent info message was emitted after child.SetLevel(Error); output=%q", buf.String())
	}
}

func TestLogger_SharedLevel_GrandchildReachesRoot(t *testing.T) {
	root, buf := newTestLogger(WithLevel(InfoLevel))
	grandchild := root.With(Int("gen", 1)).With(Int("gen", 2))

	grandchild.SetLevel(ErrorLevel)
	root.Warn("root warn")

	if strings.Contains(buf.String(), "root warn") {
		t.Errorf("root warn message was emitted after grandchild.SetLevel(Error); output=%q", buf.String())
	}
}

func TestLogger_SharedLevel_ChildGetLevel(t *testing.T) {
	parent, _ := newTestLogger(WithLevel(WarnLevel))

	if got := parent.With(Int("a", 1)).GetLevel(); got != WarnLevel {
		t.Errorf("child GetLevel() = %v, want Warn", got)
	}
}

// Level filtering must not touch l.mu: filtered calls complete while the
// configuration mutex is held by someone else.
func TestLogger_LevelFilteringDoesNotTakeMutex(t *testing.T) {
	logger, buf := newTestLogger(WithLevel(ErrorLevel))
	impl := logger.(*LoggerImpl)

	impl.mu.Lock()
	done := make(chan struct{})
	go func() {
		defer close(done)
		var wg sync.WaitGroup
		for g := 0; g < 8; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 1000; i++ {
					logger.Debug("filtered")
				}
			}()
		}
		wg.Wait()
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		impl.mu.Unlock()
		<-done
		t.Fatal("filtered Debug calls blocked on l.mu")
	}
	impl.mu.Unlock()

	if buf.Len() != 0 {
		t.Errorf("output = %q, want empty", buf.String())
	}
}

// spyWriter is an io.Writer that counts Write, Flush and Sync calls and lets
// tests inject the errors returned by Flush and Sync. It is safe for
// concurrent use.
type spyWriter struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	writes   int
	flushes  int
	syncs    int
	flushErr error
	syncErr  error
}

func (s *spyWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.writes++
	return s.buf.Write(p)
}

func (s *spyWriter) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushes++
	return s.flushErr
}

func (s *spyWriter) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncs++
	return s.syncErr
}

// counts returns the number of Write, Flush and Sync calls so far.
func (s *spyWriter) counts() (writes, flushes, syncs int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writes, s.flushes, s.syncs
}
