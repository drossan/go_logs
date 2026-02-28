package go_logs

import (
	"bytes"
	"context"
	"sync"
)

// CaptureBuffer is a thread-safe buffer for capturing log output in tests
type CaptureBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// NewCaptureBuffer creates a new CaptureBuffer
func NewCaptureBuffer() *CaptureBuffer {
	return &CaptureBuffer{}
}

// Write implements io.Writer
func (cb *CaptureBuffer) Write(p []byte) (n int, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.buf.Write(p)
}

// String returns the contents as a string
func (cb *CaptureBuffer) String() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.buf.String()
}

// Bytes returns the contents as a byte slice
func (cb *CaptureBuffer) Bytes() []byte {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.buf.Bytes()
}

// Reset clears the buffer
func (cb *CaptureBuffer) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.buf.Reset()
}

// Contains checks if the buffer contains the substring
func (cb *CaptureBuffer) Contains(substring string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return bytes.Contains(cb.buf.Bytes(), []byte(substring))
}

// ContainsAll checks if the buffer contains all substrings
func (cb *CaptureBuffer) ContainsAll(substrings ...string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	content := cb.buf.Bytes()
	for _, s := range substrings {
		if !bytes.Contains(content, []byte(s)) {
			return false
		}
	}
	return true
}

// ContainsAny checks if the buffer contains any of the substrings
func (cb *CaptureBuffer) ContainsAny(substrings ...string) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	content := cb.buf.Bytes()
	for _, s := range substrings {
		if bytes.Contains(content, []byte(s)) {
			return true
		}
	}
	return false
}

// Lines returns the contents as a slice of lines
func (cb *CaptureBuffer) Lines() []string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	content := cb.buf.String()
	if content == "" {
		return nil
	}
	return splitLines(content)
}

// LastLine returns the last line written
func (cb *CaptureBuffer) LastLine() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	lines := splitLines(cb.buf.String())
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

// LineCount returns the number of lines
func (cb *CaptureBuffer) LineCount() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return len(splitLines(cb.buf.String()))
}

// splitLines is a helper to split content into lines
func splitLines(content string) []string {
	if content == "" {
		return nil
	}
	// Remove trailing newline if present
	if len(content) > 0 && content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}
	return split(content, '\n')
}

// split splits a string by a delimiter
func split(s string, delim byte) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == delim {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

// MockLogger is a mock logger for testing
type MockLogger struct {
	mu       sync.Mutex
	entries  []MockEntry
	level    Level
	lastErr  error
}

// MockEntry represents a captured log entry
type MockEntry struct {
	Level   Level
	Message string
	Fields  []Field
}

// NewMockLogger creates a new MockLogger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		level:   InfoLevel,
		entries: make([]MockEntry, 0),
	}
}

// Log implements Logger.Log
func (m *MockLogger) Log(level Level, msg string, fields ...Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !level.shouldLog(m.level) {
		return
	}
	m.entries = append(m.entries, MockEntry{
		Level:   level,
		Message: msg,
		Fields:  fields,
	})
}

// LogCtx implements Logger.LogCtx
func (m *MockLogger) LogCtx(ctx context.Context, level Level, msg string, fields ...Field) {
	m.Log(level, msg, fields...)
}

// Trace implements Logger.Trace
func (m *MockLogger) Trace(msg string, fields ...Field) {
	m.Log(TraceLevel, msg, fields...)
}

// Debug implements Logger.Debug
func (m *MockLogger) Debug(msg string, fields ...Field) {
	m.Log(DebugLevel, msg, fields...)
}

// Info implements Logger.Info
func (m *MockLogger) Info(msg string, fields ...Field) {
	m.Log(InfoLevel, msg, fields...)
}

// Warn implements Logger.Warn
func (m *MockLogger) Warn(msg string, fields ...Field) {
	m.Log(WarnLevel, msg, fields...)
}

// Error implements Logger.Error
func (m *MockLogger) Error(msg string, fields ...Field) {
	m.Log(ErrorLevel, msg, fields...)
}

// Fatal implements Logger.Fatal (does not exit in mock)
func (m *MockLogger) Fatal(msg string, fields ...Field) {
	m.Log(FatalLevel, msg, fields...)
}

// With implements Logger.With
func (m *MockLogger) With(fields ...Field) Logger {
	return m // Return same mock for simplicity
}

// SetLevel implements Logger.SetLevel
func (m *MockLogger) SetLevel(level Level) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.level = level
}

// GetLevel implements Logger.GetLevel
func (m *MockLogger) GetLevel() Level {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.level
}

// Sync implements Logger.Sync
func (m *MockLogger) Sync() error {
	return nil
}

// Mock-specific methods

// Entries returns all captured entries
func (m *MockLogger) Entries() []MockEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]MockEntry, len(m.entries))
	copy(result, m.entries)
	return result
}

// LastEntry returns the last captured entry
func (m *MockLogger) LastEntry() *MockEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entries) == 0 {
		return nil
	}
	entry := m.entries[len(m.entries)-1]
	return &entry
}

// Reset clears all captured entries
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = make([]MockEntry, 0)
}

// Count returns the number of captured entries
func (m *MockLogger) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

// HasMessage checks if any entry contains the message
func (m *MockLogger) HasMessage(msg string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.entries {
		if e.Message == msg {
			return true
		}
	}
	return false
}

// HasLevel checks if any entry has the specified level
func (m *MockLogger) HasLevel(level Level) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.entries {
		if e.Level == level {
			return true
		}
	}
	return false
}
