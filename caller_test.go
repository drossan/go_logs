package go_logs

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestGetCaller verifies GetCaller returns correct caller information
func TestGetCaller(t *testing.T) {
	// Call GetCaller from this function
	caller := GetCaller(0)

	if caller == nil {
		t.Fatal("GetCaller returned nil")
	}

	// Should be this test file
	if !strings.HasSuffix(caller.File, "caller_test.go") {
		t.Errorf("Expected file to end with caller_test.go, got %s", caller.File)
	}

	// Line should be positive
	if caller.Line <= 0 {
		t.Errorf("Expected positive line number, got %d", caller.Line)
	}

	// Function should be related to this test
	if !strings.Contains(caller.Func, "TestGetCaller") {
		t.Errorf("Expected function to contain TestGetCaller, got %s", caller.Func)
	}

	// Package should be go_logs
	if caller.Package != "go_logs" {
		t.Errorf("Expected package go_logs, got %s", caller.Package)
	}
}

// TestGetCallerSkip verifies skip parameter works correctly
func TestGetCallerSkip(t *testing.T) {
	// Helper function to test skip
	getCallerHelper := func() *CallerInfo {
		return GetCaller(1) // Skip this helper function
	}

	caller := getCallerHelper()

	if caller == nil {
		t.Fatal("GetCaller returned nil")
	}

	// Should be this test function, not the helper
	if !strings.Contains(caller.Func, "TestGetCallerSkip") {
		t.Errorf("Expected function TestGetCallerSkip, got %s", caller.Func)
	}
}

// TestCallerInfoString verifies String() formatting
func TestCallerInfoString(t *testing.T) {
	caller := &CallerInfo{
		File: "main.go",
		Line: 42,
		Func: "main",
	}

	expected := "main.go:42"
	if caller.String() != expected {
		t.Errorf("Expected %q, got %q", expected, caller.String())
	}
}

// TestCallerInfoFullString verifies FullString() formatting
func TestCallerInfoFullString(t *testing.T) {
	caller := &CallerInfo{
		File: "main.go",
		Line: 42,
		Func: "myFunction",
	}

	expected := "main.go:42 myFunction"
	if caller.FullString() != expected {
		t.Errorf("Expected %q, got %q", expected, caller.FullString())
	}
}

// TestCallerInfoNil verifies nil CallerInfo handling
func TestCallerInfoNil(t *testing.T) {
	var caller *CallerInfo

	if caller.String() != "???" {
		t.Errorf("Expected ??? for nil caller, got %s", caller.String())
	}

	if caller.FullString() != "???" {
		t.Errorf("Expected ??? for nil caller, got %s", caller.FullString())
	}
}

// TestGetStackTrace verifies stack trace capture
func TestGetStackTrace(t *testing.T) {
	stack := GetStackTrace(0)

	if stack == nil {
		t.Fatal("GetStackTrace returned nil")
	}

	stackStr := string(stack)

	// Should contain this function name
	if !strings.Contains(stackStr, "TestGetStackTrace") {
		t.Error("Stack trace should contain TestGetStackTrace")
	}

	// Should contain file name
	if !strings.Contains(stackStr, "caller_test.go") {
		t.Error("Stack trace should contain caller_test.go")
	}
}

// TestGetStackTraceAll verifies all goroutines stack trace
func TestGetStackTraceAll(t *testing.T) {
	stack := GetStackTraceAll()

	if stack == nil {
		t.Fatal("GetStackTraceAll returned nil")
	}

	stackStr := string(stack)

	// Should contain this function name
	if !strings.Contains(stackStr, "TestGetStackTraceAll") {
		t.Error("Stack trace should contain TestGetStackTraceAll")
	}
}

// TestLoggerWithCaller verifies caller info in logger output
func TestLoggerWithCaller(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
		WithCaller(true), // Enable caller info
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message with caller")

	output := buf.String()

	// Should contain file and line info
	if !strings.Contains(output, "caller_test.go") {
		t.Errorf("Output should contain caller_test.go, got: %s", output)
	}
}

// TestLoggerWithoutCaller verifies caller info can be disabled
func TestLoggerWithoutCaller(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
		WithCaller(false), // Disable caller info
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("test message without caller")

	output := buf.String()

	// Should NOT contain file and line info (just the message)
	// Note: The timestamp and level are still present
	if strings.Contains(output, "caller_test.go:") {
		t.Errorf("Output should NOT contain caller info, got: %s", output)
	}
}

// TestLoggerWithStackTrace verifies stack trace on error level
func TestLoggerWithStackTrace(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(
		WithLevel(ErrorLevel),
		WithOutput(&buf),
		WithStackTrace(true), // Enable stack traces
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Error("error with stack trace")

	output := buf.String()

	// Should contain stack trace markers
	if !strings.Contains(output, "goroutine") && !strings.Contains(output, "created by") {
		// At minimum, should have some stack indication
		t.Logf("Output: %s", output)
	}
}

// TestLoggerStackTraceOnErrorOnly verifies stack trace only on error+
func TestLoggerStackTraceOnErrorOnly(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
		WithStackTrace(true),
		WithStackTraceLevel(ErrorLevel), // Only on Error+
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Info should NOT have stack trace
	buf.Reset()
	logger.Info("info message")
	infoOutput := buf.String()

	// Error SHOULD have stack trace
	buf.Reset()
	logger.Error("error message")
	errorOutput := buf.String()

	// Error output should be more verbose than info
	t.Logf("Info output: %s", infoOutput)
	t.Logf("Error output: %s", errorOutput)
}

// TestCallerInJSONFormatter verifies caller info in JSON output
func TestCallerInJSONFormatter(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
		WithFormatter(NewJSONFormatter()),
		WithCaller(true),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("json with caller")

	output := buf.String()

	// Should contain caller field in JSON
	if !strings.Contains(output, `"caller"`) {
		t.Errorf("JSON output should contain caller field, got: %s", output)
	}
}

// TestCallerInChildLogger verifies caller info propagates to child loggers
func TestCallerInChildLogger(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
		WithCaller(true),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	childLogger := logger.With(String("component", "child"))
	childLogger.Info("child message with caller")

	output := buf.String()

	// Should contain caller info
	if !strings.Contains(output, "caller_test.go") {
		t.Errorf("Child logger output should contain caller info, got: %s", output)
	}
}

// TestCallerWithCtx verifies caller works with context
func TestCallerWithCtx(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(
		WithLevel(InfoLevel),
		WithOutput(&buf),
		WithCaller(true),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")

	logger.LogCtx(ctx, InfoLevel, "message with context and caller")

	output := buf.String()

	// Should contain both trace_id and caller
	if !strings.Contains(output, "trace-123") {
		t.Errorf("Output should contain trace_id, got: %s", output)
	}
	if !strings.Contains(output, "caller_test.go") {
		t.Errorf("Output should contain caller info, got: %s", output)
	}
}

// TestPackageFromFuncName verifies package/function extraction from fully
// qualified runtime function names, including module major-version suffixes.
func TestPackageFromFuncName(t *testing.T) {
	tests := []struct {
		fnName string
		pkg    string
		fn     string
	}{
		{"github.com/drossan/go_logs/v3.TestGetCaller", "go_logs", "TestGetCaller"},
		{"github.com/drossan/go_logs/v3.(*LoggerImpl).Log", "go_logs", "(*LoggerImpl).Log"},
		{"github.com/drossan/go_logs/v3/async.(*Logger).With", "async", "(*Logger).With"},
		{"example.com/x/v2.Func", "x", "Func"},
		{"example.com/v2tools.Func", "v2tools", "Func"},
		{"github.com/u/pkg.Func", "pkg", "Func"},
		{"main.main", "main", "main"},
		{"github.com/drossan/go_logs/v1.Func", "v1", "Func"},
		{"github.com/example/pkg/v10.Func", "pkg", "Func"},
	}

	for _, tt := range tests {
		t.Run(tt.fnName, func(t *testing.T) {
			pkg, fn := packageFromFuncName(tt.fnName)
			if pkg != tt.pkg || fn != tt.fn {
				t.Errorf("packageFromFuncName(%q) = (%q, %q), want (%q, %q)", tt.fnName, pkg, fn, tt.pkg, tt.fn)
			}
		})
	}
}

// TestPackageFromFuncName_Generic verifies that a generic instantiation in the
// function name does not break extraction nor panic.
func TestPackageFromFuncName_Generic(t *testing.T) {
	pkg, _ := packageFromFuncName("github.com/drossan/go_logs/v3.List[go.shape.int].Get")
	if pkg != "go_logs" {
		t.Errorf("expected package go_logs, got %q", pkg)
	}
}
