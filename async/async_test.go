package async

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	go_logs "github.com/drossan/go_logs/v3"
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
		BufferSize:      100,
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

// safeBuffer is a bytes.Buffer protected by a mutex, so tests can read it while
// the worker goroutine is still writing.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// newBufferedAsync returns an async logger over a text-formatted sync logger that
// writes to a safeBuffer.
func newBufferedAsync(t *testing.T, cfg Config) (*Logger, *safeBuffer) {
	t.Helper()
	buf := &safeBuffer{}
	syncLogger, err := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(buf),
		go_logs.WithFormatter(go_logs.NewTextFormatter()),
	)
	if err != nil {
		t.Fatalf("go_logs.New: %v", err)
	}
	return WrapWithConfig(syncLogger, cfg), buf
}

// recordedCall is one call captured by recordingLogger.
type recordedCall struct {
	fatal  bool
	msg    string
	fields []go_logs.Field
}

// recordingLogger is a go_logs.Logger that records Log and Fatal calls in order
// without terminating the process. When gate is non-nil, every Log call blocks
// until gate is closed, which simulates a slow output.
type recordingLogger struct {
	mu      sync.Mutex
	calls   []recordedCall
	gate    chan struct{}
	metrics *go_logs.Metrics
}

func newRecordingLogger(gate chan struct{}) *recordingLogger {
	return &recordingLogger{gate: gate, metrics: go_logs.NewMetrics()}
}

func (r *recordingLogger) record(c recordedCall) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, c)
}

func (r *recordingLogger) Calls() []recordedCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]recordedCall(nil), r.calls...)
}

func (r *recordingLogger) Log(level go_logs.Level, msg string, fields ...go_logs.Field) {
	if r.gate != nil {
		<-r.gate
	}
	r.record(recordedCall{msg: msg, fields: fields})
}

func (r *recordingLogger) LogCtx(_ context.Context, level go_logs.Level, msg string, fields ...go_logs.Field) {
	r.Log(level, msg, fields...)
}

func (r *recordingLogger) Trace(msg string, fields ...go_logs.Field) {
	r.Log(go_logs.TraceLevel, msg, fields...)
}
func (r *recordingLogger) Debug(msg string, fields ...go_logs.Field) {
	r.Log(go_logs.DebugLevel, msg, fields...)
}
func (r *recordingLogger) Info(msg string, fields ...go_logs.Field) {
	r.Log(go_logs.InfoLevel, msg, fields...)
}
func (r *recordingLogger) Warn(msg string, fields ...go_logs.Field) {
	r.Log(go_logs.WarnLevel, msg, fields...)
}
func (r *recordingLogger) Error(msg string, fields ...go_logs.Field) {
	r.Log(go_logs.ErrorLevel, msg, fields...)
}
func (r *recordingLogger) Fatal(msg string, fields ...go_logs.Field) {
	r.record(recordedCall{fatal: true, msg: msg, fields: fields})
}
func (r *recordingLogger) With(...go_logs.Field) go_logs.Logger { return r }
func (r *recordingLogger) SetLevel(go_logs.Level)               {}
func (r *recordingLogger) GetLevel() go_logs.Level              { return go_logs.TraceLevel }
func (r *recordingLogger) Sync() error                          { return nil }
func (r *recordingLogger) GetMetrics() *go_logs.Metrics         { return r.metrics }

// shortConfig keeps timeouts large enough to detect "waited the full timeout".
var shortConfig = Config{BufferSize: 100, ShutdownTimeout: 5 * time.Second}

// Scenario: Sync desde un hijo espera a que su entrada se escriba.
func TestAsyncChild_SyncWaitsForChildEntry(t *testing.T) {
	asyncLogger, buf := newBufferedAsync(t, shortConfig)
	defer asyncLogger.Close()

	child := asyncLogger.With(go_logs.String("request_id", "req-123"))
	child.Info("child message")
	if err := child.Sync(); err != nil {
		t.Fatalf("child.Sync: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "request_id") || !strings.Contains(out, "req-123") {
		t.Errorf("child entry not written after child.Sync, got: %q", out)
	}
}

// Scenario: Sync desde el padre no espera el timeout tras logs de un hijo.
func TestAsyncChild_ParentSyncDoesNotWaitTimeout(t *testing.T) {
	asyncLogger, buf := newBufferedAsync(t, shortConfig)
	defer asyncLogger.Close()

	child := asyncLogger.With(go_logs.String("request_id", "req-1"))
	for i := 0; i < 10; i++ {
		child.Info("child msg")
	}

	start := time.Now()
	asyncLogger.Sync()
	if elapsed := time.Since(start); elapsed >= time.Second {
		t.Errorf("parent.Sync took %v, want < 1s", elapsed)
	}
	if got := strings.Count(buf.String(), "child msg"); got != 10 {
		t.Errorf("want 10 child messages in buffer, got %d", got)
	}
}

// Scenario: Los campos del hijo no contaminan a otro hijo.
func TestAsyncChild_FieldsDoNotLeakBetweenSiblings(t *testing.T) {
	asyncLogger, buf := newBufferedAsync(t, shortConfig)
	defer asyncLogger.Close()

	base := asyncLogger.With(go_logs.String("app", "x"))
	childA := base.With(go_logs.String("req", "A"))
	childB := base.With(go_logs.String("req", "B"))

	childA.Info("from-A")
	childA.Sync()
	childB.Info("from-B")
	childB.Sync()

	var lineA, lineB string
	for _, line := range strings.Split(buf.String(), "\n") {
		switch {
		case strings.Contains(line, "from-A"):
			lineA = line
		case strings.Contains(line, "from-B"):
			lineB = line
		}
	}
	if !strings.Contains(lineA, "req=A") || strings.Contains(lineA, "req=B") {
		t.Errorf("line A = %q, want req=A and no req=B", lineA)
	}
	if !strings.Contains(lineB, "req=B") || strings.Contains(lineB, "req=A") {
		t.Errorf("line B = %q, want req=B and no req=A", lineB)
	}
	if !strings.Contains(lineA, "app=x") || !strings.Contains(lineB, "app=x") {
		t.Errorf("common field app=x missing: A=%q B=%q", lineA, lineB)
	}
}

// Scenario: Fatal desde un hijo incluye los campos heredados y drena antes.
func TestAsyncChild_FatalDrainsAndIncludesFields(t *testing.T) {
	gate := make(chan struct{})
	rec := newRecordingLogger(gate)
	asyncLogger := WrapWithConfig(rec, shortConfig)
	defer asyncLogger.Close()

	child := asyncLogger.With(go_logs.String("component", "db"))
	child.Info("previous")

	fatalDone := make(chan struct{})
	go func() {
		defer close(fatalDone)
		child.Fatal("boom", go_logs.Int("code", 7))
	}()
	time.Sleep(50 * time.Millisecond) // let Fatal start while "previous" is pending
	close(gate)
	<-fatalDone

	calls := rec.Calls()
	if len(calls) != 2 {
		t.Fatalf("want 2 calls (previous, Fatal), got %+v", calls)
	}
	if calls[0].fatal || calls[0].msg != "previous" {
		t.Errorf("first call = %+v, want Log(previous) before Fatal", calls[0])
	}
	if !calls[1].fatal || calls[1].msg != "boom" {
		t.Fatalf("second call = %+v, want Fatal(boom)", calls[1])
	}
	keys := map[string]bool{}
	for _, f := range calls[1].fields {
		keys[f.Key()] = true
	}
	if !keys["component"] || !keys["code"] {
		t.Errorf("Fatal fields = %+v, want component and code", calls[1].fields)
	}
}

// Scenario: Close en la raíz es idempotente.
func TestAsyncClose_Idempotent(t *testing.T) {
	asyncLogger, buf := newBufferedAsync(t, shortConfig)
	for i := 0; i < 20; i++ {
		asyncLogger.Info("pending msg")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("second Close panicked: %v", r)
		}
	}()
	if err := asyncLogger.Close(); err != nil {
		t.Errorf("first Close: %v", err)
	}
	if err := asyncLogger.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
	if got := strings.Count(buf.String(), "pending msg"); got != 20 {
		t.Errorf("want 20 messages written, got %d", got)
	}
}

// Scenario: Close desde un hijo no apaga el pipeline.
func TestAsyncClose_ChildIsNoOp(t *testing.T) {
	asyncLogger, buf := newBufferedAsync(t, shortConfig)
	defer asyncLogger.Close()

	child := asyncLogger.With(go_logs.String("k", "v")).(*Logger)
	if err := child.Close(); err != nil {
		t.Errorf("child.Close: %v", err)
	}

	asyncLogger.Info("after child close")
	asyncLogger.Sync()
	if !strings.Contains(buf.String(), "after child close") {
		t.Errorf("parent stopped logging after child.Close, got: %q", buf.String())
	}
}

// Scenario: Un hijo creado tras Close no puede volver a loguear.
func TestAsyncClose_ChildCreatedAfterCloseIsDiscarded(t *testing.T) {
	asyncLogger, buf := newBufferedAsync(t, shortConfig)
	asyncLogger.Close()

	child := asyncLogger.With(go_logs.String("k", "v"))
	child.Info("late child")

	start := time.Now()
	child.Sync()
	if elapsed := time.Since(start); elapsed >= time.Second {
		t.Errorf("child.Sync after Close took %v, want < 1s (entry must not be pending)", elapsed)
	}
	if strings.Contains(buf.String(), "late child") {
		t.Errorf("entry from child created after Close was written: %q", buf.String())
	}
}

// Scenario: Close en la raíz concurrente con logs de un hijo no produce pánico ni carreras.
func TestAsyncClose_ConcurrentWithChildLogging(t *testing.T) {
	cfg := Config{BufferSize: 100, ShutdownTimeout: 500 * time.Millisecond}
	asyncLogger, _ := newBufferedAsync(t, cfg)
	child := asyncLogger.With(go_logs.String("k", "v"))

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				child.Info("spam")
			}
		}
	}()

	time.Sleep(20 * time.Millisecond)
	start := time.Now()
	err := asyncLogger.Close()
	elapsed := time.Since(start)
	close(stop)
	wg.Wait()

	if err != nil {
		t.Errorf("Close: %v", err)
	}
	if elapsed > cfg.ShutdownTimeout {
		t.Errorf("Close took %v, want <= %v", elapsed, cfg.ShutdownTimeout)
	}
}

// Scenario: ShutdownTimeout no positivo cae al valor por defecto.
func TestAsyncConfig_NonPositiveShutdownTimeoutDefaults(t *testing.T) {
	for _, timeout := range []time.Duration{0, -time.Second} {
		asyncLogger, _ := newBufferedAsync(t, Config{BufferSize: 10, ShutdownTimeout: timeout})
		if got := effectiveShutdownTimeout(asyncLogger); got != 5*time.Second {
			t.Errorf("ShutdownTimeout(%v) → %v, want 5s", timeout, got)
		}
		asyncLogger.Close()
	}
}

// Scenario: Registrar tras Close en la raíz se descarta sin pánico.
func TestAsyncClose_LogAfterCloseIsDiscarded(t *testing.T) {
	asyncLogger, buf := newBufferedAsync(t, shortConfig)
	asyncLogger.Close()

	asyncLogger.Info("late")
	asyncLogger.Sync()
	if strings.Contains(buf.String(), "late") {
		t.Errorf("entry logged after Close was written: %q", buf.String())
	}
}

// assertDropsWithoutBlocking logs 10 messages through logger (whose root has a
// buffer of 1 and a blocked sync logger) and checks drops are counted without
// blocking the caller.
func assertDropsWithoutBlocking(t *testing.T, logger go_logs.Logger, rec *recordingLogger) {
	t.Helper()
	start := time.Now()
	for i := 0; i < 10; i++ {
		logger.Info("flood")
	}
	if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
		t.Errorf("logging blocked the caller for %v", elapsed)
	}
	if rec.metrics.Dropped() == 0 {
		t.Error("want dropped > 0 with a full buffer")
	}
}

// Scenario: Buffer lleno descarta y contabiliza.
func TestAsyncBufferFull_DropsAndCounts(t *testing.T) {
	gate := make(chan struct{})
	rec := newRecordingLogger(gate)
	asyncLogger := WrapWithConfig(rec, Config{BufferSize: 1, ShutdownTimeout: 100 * time.Millisecond})
	defer func() { close(gate); asyncLogger.Close() }()

	assertDropsWithoutBlocking(t, asyncLogger, rec)
}

// Scenario: Buffer lleno descarta y contabiliza también cuando loguea un hijo.
func TestAsyncBufferFull_ChildDropsAndCounts(t *testing.T) {
	gate := make(chan struct{})
	rec := newRecordingLogger(gate)
	asyncLogger := WrapWithConfig(rec, Config{BufferSize: 1, ShutdownTimeout: 100 * time.Millisecond})
	defer func() { close(gate); asyncLogger.Close() }()

	assertDropsWithoutBlocking(t, asyncLogger.With(go_logs.String("k", "v")), rec)
}

// Close on a child logger must not stop the worker even if called several times.
func TestAsyncClose_ChildRepeatedCloseIsSafe(t *testing.T) {
	asyncLogger, _ := newBufferedAsync(t, shortConfig)
	defer asyncLogger.Close()
	child := asyncLogger.With().(*Logger)
	for i := 0; i < 3; i++ {
		if err := child.Close(); err != nil {
			t.Errorf("child.Close #%d: %v", i, err)
		}
	}
}

// effectiveShutdownTimeout exposes the timeout the logger will actually use.
func effectiveShutdownTimeout(l *Logger) time.Duration {
	return l.core.config.ShutdownTimeout
}
