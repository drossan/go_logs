package go_logs

import (
	"strings"
	"testing"
	"time"
)

// TestSampler_Basic tests basic sampling functionality
func TestSampler_Basic(t *testing.T) {
	sampler := NewSampler(3, time.Minute)

	// First 3 messages should be allowed
	for i := 0; i < 3; i++ {
		if !sampler.Allow() {
			t.Errorf("Message %d should be allowed", i)
		}
	}

	// 4th message should be dropped
	if sampler.Allow() {
		t.Error("Message 4 should be dropped")
	}
}

// TestSampler_Reset tests that counter resets after interval
func TestSampler_Reset(t *testing.T) {
	// Very short interval for testing
	sampler := NewSampler(2, 10*time.Millisecond)

	// Use up the threshold
	sampler.Allow()
	sampler.Allow()

	// Should be blocked now
	if sampler.Allow() {
		t.Error("Should be blocked")
	}

	// Wait for interval to pass
	time.Sleep(15 * time.Millisecond)

	// Should be allowed again
	if !sampler.Allow() {
		t.Error("Should be allowed after reset")
	}
}

// TestSampler_GetCount tests count tracking
func TestSampler_GetCount(t *testing.T) {
	sampler := NewSampler(10, time.Minute)

	if sampler.GetCount() != 0 {
		t.Error("Initial count should be 0")
	}

	sampler.Allow()
	if sampler.GetCount() != 1 {
		t.Error("Count should be 1")
	}

	sampler.Allow()
	if sampler.GetCount() != 2 {
		t.Error("Count should be 2")
	}
}

// TestSampler_Reset_Method tests manual reset
func TestSampler_Reset_Method(t *testing.T) {
	sampler := NewSampler(10, time.Minute)

	sampler.Allow()
	sampler.Allow()
	sampler.Allow()

	if sampler.GetCount() != 3 {
		t.Errorf("Count should be 3, got %d", sampler.GetCount())
	}

	sampler.Reset()

	if sampler.GetCount() != 0 {
		t.Error("Count should be 0 after reset")
	}
}

// TestSamplingWriter_Basic tests SamplingWriter
func TestSamplingWriter_Basic(t *testing.T) {
	buf := NewCaptureBuffer()
	sw := NewSamplingWriter(buf, 3, time.Minute)

	// First 3 writes should succeed
	for i := 0; i < 3; i++ {
		n, err := sw.Write([]byte("test\n"))
		if err != nil {
			t.Errorf("Write %d failed: %v", i, err)
		}
		if n != 5 {
			t.Errorf("Write %d: expected 5 bytes, got %d", i, n)
		}
	}

	// 4th write should be dropped (returns len but doesn't write)
	n, _ := sw.Write([]byte("dropped\n"))
	if n != 8 {
		t.Errorf("Dropped write should return 8, got %d", n)
	}

	// Buffer should only have 3 lines
	if buf.LineCount() != 3 {
		t.Errorf("Expected 3 lines, got %d", buf.LineCount())
	}
}

// TestSamplingWriter_Callback tests drop callback
func TestSamplingWriter_Callback(t *testing.T) {
	buf := NewCaptureBuffer()
	droppedCount := 0

	sw := NewSamplingWriterWithCallback(buf, 2, time.Minute, func(dropped int) {
		droppedCount += dropped
	})

	// Use up threshold
	sw.Write([]byte("test1\n"))
	sw.Write([]byte("test2\n"))

	// This should trigger callback
	sw.Write([]byte("dropped\n"))

	if droppedCount != 1 {
		t.Errorf("Expected 1 dropped, got %d", droppedCount)
	}
}

// TestGlobalLogger_SetDefault tests SetDefault
func TestGlobalLogger_SetDefault(t *testing.T) {
	// Create a mock logger
	mock := NewMockLogger()

	// Set as default
	SetDefault(mock)

	// Use global functions
	GlobalInfo("test info")
	GlobalError("test error")

	// Verify entries
	entries := mock.Entries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Reset for other tests
	mock.Reset()
}

// TestGlobalLogger_Default tests Default() function
func TestGlobalLogger_Default(t *testing.T) {
	// Clear default
	defaultLogger = nil

	// Should return a valid logger
	logger := Default()
	if logger == nil {
		t.Error("Default() should return a logger")
	}
}

// TestGlobalLogger_SetDefaultLevel tests SetDefaultLevel
func TestGlobalLogger_SetDefaultLevel(t *testing.T) {
	mock := NewMockLogger()
	SetDefault(mock)

	// Set level to Error
	SetDefaultLevel(ErrorLevel)

	// Info should be filtered
	GlobalInfo("should be filtered")
	if mock.Count() != 0 {
		t.Error("Info should be filtered at Error level")
	}

	// Error should pass
	GlobalError("should pass")
	if mock.Count() != 1 {
		t.Errorf("Expected 1 entry, got %d", mock.Count())
	}

	// Cleanup
	mock.Reset()
}

// TestCaptureBuffer_Basic tests basic buffer operations
func TestCaptureBuffer_Basic(t *testing.T) {
	buf := NewCaptureBuffer()

	// Write some data
	buf.Write([]byte("line1\n"))
	buf.Write([]byte("line2\n"))
	buf.Write([]byte("line3\n"))

	// Check string
	if !strings.Contains(buf.String(), "line1") {
		t.Error("Should contain line1")
	}

	// Check lines
	lines := buf.Lines()
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// Check last line
	if buf.LastLine() != "line3" {
		t.Errorf("Expected 'line3', got '%s'", buf.LastLine())
	}
}

// TestCaptureBuffer_Contains tests Contains method
func TestCaptureBuffer_Contains(t *testing.T) {
	buf := NewCaptureBuffer()
	buf.Write([]byte("hello world foo bar"))

	if !buf.Contains("world") {
		t.Error("Should contain 'world'")
	}

	if buf.Contains("missing") {
		t.Error("Should not contain 'missing'")
	}
}

// TestCaptureBuffer_ContainsAll tests ContainsAll method
func TestCaptureBuffer_ContainsAll(t *testing.T) {
	buf := NewCaptureBuffer()
	buf.Write([]byte("hello world foo bar"))

	if !buf.ContainsAll("hello", "world", "foo") {
		t.Error("Should contain all substrings")
	}

	if buf.ContainsAll("hello", "missing") {
		t.Error("Should not contain 'missing'")
	}
}

// TestCaptureBuffer_Reset tests Reset method
func TestCaptureBuffer_Reset(t *testing.T) {
	buf := NewCaptureBuffer()
	buf.Write([]byte("some data"))

	if buf.String() == "" {
		t.Error("Should have data")
	}

	buf.Reset()

	if buf.String() != "" {
		t.Error("Should be empty after reset")
	}
}

// TestMockLogger_Basic tests basic mock logger functionality
func TestMockLogger_Basic(t *testing.T) {
	mock := NewMockLogger()
	mock.SetLevel(DebugLevel) // Allow all levels

	// Log some messages
	mock.Info("info message")
	mock.Error("error message")
	mock.Debug("debug message")

	// Check count
	if mock.Count() != 3 {
		t.Errorf("Expected 3 entries, got %d", mock.Count())
	}

	// Check last entry
	last := mock.LastEntry()
	if last == nil || last.Message != "debug message" {
		t.Error("Last entry should be 'debug message'")
	}

	// Check has message
	if !mock.HasMessage("info message") {
		t.Error("Should have 'info message'")
	}

	// Check has level
	if !mock.HasLevel(ErrorLevel) {
		t.Error("Should have Error level entry")
	}
}

// TestMockLogger_LevelFiltering tests level filtering in mock
func TestMockLogger_LevelFiltering(t *testing.T) {
	mock := NewMockLogger()
	mock.SetLevel(WarnLevel)

	// These should be filtered
	mock.Debug("debug")
	mock.Info("info")

	if mock.Count() != 0 {
		t.Error("Debug and Info should be filtered at WarnLevel")
	}

	// These should pass
	mock.Warn("warn")
	mock.Error("error")

	if mock.Count() != 2 {
		t.Errorf("Expected 2 entries, got %d", mock.Count())
	}
}

// TestMockLogger_Reset tests mock reset
func TestMockLogger_Reset(t *testing.T) {
	mock := NewMockLogger()

	mock.Info("message1")
	mock.Info("message2")

	if mock.Count() != 2 {
		t.Errorf("Expected 2 entries, got %d", mock.Count())
	}

	mock.Reset()

	if mock.Count() != 0 {
		t.Error("Should be empty after reset")
	}
}

// TestMockLogger_WithFields tests mock with fields
func TestMockLogger_WithFields(t *testing.T) {
	mock := NewMockLogger()

	mock.Info("test", String("key", "value"), Int("count", 42))

	entries := mock.Entries()
	if len(entries) != 1 {
		t.Fatal("Expected 1 entry")
	}

	entry := entries[0]
	if entry.Message != "test" {
		t.Errorf("Expected 'test', got '%s'", entry.Message)
	}

	if len(entry.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(entry.Fields))
	}
}
