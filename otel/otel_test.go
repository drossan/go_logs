package otel

import (
	"testing"
	"time"

	"github.com/drossan/go_logs"
)

// TestOTLPExporter_New tests exporter creation
func TestOTLPExporter_New(t *testing.T) {
	exporter := NewOTLPExporter("http://localhost:4318/v1/logs")

	if exporter == nil {
		t.Fatal("Exporter should not be nil")
	}

	if exporter.endpoint != "http://localhost:4318/v1/logs" {
		t.Errorf("Expected endpoint, got %s", exporter.endpoint)
	}

	exporter.Close()
}

// TestOTLPExporter_Export tests exporting entries
func TestOTLPExporter_Export(t *testing.T) {
	exporter := NewOTLPExporter("http://localhost:4318/v1/logs")
	defer exporter.Close()

	entry := &go_logs.Entry{
		Level:     go_logs.InfoLevel,
		Message:   "test message",
		Timestamp: time.Now(),
		Fields:    []go_logs.Field{go_logs.String("key", "value")},
	}

	err := exporter.Export(entry)
	if err != nil {
		t.Errorf("Export should not fail: %v", err)
	}

	// Check pending
	if len(exporter.pending) != 1 {
		t.Errorf("Expected 1 pending, got %d", len(exporter.pending))
	}
}

// TestOTLPExporter_MarshalOTLP tests OTLP marshaling
func TestOTLPExporter_MarshalOTLP(t *testing.T) {
	exporter := NewOTLPExporter("http://localhost:4318/v1/logs")
	defer exporter.Close()

	entries := []go_logs.Entry{
		{
			Level:     go_logs.InfoLevel,
			Message:   "test message",
			Timestamp: time.Now(),
			Fields:    []go_logs.Field{go_logs.String("key", "value")},
		},
	}

	body, err := exporter.marshalOTLP(entries)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if len(body) == 0 {
		t.Error("Body should not be empty")
	}

	// Should contain the message
	if string(body) == "" || len(body) < 10 {
		t.Errorf("Body seems too short: %s", string(body))
	}
}

// TestOTLPHook_New tests hook creation
func TestOTLPHook_New(t *testing.T) {
	hook := NewOTLPHook("http://localhost:4318/v1/logs")
	defer hook.Close()

	if hook == nil {
		t.Fatal("Hook should not be nil")
	}

	if hook.minLevel != go_logs.InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", hook.minLevel)
	}
}

// TestOTLPHook_Run tests hook Run method
func TestOTLPHook_Run(t *testing.T) {
	hook := NewOTLPHook("http://localhost:4318/v1/logs")
	defer hook.Close()

	// Info level should pass
	entry := &go_logs.Entry{
		Level:     go_logs.InfoLevel,
		Message:   "test",
		Timestamp: time.Now(),
	}

	err := hook.Run(entry)
	if err != nil {
		t.Errorf("Run should not fail: %v", err)
	}

	// Debug level should be filtered (no error, just returns nil)
	entry2 := &go_logs.Entry{
		Level:     go_logs.DebugLevel,
		Message:   "filtered",
		Timestamp: time.Now(),
	}

	err = hook.Run(entry2)
	if err != nil {
		t.Errorf("Run should not fail for filtered: %v", err)
	}
}

// TestOTLPHook_SetLevel tests level setting
func TestOTLPHook_SetLevel(t *testing.T) {
	hook := NewOTLPHook("http://localhost:4318/v1/logs")
	defer hook.Close()

	hook.SetLevel(go_logs.ErrorLevel)

	if hook.minLevel != go_logs.ErrorLevel {
		t.Errorf("Expected ErrorLevel, got %v", hook.minLevel)
	}
}

// TestOTLPConfig tests configuration
func TestOTLPConfig(t *testing.T) {
	exporter := NewOTLPExporterWithConfig(OTLPConfig{
		Endpoint:      "http://custom:4318",
		MaxPending:    50,
		FlushInterval: 2 * time.Second,
		Timeout:       10 * time.Second,
	})
	defer exporter.Close()

	if exporter.endpoint != "http://custom:4318" {
		t.Errorf("Expected custom endpoint, got %s", exporter.endpoint)
	}

	if exporter.maxPending != 50 {
		t.Errorf("Expected 50 maxPending, got %d", exporter.maxPending)
	}
}
