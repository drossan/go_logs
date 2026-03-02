package go_logs_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/drossan/go_logs"
)

// Example_textFormatter demonstrates using TextFormatter for development.
func Example_textFormatter() {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create logger with TextFormatter (default)
	logger, _ := go_logs.New(
		go_logs.WithOutput(&buf),
		go_logs.WithLevel(go_logs.InfoLevel),
	)

	// Log a message with structured fields
	logger.Info("Server started",
		go_logs.String("host", "localhost"),
		go_logs.Int("port", 8080),
		go_logs.Bool("debug", false),
	)

	// Output will be colored and human-readable:
	// [2026/02/28 17:30:00] INFO Server started host=localhost port=8080 debug=false
	output := buf.String()
	println(strings.TrimSpace(output))
}

// Example_jsonFormatter demonstrates using JSONFormatter for production.
func Example_jsonFormatter() {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create logger with JSONFormatter
	logger, _ := go_logs.New(
		go_logs.WithFormatter(go_logs.NewJSONFormatter()),
		go_logs.WithOutput(&buf),
		go_logs.WithLevel(go_logs.InfoLevel),
	)

	// Log a message with structured fields
	logger.Info("Server started",
		go_logs.String("host", "localhost"),
		go_logs.Int("port", 8080),
		go_logs.Bool("debug", false),
	)

	// Output will be valid JSON:
	// {"timestamp":"2026-02-28T17:30:00Z","level":"INFO","message":"Server started","fields":{"host":"localhost","port":8080,"debug":false}}
	output := buf.String()

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err == nil {
		println("Valid JSON produced")
	}
}

// Example_textFormatterWithoutColors demonstrates disabling colors.
func Example_textFormatterWithoutColors() {
	var buf bytes.Buffer

	// Create TextFormatter without colors (for files, pipes, etc.)
	config := go_logs.FormatterConfig{
		EnableColors:    false,
		EnableTimestamp: true,
		EnableLevel:     true,
		TimestampFormat: "2006/01/02 15:04:05",
	}
	formatter := go_logs.NewTextFormatterWithConfig(config)

	logger, _ := go_logs.New(
		go_logs.WithFormatter(formatter),
		go_logs.WithOutput(&buf),
	)

	logger.Error("Database error",
		go_logs.String("host", "db.example.com"),
		go_logs.Int("port", 5432),
	)

	// Output will have no ANSI color codes
	output := buf.String()
	println(strings.TrimSpace(output))
}

// Example_jsonFormatterWithoutTimestamp demonstrates customizing JSONFormatter.
func Example_jsonFormatterWithoutTimestamp() {
	var buf bytes.Buffer

	// Create JSONFormatter without timestamp
	config := go_logs.FormatterConfig{
		EnableTimestamp: false,
		EnableLevel:     true,
	}
	formatter := go_logs.NewJSONFormatterWithConfig(config)

	logger, _ := go_logs.New(
		go_logs.WithFormatter(formatter),
		go_logs.WithOutput(&buf),
	)

	logger.Info("Request received")

	// Output will not include timestamp:
	// {"level":"INFO","message":"Request received"}
	output := buf.String()
	println(strings.TrimSpace(output))
}

// Example_loggerWithAllFieldTypes demonstrates all field types.
func Example_loggerWithAllFieldTypes() {
	var buf bytes.Buffer

	logger, _ := go_logs.New(
		go_logs.WithOutput(&buf),
	)

	// Log with all field types
	logger.Info("User action",
		go_logs.String("username", "john"),
		go_logs.Int("user_id", 12345),
		go_logs.Int64("timestamp_ms", 1677634567890),
		go_logs.Float64("score", 95.5),
		go_logs.Bool("verified", true),
	)

	output := buf.String()
	println(strings.TrimSpace(output))
}

// Example_loggerWithErrorField demonstrates error field formatting.
func Example_loggerWithErrorField() {
	var buf bytes.Buffer

	logger, _ := go_logs.New(
		go_logs.WithOutput(&buf),
	)

	// Log with error field
	err := &testError{"connection timeout"}
	logger.Error("Database connection failed",
		go_logs.Err(err),
		go_logs.String("host", "localhost"),
	)

	output := buf.String()
	println(strings.TrimSpace(output))
}

// Example_childLoggerWithFormatter demonstrates child loggers with formatters.
func Example_childLoggerWithFormatter() {
	var buf bytes.Buffer

	// Create base logger
	baseLogger, _ := go_logs.New(
		go_logs.WithOutput(&buf),
		go_logs.WithLevel(go_logs.InfoLevel),
	)

	// Create child logger with service context
	serviceLogger := baseLogger.With(
		go_logs.String("service", "api"),
		go_logs.String("version", "1.0.0"),
	)

	// All logs from serviceLogger will include service and version
	serviceLogger.Info("Handling request",
		go_logs.String("endpoint", "/users"),
		go_logs.String("method", "GET"),
	)

	output := buf.String()
	println(strings.TrimSpace(output))
}

// Example_logLevelFiltering demonstrates level-based filtering.
func Example_logLevelFiltering() {
	var buf bytes.Buffer

	logger, _ := go_logs.New(
		go_logs.WithOutput(&buf),
		go_logs.WithLevel(go_logs.WarnLevel), // Only log Warn and above
	)

	// This will be filtered out (Debug < Warn)
	logger.Debug("Detailed diagnostic information")

	// This will be logged (Warn >= Warn)
	logger.Warn("High memory usage",
		go_logs.Float64("usage_percent", 85.5),
	)

	// This will be logged (Error >= Warn)
	logger.Error("Database connection failed")

	output := buf.String()
	println(strings.TrimSpace(output))
}

// Example_multipleLoggers demonstrates multiple logger instances.
func Example_multipleLoggers() {
	var accessBuf bytes.Buffer
	var errorBuf bytes.Buffer

	// Access logger with JSON format
	accessLogger, _ := go_logs.New(
		go_logs.WithFormatter(go_logs.NewJSONFormatter()),
		go_logs.WithOutput(&accessBuf),
		go_logs.WithLevel(go_logs.InfoLevel),
	)

	// Error logger with text format
	errorLogger, _ := go_logs.New(
		go_logs.WithOutput(&errorBuf),
		go_logs.WithLevel(go_logs.ErrorLevel),
	)

	// Log to different loggers
	accessLogger.Info("API request",
		go_logs.String("method", "GET"),
		go_logs.String("path", "/api/users"),
	)

	errorLogger.Error("Database error",
		go_logs.String("query", "SELECT * FROM users"),
	)

	println("Access log:", strings.TrimSpace(accessBuf.String()))
	println("Error log:", strings.TrimSpace(errorBuf.String()))
}

// Example_formatterWithEnvVar demonstrates using LOG_FORMAT environment variable.
func Example_formatterWithEnvVar() {
	// Set environment variable before creating logger
	os.Setenv("LOG_FORMAT", "json")
	defer os.Unsetenv("LOG_FORMAT")

	// Create logger - it will automatically use JSONFormatter
	// because LOG_FORMAT=json is set
	logger, _ := go_logs.New()

	// This will output JSON format
	logger.Info("This will be JSON formatted")
}

// Helper types for examples
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
