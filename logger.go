package go_logs

import (
	"context"
)

// Logger defines the main logging interface.
//
// This interface enables dependency injection and allows multiple logger instances
// with different configurations. It provides both structured logging with fields
// and convenience methods for common log levels.
//
// Example:
//
//	logger, _ := go_logs.New(
//	    go_logs.WithLevel(go_logs.InfoLevel),
//	    go_logs.WithOutput(os.Stdout),
//	)
//	logger.Info("Server started", go_logs.String("port", "8080"))
type Logger interface {
	// Log logs a message at the specified level with structured fields.
	//
	// This is the core logging method. All other logging methods are built on top of it.
	// The level is checked against the configured threshold before any processing occurs.
	//
	// Parameters:
	//   level  - The severity level (Trace, Debug, Info, Warn, Error, Fatal)
	//   msg    - The log message
	//   fields - Structured key-value pairs (use String(), Int(), Err() helpers)
	//
	// Example:
	//   logger.Log(go_logs.InfoLevel, "User logged in",
	//       go_logs.String("username", "john"),
	//       go_logs.String("ip", "192.168.1.1"))
	Log(level Level, msg string, fields ...Field)

	// LogCtx logs a message with context support.
	//
	// The context is used to extract distributed tracing information (trace_id, span_id, etc.)
	// and automatically include it in the log entry. This is essential for microservices.
	//
	// Parameters:
	//   ctx    - The context containing tracing information
	//   level  - The severity level
	//   msg    - The log message
	//   fields - Structured key-value pairs
	//
	// Example:
	//   ctx := context.WithValue(context.Background(), traceIDKey, "abc-123")
	//   logger.LogCtx(ctx, go_logs.InfoLevel, "Request received")
	LogCtx(ctx context.Context, level Level, msg string, fields ...Field)

	// Trace logs a message at TraceLevel.
	// Use for extremely detailed, high-volume logging (e.g., function entry/exit).
	//
	// Example:
	//   logger.Trace("Function called", go_logs.String("function", "processData"))
	Trace(msg string, fields ...Field)

	// Debug logs a message at DebugLevel.
	// Use for detailed diagnostic information for troubleshooting.
	//
	// Example:
	//   logger.Debug("Processing request", go_logs.String("endpoint", "/api/users"))
	Debug(msg string, fields ...Field)

	// Info logs a message at InfoLevel.
	// Use for general informational messages about normal operation.
	//
	// Example:
	//   logger.Info("Server started", go_logs.Int("port", 8080))
	Info(msg string, fields ...Field)

	// Warn logs a message at WarnLevel.
	// Use for potentially harmful situations that don't prevent continued operation.
	//
	// Example:
	//   logger.Warn("High memory usage", go_logs.Float64("usage_percent", 85.5))
	Warn(msg string, fields ...Field)

	// Error logs a message at ErrorLevel.
	// Use for error events that might still allow the application to continue.
	//
	// Example:
	//   logger.Error("Database connection failed",
	//       go_logs.Err(err),
	//       go_logs.String("host", "localhost"))
	Error(msg string, fields ...Field)

	// Fatal logs a message at FatalLevel and terminates the application.
	// Use for critical errors that require immediate termination.
	//
	// After logging, this method calls os.Exit(1) or similar to terminate.
	//
	// Example:
	//   logger.Fatal("Cannot load config", go_logs.Err(err))
	Fatal(msg string, fields ...Field)

	// With creates a child logger with pre-pended fields.
	//
	// The child logger inherits all configuration from the parent but adds
	// the provided fields to every log message. This is useful for adding
	// context like request_id, user_id, service name, etc.
	//
	// The parent logger is not modified.
	//
	// Example:
	//   baseLogger, _ := go_logs.New()
	//   requestLogger := baseLogger.With(
	//       go_logs.String("request_id", "abc-123"),
	//       go_logs.String("user_id", "456"),
	//   )
	//   // All logs from requestLogger will include request_id and user_id
	//   requestLogger.Info("Processing request")
	With(fields ...Field) Logger

	// SetLevel changes the minimum log level threshold.
	//
	// Only messages at or above this level will be logged.
	// This can be changed at runtime to enable/disable verbose logging.
	//
	// Example:
	//   logger.SetLevel(go_logs.DebugLevel) // Enable debug logging
	SetLevel(level Level)

	// GetLevel returns the current minimum log level threshold.
	//
	// Example:
	//   if logger.GetLevel() >= go_logs.DebugLevel {
	//       // Expensive debug logging is enabled
	//   }
	GetLevel() Level

	// Sync flushes any buffered log entries.
	//
	// Call this before application shutdown to ensure all logs are written.
	// Returns an error if flushing fails.
	//
	// Example:
	//   defer logger.Sync()
	Sync() error
}

// New creates a new Logger with the given options.
//
// This is the primary constructor for Logger instances. It returns a Logger
// interface that can be used for all logging operations.
//
// If no options are provided, sensible defaults are used:
//   - Level: InfoLevel
//   - Output: os.Stdout
//   - Formatter: TextFormatter
//   - No hooks
//   - No redaction
//
// Parameters:
//
//	opts - Zero or more Option functions to configure the logger
//
// Returns:
//
//	logger - A configured Logger instance
//	error  - An error if configuration fails (e.g., invalid file path)
//
// Example:
//
//	logger, err := go_logs.New(
//	    go_logs.WithLevel(go_logs.DebugLevel),
//	    go_logs.WithOutput(os.Stdout),
//	    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func New(opts ...Option) (Logger, error) {
	// Create a new logger with the given options
	return NewLogger(opts...)
}
