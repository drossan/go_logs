package async

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	go_logs "github.com/drossan/go_logs"
)

// TestAsyncLogger_Wrap Creates an async wrapper around a sync logger
func TestAsyncLogger_Wrap(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
		go_logs.WithFormatter(go_logs.NewTextFormatter()),
	)

	asyncLogger := Wrap(syncLogger, 100)
	defer asyncLogger.Sync()

	asyncLogger.Info("test message")
	asyncLogger.Sync() // Ensure message is written

	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Expected log message to be written, got: %s", buf.String())
	}
}

// TestAsyncLogger_NonBlocking tests that logging doesn't block
func TestAsyncLogger_NonBlocking(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	asyncLogger := Wrap(syncLogger, 1000)
	defer asyncLogger.Sync()

	// Log many messages - should not block even with slow output
	start := time.Now()
	for i := 0; i < 100; i++ {
		asyncLogger.Info("message")
	}
	elapsed := time.Since(start)

	// Should complete in < 10ms (non-blocking)
	if elapsed > 10*time.Millisecond {
		t.Errorf("Logging took too long: %v (expected < 10ms)", elapsed)
	}

	asyncLogger.Sync()
}

// TestAsyncLogger_OrderPreserved tests that log order is preserved
func TestAsyncLogger_OrderPreserved(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
		go_logs.WithFormatter(go_logs.NewTextFormatter()),
	)

	asyncLogger := Wrap(syncLogger, 100)
	defer asyncLogger.Sync()

	// Log messages in sequence
	for i := 0; i < 10; i++ {
		asyncLogger.Info("message", go_logs.Int("seq", i))
	}
	asyncLogger.Sync()

	output := buf.String()
	// Verify order is preserved
	for i := 0; i < 9; i++ {
		curr := "seq=" + string(rune('0'+i))
		next := "seq=" + string(rune('0'+i+1))
		if !strings.Contains(output, curr) {
			continue
		}
		currIdx := strings.Index(output, curr)
		nextIdx := strings.Index(output, next)
		if nextIdx != -1 && currIdx > nextIdx {
			t.Errorf("Order not preserved: %s appears after %s", curr, next)
			break
		}
	}
}

// TestAsyncLogger_DroppedLogs tests buffer overflow handling
func TestAsyncLogger_DroppedLogs(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	// Small buffer to force drops
	asyncLogger := Wrap(syncLogger, 5)

	// Log more messages than buffer can hold
	for i := 0; i < 100; i++ {
		asyncLogger.Info("message")
	}

	asyncLogger.Sync()
	metrics := asyncLogger.GetMetrics()

	if metrics.Dropped() == 0 {
		t.Error("Expected some logs to be dropped with small buffer")
	}
}

// TestAsyncLogger_Sync tests that Sync waits for all messages
func TestAsyncLogger_Sync(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	asyncLogger := Wrap(syncLogger, 100)

	// Log messages
	for i := 0; i < 50; i++ {
		asyncLogger.Info("message")
	}

	// Sync should block until all messages are written
	asyncLogger.Sync()

	// All messages should be in buffer
	output := buf.String()
	count := strings.Count(output, "message")
	if count != 50 {
		t.Errorf("Expected 50 messages, got %d", count)
	}
}

// TestAsyncLogger_ConcurrentSafety tests concurrent logging
func TestAsyncLogger_ConcurrentSafety(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	asyncLogger := Wrap(syncLogger, 1000)
	defer asyncLogger.Sync()

	var wg sync.WaitGroup
	numGoroutines := 50
	logsPerGoroutine := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				asyncLogger.Info("concurrent", go_logs.Int("goroutine", id))
			}
		}(i)
	}

	wg.Wait()
	asyncLogger.Sync()

	expected := numGoroutines * logsPerGoroutine
	output := buf.String()
	count := strings.Count(output, "concurrent")
	if count != expected {
		t.Errorf("Expected %d messages, got %d", expected, count)
	}
}

// TestAsyncLogger_WithContext tests context propagation
func TestAsyncLogger_WithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
		go_logs.WithFormatter(go_logs.NewTextFormatter()),
	)

	asyncLogger := Wrap(syncLogger, 100)
	defer asyncLogger.Sync()

	// Use go_logs context API for proper extraction
	ctx := go_logs.WithTraceID(context.Background(), "abc-123")
	asyncLogger.LogCtx(ctx, go_logs.InfoLevel, "with context")
	asyncLogger.Sync()

	if !strings.Contains(buf.String(), "trace_id") {
		t.Error("Expected context fields to be included")
	}
}

// TestAsyncLogger_WithFields tests child logger with fields
func TestAsyncLogger_WithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
		go_logs.WithFormatter(go_logs.NewTextFormatter()),
	)

	asyncLogger := Wrap(syncLogger, 100)
	defer asyncLogger.Sync()

	// Use child logger with pre-pended fields
	childLogger := asyncLogger.With(go_logs.String("request_id", "req-123"))
	childLogger.Info("child message")
	childLogger.Sync() // Sync on child logger

	output := buf.String()
	if !strings.Contains(output, "request_id") {
		t.Errorf("Expected child logger fields to be included, got: %s", output)
	}
	if !strings.Contains(output, "req-123") {
		t.Errorf("Expected field value to be logged, got: %s", output)
	}
}

// TestAsyncLogger_MetricsShared tests that metrics are shared with sync logger
func TestAsyncLogger_MetricsShared(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	asyncLogger := Wrap(syncLogger, 100)
	defer asyncLogger.Sync()

	metrics := asyncLogger.GetMetrics()
	metrics.Reset()

	asyncLogger.Info("message 1")
	asyncLogger.Info("message 2")
	asyncLogger.Sync()

	if metrics.Count(go_logs.InfoLevel) != 2 {
		t.Errorf("Expected 2 info logs in metrics, got %d", metrics.Count(go_logs.InfoLevel))
	}
}

// TestAsyncLogger_LevelFiltering tests that level filtering works
func TestAsyncLogger_LevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	asyncLogger := Wrap(syncLogger, 100)
	defer asyncLogger.Sync()

	metrics := asyncLogger.GetMetrics()
	metrics.Reset()

	asyncLogger.Debug("should be filtered")
	asyncLogger.Info("should be logged")
	asyncLogger.Sync()

	if metrics.Count(go_logs.DebugLevel) != 0 {
		t.Error("Debug logs should be filtered")
	}
	if metrics.Count(go_logs.InfoLevel) != 1 {
		t.Error("Info log should be counted")
	}
}

// TestAsyncLogger_Close tests that Close properly shuts down
func TestAsyncLogger_Close(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	asyncLogger := Wrap(syncLogger, 100)

	asyncLogger.Info("message before close")

	// Close should flush and shutdown
	if err := asyncLogger.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify message was written
	if !strings.Contains(buf.String(), "message before close") {
		t.Error("Expected message to be written before close")
	}
}

// TestAsyncLogger_ShutdownTimeout tests shutdown timeout
func TestAsyncLogger_ShutdownTimeout(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	asyncLogger := WrapWithConfig(syncLogger, Config{
		BufferSize:     100,
		ShutdownTimeout: 100 * time.Millisecond,
	})

	// Log many messages
	for i := 0; i < 50; i++ {
		asyncLogger.Info("message")
	}

	// Should timeout and not block forever
	start := time.Now()
	asyncLogger.Close()
	elapsed := time.Since(start)

	if elapsed > 200*time.Millisecond {
		t.Errorf("Close took too long: %v", elapsed)
	}
}

// TestAsyncLogger_SetLevel tests dynamic level change
func TestAsyncLogger_SetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	syncLogger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
	)

	asyncLogger := Wrap(syncLogger, 100)
	defer asyncLogger.Sync()

	// Initially debug is filtered
	asyncLogger.Debug("should be filtered")
	asyncLogger.Sync()

	if strings.Contains(buf.String(), "should be filtered") {
		t.Error("Debug should be filtered initially")
	}

	// Change level
	asyncLogger.SetLevel(go_logs.DebugLevel)

	// Now debug should pass
	buf.Reset()
	asyncLogger.Debug("should pass now")
	asyncLogger.Sync()

	if !strings.Contains(buf.String(), "should pass now") {
		t.Error("Debug should pass after level change")
	}
}

// TestAsyncLogger_GetLevel tests getting current level
func TestAsyncLogger_GetLevel(t *testing.T) {
	syncLogger, _ := go_logs.New(go_logs.WithLevel(go_logs.WarnLevel))
	asyncLogger := Wrap(syncLogger, 100)
	defer asyncLogger.Sync()

	if asyncLogger.GetLevel() != go_logs.WarnLevel {
		t.Errorf("Expected WarnLevel, got %v", asyncLogger.GetLevel())
	}
}
