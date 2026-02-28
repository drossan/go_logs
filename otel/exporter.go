package otel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/drossan/go_logs"
)

// OTLPExporter sends logs to an OpenTelemetry collector via OTLP protocol
type OTLPExporter struct {
	endpoint    string
	headers     map[string]string
	client      *http.Client
	mu          sync.Mutex
	pending     []go_logs.Entry
	maxPending  int
	flushInterval time.Duration
	stopCh      chan struct{}
	flushCh     chan struct{}
}

// OTLPConfig holds configuration for the OTLP exporter
type OTLPConfig struct {
	// Endpoint is the OTLP collector URL (e.g., "http://localhost:4318/v1/logs")
	Endpoint string

	// Headers are additional HTTP headers to send
	Headers map[string]string

	// MaxPending is the maximum number of entries to buffer before flushing
	MaxPending int

	// FlushInterval is how often to flush pending entries
	FlushInterval time.Duration

	// Timeout is the HTTP client timeout
	Timeout time.Duration
}

// NewOTLPExporter creates a new OTLP exporter with default configuration
func NewOTLPExporter(endpoint string) *OTLPExporter {
	return NewOTLPExporterWithConfig(OTLPConfig{
		Endpoint:      endpoint,
		MaxPending:    100,
		FlushInterval: 5 * time.Second,
		Timeout:       30 * time.Second,
	})
}

// NewOTLPExporterWithConfig creates a new OTLP exporter with custom configuration
func NewOTLPExporterWithConfig(config OTLPConfig) *OTLPExporter {
	if config.MaxPending <= 0 {
		config.MaxPending = 100
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	e := &OTLPExporter{
		endpoint:      config.Endpoint,
		headers:       config.Headers,
		client:        &http.Client{Timeout: config.Timeout},
		maxPending:     config.MaxPending,
		flushInterval:  config.FlushInterval,
		pending:        make([]go_logs.Entry, 0, config.MaxPending),
		stopCh:         make(chan struct{}),
		flushCh:        make(chan struct{}, 1),
	}

	// Start background flusher
	go e.backgroundFlush()

	return e
}

// Export exports a log entry to the OTLP collector
func (e *OTLPExporter) Export(entry *go_logs.Entry) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.pending = append(e.pending, *entry)

	// Flush if buffer is full
	if len(e.pending) >= e.maxPending {
		go e.Flush()
	}

	return nil
}

// Flush sends all pending entries to the collector
func (e *OTLPExporter) Flush() error {
	e.mu.Lock()
	entries := e.pending
	e.pending = make([]go_logs.Entry, 0, e.maxPending)
	e.mu.Unlock()

	if len(entries) == 0{
		return nil
	}

	// Convert entries to OTLP format
	body, err := e.marshalOTLP(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal OTLP: %w", err)
	}

	// Send to collector
	req, err := http.NewRequest("POST", e.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range e.headers {
		req.Header.Set(k, v)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("collector returned status %d", resp.StatusCode)
	}

	return nil
}

// marshalOTLP converts entries to OTLP JSON format
func (e *OTLPExporter) marshalOTLP(entries []go_logs.Entry) ([]byte, error) {
	// Simplified OTLP-style format
	// In production, use go.opentelemetry.io/otel for proper OTLP encoding
	logs := make([]map[string]interface{}, len(entries))

	for i, entry := range entries {
		attrs := map[string]interface{}{
		"severity": entry.Level.String(),
		}

		// Add fields as attributes
		for _, f := range entry.Fields {
			attrs[f.Key()] = f.Value()
		}

		// Add caller info if present
		if entry.Caller != nil {
			attrs["code.filepath"] = entry.Caller.File
			attrs["code.lineno"] = entry.Caller.Line
			attrs["code.function"] = entry.Caller.Func
		}

		logs[i] = map[string]interface{}{
			"observedTimestamp": entry.Timestamp.Format(time.RFC3339Nano),
			"body":               entry.Message,
			"attributes":         attrs,
		}
	}

	return json.Marshal(map[string]interface{}{
		"resourceLogs": logs,
	})
}

// backgroundFlush periodically flushes pending entries
func (e *OTLPExporter) backgroundFlush() {
	ticker := time.NewTicker(e.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.Flush()
		case <-e.flushCh:
			e.Flush()
		case <-e.stopCh:
			return
		}
	}
}

// Close stops the exporter and flushes remaining entries
func (e *OTLPExporter) Close() error {
	close(e.stopCh)
	close(e.flushCh)
	return e.Flush()
}

// OTLPHook implements go_logs.Hook interface
type OTLPHook struct {
	exporter *OTLPExporter
	minLevel go_logs.Level
}

// NewOTLPHook creates a new OTLP hook
func NewOTLPHook(endpoint string) *OTLPHook {
	return &OTLPHook{
		exporter: NewOTLPExporter(endpoint),
		minLevel: go_logs.InfoLevel,
	}
}

// NewOTLPHookWithExporter creates a hook with a custom exporter
func NewOTLPHookWithExporter(exporter *OTLPExporter, minLevel go_logs.Level) *OTLPHook {
	return &OTLPHook{
		exporter: exporter,
		minLevel: minLevel,
	}
}

// Run implements go_logs.Hook interface
func (h *OTLPHook) Run(entry *go_logs.Entry) error {
	if entry.Level < h.minLevel {
		return nil
	}
	return h.exporter.Export(entry)
}

// SetLevel sets the minimum level for this hook
func (h *OTLPHook) SetLevel(level go_logs.Level) {
	h.minLevel = level
}

// Close closes the hook and its exporter
func (h *OTLPHook) Close() error {
	return h.exporter.Close()
}
