package go_logs

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// newTestLogger creates a logger with a buffer for testing
func newTestLogger(opts ...Option) (Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	allOpts := append([]Option{WithOutput(buf), WithLevel(TraceLevel)}, opts...)
	logger, _ := New(allOpts...)
	return logger, buf
}

// TestWithTraceID tests that WithTraceID adds trace_id to context
func TestWithTraceID(t *testing.T) {
	ctx := context.Background()

	traceID := "abc-123-def-456"
	ctx = WithTraceID(ctx, traceID)

	extracted := GetTraceID(ctx)
	if extracted != traceID {
		t.Errorf("GetTraceID() = %s, want %s", extracted, traceID)
	}
}

// TestWithSpanID tests that WithSpanID adds span_id to context
func TestWithSpanID(t *testing.T) {
	ctx := context.Background()

	spanID := "span-xyz"
	ctx = WithSpanID(ctx, spanID)

	extracted := GetSpanID(ctx)
	if extracted != spanID {
		t.Errorf("GetSpanID() = %s, want %s", extracted, spanID)
	}
}

// TestWithRequestID tests that WithRequestID adds request_id to context
func TestWithRequestID(t *testing.T) {
	ctx := context.Background()

	requestID := "req-123"
	ctx = WithRequestID(ctx, requestID)

	extracted := GetRequestID(ctx)
	if extracted != requestID {
		t.Errorf("GetRequestID() = %s, want %s", extracted, requestID)
	}
}

// TestExtractFieldsFromContext tests extracting fields from context
func TestExtractFieldsFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		wantKeys []string
	}{
		{
			name: "extract trace_id",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return WithTraceID(ctx, "trace-abc")
			},
			wantKeys: []string{"trace_id"},
		},
		{
			name: "extract span_id",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return WithSpanID(ctx, "span-xyz")
			},
			wantKeys: []string{"span_id"},
		},
		{
			name: "extract request_id",
			setupCtx: func() context.Context {
				ctx := context.Background()
				return WithRequestID(ctx, "req-123")
			},
			wantKeys: []string{"request_id"},
		},
		{
			name: "extract all fields",
			setupCtx: func() context.Context {
				ctx := context.Background()
				ctx = WithTraceID(ctx, "trace-abc")
				ctx = WithSpanID(ctx, "span-xyz")
				ctx = WithRequestID(ctx, "req-123")
				return ctx
			},
			wantKeys: []string{"trace_id", "span_id", "request_id"},
		},
		{
			name: "empty context returns no fields",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			fields := ExtractFieldsFromContext(ctx)

			if len(fields) != len(tt.wantKeys) {
				t.Errorf("ExtractFieldsFromContext() returned %d fields, want %d", len(fields), len(tt.wantKeys))
				return
			}

			// Check that all expected keys are present
			fieldKeys := make(map[string]bool)
			for _, f := range fields {
				fieldKeys[f.Key()] = true
			}

			for _, key := range tt.wantKeys {
				if !fieldKeys[key] {
					t.Errorf("ExtractFieldsFromContext() missing key: %s", key)
				}
			}
		})
	}
}

// TestLogCtx_ExtractsContextFields tests that LogCtx extracts and uses context fields
func TestLogCtx_ExtractsContextFields(t *testing.T) {
	logger, buf := newTestLogger()

	// Create context with trace_id
	traceID := "trace-test-123"
	ctx := WithTraceID(context.Background(), traceID)

	// Log with context
	logger.LogCtx(ctx, InfoLevel, "test message", String("key", "value"))

	output := buf.String()

	// Should contain the trace_id from context
	if !strings.Contains(output, "trace_id="+traceID) {
		t.Errorf("LogCtx output should contain trace_id from context\nGot: %s", output)
	}

	// Should contain the explicit field
	if !strings.Contains(output, "key=value") {
		t.Errorf("LogCtx output should contain explicit field\nGot: %s", output)
	}

	// Should contain the message
	if !strings.Contains(output, "test message") {
		t.Errorf("LogCtx output should contain message\nGot: %s", output)
	}
}

// TestLogger_With_ChildLogger tests child logger functionality
func TestLogger_With_ChildLogger(t *testing.T) {
	parent, buf := newTestLogger()

	// Create child logger with service field
	child := parent.With(String("service", "my-service"))

	// Parent log should NOT have service field
	parent.Info("parent message")

	// Child log should have service field
	child.Info("child message", String("request_id", "req-123"))

	output := buf.String()

	// Check parent message (no service field)
	parentLines := strings.Split(output, "\n")
	var parentLine string
	for _, line := range parentLines {
		if strings.Contains(line, "parent message") {
			parentLine = line
			break
		}
	}
	if parentLine != "" && strings.Contains(parentLine, "service=my-service") {
		t.Errorf("Parent log should NOT contain service field\nGot: %s", parentLine)
	}

	// Check child message (should have service field)
	childLines := strings.Split(output, "\n")
	var childLine string
	for _, line := range childLines {
		if strings.Contains(line, "child message") {
			childLine = line
			break
		}
	}
	if childLine == "" {
		t.Fatal("Child log not found in output")
	}
	if !strings.Contains(childLine, "service=my-service") {
		t.Errorf("Child log should contain service field\nGot: %s", childLine)
	}
	if !strings.Contains(childLine, "request_id=req-123") {
		t.Errorf("Child log should contain request_id field\nGot: %s", childLine)
	}
}

// TestLogger_With_ChildOfChild tests nested child loggers
func TestLogger_With_ChildOfChild(t *testing.T) {
	base, buf := newTestLogger()

	// Create child with service
	child1 := base.With(String("service", "api"))

	// Create grandchild with request_id
	child2 := child1.With(String("request_id", "req-123"))

	// Grandchild should have both fields
	child2.Info("processing request")

	output := buf.String()

	if !strings.Contains(output, "service=api") {
		t.Errorf("Grandchild log should contain service field\nGot: %s", output)
	}
	if !strings.Contains(output, "request_id=req-123") {
		t.Errorf("Grandchild log should contain request_id field\nGot: %s", output)
	}
}

// TestLogger_With_ParentNotModified tests that parent is not modified when creating child
func TestLogger_With_ParentNotModified(t *testing.T) {
	parent, buf := newTestLogger()

	// Create child with extra field
	child := parent.With(String("child_field", "child_value"))

	// Log from parent (should NOT have child_field)
	parent.Info("parent message")

	// Log from child (should have child_field)
	child.Info("child message")

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Find parent line
	var parentLine string
	for _, line := range lines {
		if strings.Contains(line, "parent message") {
			parentLine = line
			break
		}
	}

	if parentLine != "" && strings.Contains(parentLine, "child_field") {
		t.Errorf("Parent should NOT be modified by child creation\nGot: %s", parentLine)
	}
}

// TestContextFieldHelpers tests the helper functions for context fields
func TestContextFieldHelpers(t *testing.T) {
	t.Run("GetTraceID returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		traceID := GetTraceID(ctx)
		if traceID != "" {
			t.Errorf("GetTraceID() = %s, want empty string", traceID)
		}
	})

	t.Run("GetSpanID returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		spanID := GetSpanID(ctx)
		if spanID != "" {
			t.Errorf("GetSpanID() = %s, want empty string", spanID)
		}
	})

	t.Run("GetRequestID returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		requestID := GetRequestID(ctx)
		if requestID != "" {
			t.Errorf("GetRequestID() = %s, want empty string", requestID)
		}
	})

	t.Run("multiple values in context", func(t *testing.T) {
		ctx := context.Background()
		ctx = WithTraceID(ctx, "trace-1")
		ctx = WithSpanID(ctx, "span-1")
		ctx = WithRequestID(ctx, "req-1")

		if got := GetTraceID(ctx); got != "trace-1" {
			t.Errorf("GetTraceID() = %s, want trace-1", got)
		}
		if got := GetSpanID(ctx); got != "span-1" {
			t.Errorf("GetSpanID() = %s, want span-1", got)
		}
		if got := GetRequestID(ctx); got != "req-1" {
			t.Errorf("GetRequestID() = %s, want req-1", got)
		}
	})
}
