package go_logs

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// TestTextFormatter_Format tests the TextFormatter.Format method.
func TestTextFormatter_Format(t *testing.T) {
	formatter := NewTextFormatter()

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "test message",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields:    []Field{String("key", "value")},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	outputStr := string(output)

	// Check that output contains expected components
	if !strings.Contains(outputStr, "test message") {
		t.Errorf("Format() output should contain message 'test message', got %s", outputStr)
	}

	if !strings.Contains(outputStr, "key=value") {
		t.Errorf("Format() output should contain field 'key=value', got %s", outputStr)
	}

	if !strings.Contains(outputStr, "INFO") {
		t.Errorf("Format() output should contain level 'INFO', got %s", outputStr)
	}

	if !strings.Contains(outputStr, "2026/02/28") {
		t.Errorf("Format() output should contain timestamp '2026/02/28', got %s", outputStr)
	}

	// Check trailing newline
	if !strings.HasSuffix(outputStr, "\n") {
		t.Errorf("Format() output should end with newline, got %s", outputStr)
	}
}

// TestTextFormatter_FormatWithMultipleFields tests formatting with multiple fields.
func TestTextFormatter_FormatWithMultipleFields(t *testing.T) {
	formatter := NewTextFormatter()

	entry := &Entry{
		Level:     ErrorLevel,
		Message:   "database error",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields: []Field{
			String("host", "localhost"),
			Int("port", 5432),
			String("user", "admin"),
		},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	outputStr := string(output)

	// Check all fields are present
	if !strings.Contains(outputStr, "host=localhost") {
		t.Errorf("Format() output should contain 'host=localhost', got %s", outputStr)
	}

	if !strings.Contains(outputStr, "port=5432") {
		t.Errorf("Format() output should contain 'port=5432', got %s", outputStr)
	}

	if !strings.Contains(outputStr, "user=admin") {
		t.Errorf("Format() output should contain 'user=admin', got %s", outputStr)
	}
}

// TestTextFormatter_FormatWithSpecialCharacters tests formatting fields with spaces and special chars.
func TestTextFormatter_FormatWithSpecialCharacters(t *testing.T) {
	formatter := NewTextFormatter()

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "user action",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields: []Field{
			String("action", "login failed"),
			String("ip", "192.168.1.1"),
		},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	outputStr := string(output)

	// Fields with spaces should be quoted
	if !strings.Contains(outputStr, "action=\"login failed\"") {
		t.Errorf("Format() output should contain quoted 'action=\"login failed\"', got %s", outputStr)
	}

	// Fields without spaces should not be quoted
	if !strings.Contains(outputStr, "ip=192.168.1.1") {
		t.Errorf("Format() output should contain 'ip=192.168.1.1', got %s", outputStr)
	}
}

// TestTextFormatter_FormatWithError tests formatting error fields.
func TestTextFormatter_FormatWithError(t *testing.T) {
	formatter := NewTextFormatter()

	testErr := ErrField("test error")

	entry := &Entry{
		Level:     ErrorLevel,
		Message:   "operation failed",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields:    []Field{Err(testErr), String("component", "database")},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	outputStr := string(output)

	if !strings.Contains(outputStr, "error=test error") {
		t.Errorf("Format() output should contain 'error=test error', got %s", outputStr)
	}

	if !strings.Contains(outputStr, "component=database") {
		t.Errorf("Format() output should contain 'component=database', got %s", outputStr)
	}
}

// TestTextFormatter_WithoutColors tests formatting with colors disabled.
func TestTextFormatter_WithoutColors(t *testing.T) {
	config := FormatterConfig{
		EnableColors:    false,
		EnableTimestamp: true,
		EnableLevel:     true,
		TimestampFormat: "2006/01/02 15:04:05",
	}
	formatter := NewTextFormatterWithConfig(config)

	entry := &Entry{
		Level:     WarnLevel,
		Message:   "warning message",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	outputStr := string(output)

	// Should still contain all components without ANSI color codes
	if !strings.Contains(outputStr, "WARN") {
		t.Errorf("Format() output should contain 'WARN', got %s", outputStr)
	}

	if !strings.Contains(outputStr, "warning message") {
		t.Errorf("Format() output should contain 'warning message', got %s", outputStr)
	}
}

// TestTextFormatter_WithoutTimestamp tests formatting without timestamp.
func TestTextFormatter_WithoutTimestamp(t *testing.T) {
	config := FormatterConfig{
		EnableColors:    false,
		EnableTimestamp: false,
		EnableLevel:     true,
		TimestampFormat: "2006/01/02 15:04:05",
	}
	formatter := NewTextFormatterWithConfig(config)

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "test",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	outputStr := string(output)

	// Should not contain timestamp
	if strings.Contains(outputStr, "2026/02/28") {
		t.Errorf("Format() output should not contain timestamp, got %s", outputStr)
	}

	// Should still contain level and message
	if !strings.Contains(outputStr, "INFO") {
		t.Errorf("Format() output should contain 'INFO', got %s", outputStr)
	}

	if !strings.Contains(outputStr, "test") {
		t.Errorf("Format() output should contain 'test', got %s", outputStr)
	}
}

// TestTextFormatter_EmptyFields tests formatting entry with no fields.
func TestTextFormatter_EmptyFields(t *testing.T) {
	formatter := NewTextFormatter()

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "simple message",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields:    []Field{},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	outputStr := string(output)

	if !strings.Contains(outputStr, "simple message") {
		t.Errorf("Format() output should contain 'simple message', got %s", outputStr)
	}

	// Should not have trailing space before newline
	if strings.Contains(outputStr, "simple message \n") {
		t.Errorf("Format() output should not have trailing space, got %s", outputStr)
	}
}

// TestJSONFormatter_Format tests the JSONFormatter.Format method.
func TestJSONFormatter_Format(t *testing.T) {
	formatter := NewJSONFormatter()

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "test message",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields:    []Field{String("key", "value")},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse JSON to verify it's valid
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("Format() output is not valid JSON: %v\nOutput: %s", err, string(output))
	}

	// Check required fields
	if parsed["level"] != "INFO" {
		t.Errorf("JSON level should be 'INFO', got %v", parsed["level"])
	}

	if parsed["message"] != "test message" {
		t.Errorf("JSON message should be 'test message', got %v", parsed["message"])
	}

	// Check timestamp is present
	if _, ok := parsed["timestamp"]; !ok {
		t.Errorf("JSON should contain timestamp, got %v", parsed)
	}

	// Check fields object
	fields, ok := parsed["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("JSON fields should be an object, got %T", parsed["fields"])
	}

	if fields["key"] != "value" {
		t.Errorf("JSON field 'key' should be 'value', got %v", fields["key"])
	}

	// Check trailing newline
	if !strings.HasSuffix(string(output), "\n") {
		t.Errorf("Format() output should end with newline, got %s", string(output))
	}
}

// TestJSONFormatter_FormatWithMultipleFields tests JSON formatting with multiple fields.
func TestJSONFormatter_FormatWithMultipleFields(t *testing.T) {
	formatter := NewJSONFormatter()

	entry := &Entry{
		Level:     ErrorLevel,
		Message:   "database error",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields: []Field{
			String("host", "localhost"),
			Int("port", 5432),
			String("user", "admin"),
		},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("Format() output is not valid JSON: %v", err)
	}

	// Check fields
	fields, ok := parsed["fields"].(map[string]interface{})
	if !ok {
		t.Fatalf("JSON fields should be an object, got %T", parsed["fields"])
	}

	if fields["host"] != "localhost" {
		t.Errorf("JSON field 'host' should be 'localhost', got %v", fields["host"])
	}

	if fields["port"] != float64(5432) { // JSON numbers are float64
		t.Errorf("JSON field 'port' should be 5432, got %v", fields["port"])
	}

	if fields["user"] != "admin" {
		t.Errorf("JSON field 'user' should be 'admin', got %v", fields["user"])
	}
}

// TestJSONFormatter_FormatWithSpecialCharacters tests JSON escaping of special characters.
func TestJSONFormatter_FormatWithSpecialCharacters(t *testing.T) {
	formatter := NewJSONFormatter()

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "message with \"quotes\" and 'apostrophes'",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields: []Field{
			String("json", "{\"key\": \"value\"}"),
			String("newlines", "line1\nline2"),
		},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse JSON - should not fail even with special characters
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("Format() output is not valid JSON: %v", err)
	}

	// Check message is properly escaped
	if parsed["message"] != "message with \"quotes\" and 'apostrophes'" {
		t.Errorf("JSON message not properly escaped, got %v", parsed["message"])
	}

	// Check fields are properly escaped
	fields := parsed["fields"].(map[string]interface{})
	if fields["json"] != "{\"key\": \"value\"}" {
		t.Errorf("JSON field 'json' not properly escaped, got %v", fields["json"])
	}

	if fields["newlines"] != "line1\nline2" {
		t.Errorf("JSON field 'newlines' not properly escaped, got %v", fields["newlines"])
	}
}

// TestJSONFormatter_FormatWithNullAndBool tests JSON formatting of null and boolean values.
func TestJSONFormatter_FormatWithNullAndBool(t *testing.T) {
	formatter := NewJSONFormatter()

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "test",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields: []Field{
			Bool("enabled", true),
			Bool("disabled", false),
			Any("nil_value", nil),
		},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("Format() output is not valid JSON: %v", err)
	}

	// Check boolean values
	fields := parsed["fields"].(map[string]interface{})
	if fields["enabled"] != true {
		t.Errorf("JSON field 'enabled' should be true, got %v", fields["enabled"])
	}

	if fields["disabled"] != false {
		t.Errorf("JSON field 'disabled' should be false, got %v", fields["disabled"])
	}

	if fields["nil_value"] != nil {
		t.Errorf("JSON field 'nil_value' should be null, got %v", fields["nil_value"])
	}
}

// TestJSONFormatter_FormatWithError tests JSON formatting with error fields.
func TestJSONFormatter_FormatWithError(t *testing.T) {
	formatter := NewJSONFormatter()

	testErr := ErrField("database connection failed")

	entry := &Entry{
		Level:     ErrorLevel,
		Message:   "operation failed",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields:    []Field{Err(testErr), String("component", "database")},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("Format() output is not valid JSON: %v", err)
	}

	// Check error field
	fields := parsed["fields"].(map[string]interface{})
	if fields["error"] != "database connection failed" {
		t.Errorf("JSON error field should be 'database connection failed', got %v", fields["error"])
	}

	if fields["component"] != "database" {
		t.Errorf("JSON component field should be 'database', got %v", fields["component"])
	}
}

// TestJSONFormatter_WithoutTimestamp tests JSON formatting without timestamp.
func TestJSONFormatter_WithoutTimestamp(t *testing.T) {
	config := FormatterConfig{
		EnableTimestamp: false,
		EnableLevel:     true,
	}
	formatter := NewJSONFormatterWithConfig(config)

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "test",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("Format() output is not valid JSON: %v", err)
	}

	// Should not contain timestamp
	if _, ok := parsed["timestamp"]; ok {
		t.Errorf("JSON should not contain timestamp when disabled, got %v", parsed)
	}

	// Should still contain level and message
	if parsed["level"] != "INFO" {
		t.Errorf("JSON level should be 'INFO', got %v", parsed["level"])
	}

	if parsed["message"] != "test" {
		t.Errorf("JSON message should be 'test', got %v", parsed["message"])
	}
}

// TestJSONFormatter_EmptyFields tests JSON formatting with no fields.
func TestJSONFormatter_EmptyFields(t *testing.T) {
	formatter := NewJSONFormatter()

	entry := &Entry{
		Level:     InfoLevel,
		Message:   "simple message",
		Timestamp: time.Date(2026, 2, 28, 17, 30, 0, 0, time.UTC),
		Fields:    []Field{},
	}

	output, err := formatter.Format(entry)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("Format() output is not valid JSON: %v", err)
	}

	// Should not contain fields object when empty
	if _, ok := parsed["fields"]; ok {
		t.Errorf("JSON should not contain 'fields' key when no fields present, got %v", parsed)
	}

	// Should still contain other fields
	if parsed["message"] != "simple message" {
		t.Errorf("JSON message should be 'simple message', got %v", parsed["message"])
	}
}

// Helper function for tests
func ErrField(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
