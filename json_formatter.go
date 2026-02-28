package go_logs

import (
	"encoding/json"
	"time"
)

// JSONFormatter formats log entries as JSON for log aggregators.
// This is ideal for production environments using ELK, Loki, Datadog, etc.
//
// Output format (single line):
//
//	{"timestamp":"2026-02-28T17:30:00Z","level":"INFO","message":"connection established","fields":{"host":"db.example.com","port":5432}}
//
// The JSON output:
//   - Uses RFC3339 timestamp format for maximum compatibility
//   - Includes level as uppercase string
//   - Includes message
//   - Includes fields as a nested object (only if fields exist)
//   - Is always valid JSON (properly escapes special characters)
type JSONFormatter struct {
	config FormatterConfig
}

// NewJSONFormatter creates a new JSONFormatter with default configuration.
//
// Default configuration:
//   - EnableTimestamp: true
//   - EnableLevel: true
//   - TimestampFormat: RFC3339 (for log aggregator compatibility)
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		config: FormatterConfig{
			EnableTimestamp: true,
			EnableLevel:     true,
			TimestampFormat: time.RFC3339,
		},
	}
}

// NewJSONFormatterWithConfig creates a new JSONFormatter with custom configuration.
// This allows disabling timestamps or levels, or using a different timestamp format.
func NewJSONFormatterWithConfig(config FormatterConfig) *JSONFormatter {
	return &JSONFormatter{
		config: config,
	}
}

// JSONLogEntry represents the JSON structure of a log entry.
// This is used internally by JSONFormatter for JSON marshaling.
type JSONLogEntry struct {
	Timestamp string                 `json:"timestamp,omitempty"`
	Level     string                 `json:"level,omitempty"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// Format converts a log Entry to JSON bytes.
// Returns JSON with a trailing newline (standard for log files).
// The JSON is always valid and properly escaped.
func (f *JSONFormatter) Format(entry *Entry) ([]byte, error) {
	// Create JSON entry structure
	jsonEntry := JSONLogEntry{
		Message: entry.Message,
	}

	// Add timestamp if enabled
	if f.config.EnableTimestamp {
		jsonEntry.Timestamp = entry.Timestamp.Format(f.config.TimestampFormat)
	}

	// Add level if enabled
	if f.config.EnableLevel {
		jsonEntry.Level = entry.Level.String()
	}

	// Add fields if any exist
	if len(entry.Fields) > 0 {
		jsonEntry.Fields = make(map[string]interface{}, len(entry.Fields))
		for _, field := range entry.Fields {
			// For error types, use the Error() method to get the message
			if field.Type() == ErrorType {
				if err, ok := field.Value().(error); ok && err != nil {
					jsonEntry.Fields[field.Key()] = err.Error()
				} else {
					jsonEntry.Fields[field.Key()] = nil
				}
			} else {
				// Use field value directly - JSON encoder will handle proper escaping
				jsonEntry.Fields[field.Key()] = field.Value()
			}
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(jsonEntry)
	if err != nil {
		return nil, err
	}

	// Add trailing newline (standard for log files)
	data = append(data, '\n')
	return data, nil
}

// SetEnableTimestamp enables or disables timestamp output.
func (f *JSONFormatter) SetEnableTimestamp(enabled bool) {
	f.config.EnableTimestamp = enabled
}

// SetEnableLevel enables or disables level output.
func (f *JSONFormatter) SetEnableLevel(enabled bool) {
	f.config.EnableLevel = enabled
}

// SetTimestampFormat sets the timestamp format string.
// Default is RFC3339 for maximum compatibility with log aggregators.
// See https://pkg.go.dev/time#pkg-constants for reference.
func (f *JSONFormatter) SetTimestampFormat(format string) {
	f.config.TimestampFormat = format
}
