package go_logs

import (
	"sync"
	"testing"
)

// TestMetrics_NewLogger_HasMetrics tests that a new logger has metrics enabled by default
func TestMetrics_NewLogger_HasMetrics(t *testing.T) {
	logger, err := New()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Logger should implement MetricsGetter interface
	mg, ok := logger.(interface{ GetMetrics() *Metrics })
	if !ok {
		t.Fatal("Logger should implement GetMetrics()")
	}

	metrics := mg.GetMetrics()
	if metrics == nil {
		t.Fatal("Metrics should not be nil")
	}
}

// TestMetrics_CountByLevel tests that logs are counted by level
func TestMetrics_CountByLevel(t *testing.T) {
	logger, _ := New(WithLevel(DebugLevel))
	mg := logger.(interface{ GetMetrics() *Metrics })
	metrics := mg.GetMetrics()

	// Reset metrics
	metrics.Reset()

	// Log at different levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Info("another info")
	logger.Warn("warning")
	logger.Error("error")

	// Check counts
	if metrics.Count(DebugLevel) != 1 {
		t.Errorf("Expected 1 debug logs, got %d", metrics.Count(DebugLevel))
	}
	if metrics.Count(InfoLevel) != 2 {
		t.Errorf("Expected 2 info logs, got %d", metrics.Count(InfoLevel))
	}
	if metrics.Count(WarnLevel) != 1 {
		t.Errorf("Expected 1 warn logs, got %d", metrics.Count(WarnLevel))
	}
	if metrics.Count(ErrorLevel) != 1 {
		t.Errorf("Expected 1 error logs, got %d", metrics.Count(ErrorLevel))
	}
}

// TestMetrics_FilteredLogsNotCounted tests that filtered logs are not counted
func TestMetrics_FilteredLogsNotCounted(t *testing.T) {
	logger, _ := New(WithLevel(InfoLevel))
	mg := logger.(interface{ GetMetrics() *Metrics })
	metrics := mg.GetMetrics()

	metrics.Reset()

	// These should be filtered
	logger.Trace("trace message")
	logger.Debug("debug message")

	// These should be counted
	logger.Info("info message")

	if metrics.Count(TraceLevel) != 0 {
		t.Errorf("Filtered trace logs should not be counted, got %d", metrics.Count(TraceLevel))
	}
	if metrics.Count(DebugLevel) != 0 {
		t.Errorf("Filtered debug logs should not be counted, got %d", metrics.Count(DebugLevel))
	}
	if metrics.Count(InfoLevel) != 1 {
		t.Errorf("Expected 1 info log, got %d", metrics.Count(InfoLevel))
	}
}

// TestMetrics_TotalCount tests total log count
func TestMetrics_TotalCount(t *testing.T) {
	logger, _ := New(WithLevel(DebugLevel))
	mg := logger.(interface{ GetMetrics() *Metrics })
	metrics := mg.GetMetrics()

	metrics.Reset()

	logger.Debug("1")
	logger.Info("2")
	logger.Warn("3")

	if metrics.Total() != 3 {
		t.Errorf("Expected total 3 logs, got %d", metrics.Total())
	}
}

// TestMetrics_ConcurrentSafety tests that metrics are thread-safe
func TestMetrics_ConcurrentSafety(t *testing.T) {
	logger, _ := New(WithLevel(DebugLevel))
	mg := logger.(interface{ GetMetrics() *Metrics })
	metrics := mg.GetMetrics()

	metrics.Reset()

	var wg sync.WaitGroup
	numGoroutines := 100
	logsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Info("concurrent log")
			}
		}()
	}

	wg.Wait()

	expected := int64(numGoroutines * logsPerGoroutine)
	if metrics.Count(InfoLevel) != expected {
		t.Errorf("Expected %d info logs, got %d", expected, metrics.Count(InfoLevel))
	}
}

// TestMetrics_Reset tests that Reset clears all counters
func TestMetrics_Reset(t *testing.T) {
	logger, _ := New(WithLevel(DebugLevel))
	mg := logger.(interface{ GetMetrics() *Metrics })
	metrics := mg.GetMetrics()

	logger.Info("message 1")
	logger.Info("message 2")

	if metrics.Count(InfoLevel) != 2 {
		t.Fatalf("Expected 2 logs before reset, got %d", metrics.Count(InfoLevel))
	}

	metrics.Reset()

	if metrics.Count(InfoLevel) != 0 {
		t.Errorf("Expected 0 logs after reset, got %d", metrics.Count(InfoLevel))
	}
	if metrics.Total() != 0 {
		t.Errorf("Expected total 0 after reset, got %d", metrics.Total())
	}
}

// TestMetrics_ChildLoggerInheritsMetrics tests that child loggers share metrics with parent
func TestMetrics_ChildLoggerInheritsMetrics(t *testing.T) {
	logger, _ := New(WithLevel(DebugLevel))
	childLogger := logger.With(String("request_id", "123"))

	parentMg := logger.(interface{ GetMetrics() *Metrics })
	childMg := childLogger.(interface{ GetMetrics() *Metrics })

	parentMetrics := parentMg.GetMetrics()
	childMetrics := childMg.GetMetrics()

	parentMetrics.Reset()

	// Log from child
	childLogger.Info("child log")

	// Parent should see it
	if parentMetrics.Count(InfoLevel) != 1 {
		t.Errorf("Parent should see child logs, got %d", parentMetrics.Count(InfoLevel))
	}

	// Child should see it too (same metrics instance)
	if childMetrics.Count(InfoLevel) != 1 {
		t.Errorf("Child metrics should match parent, got %d", childMetrics.Count(InfoLevel))
	}
}

// TestMetrics_AllLevels tests all log levels are tracked
func TestMetrics_AllLevels(t *testing.T) {
	logger, _ := New(WithLevel(TraceLevel))
	mg := logger.(interface{ GetMetrics() *Metrics })
	metrics := mg.GetMetrics()

	metrics.Reset()

	logger.Trace("trace")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	levels := []Level{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel}
	for _, level := range levels {
		if metrics.Count(level) != 1 {
			t.Errorf("Expected 1 log for level %v, got %d", level, metrics.Count(level))
		}
	}

	if metrics.Total() != 5 {
		t.Errorf("Expected total 5 logs, got %d", metrics.Total())
	}
}

// TestMetrics_Snapshot tests getting a snapshot of metrics
func TestMetrics_Snapshot(t *testing.T) {
	logger, _ := New(WithLevel(DebugLevel))
	mg := logger.(interface{ GetMetrics() *Metrics })
	metrics := mg.GetMetrics()

	metrics.Reset()

	logger.Info("message 1")
	logger.Info("message 2")
	logger.Error("error 1")

	snapshot := metrics.Snapshot()

	if snapshot.Total != 3 {
		t.Errorf("Expected snapshot total 3, got %d", snapshot.Total)
	}
	if snapshot.ByLevel[InfoLevel] != 2 {
		t.Errorf("Expected snapshot 2 info, got %d", snapshot.ByLevel[InfoLevel])
	}
	if snapshot.ByLevel[ErrorLevel] != 1 {
		t.Errorf("Expected snapshot 1 error, got %d", snapshot.ByLevel[ErrorLevel])
	}
}
