# API Reference

Complete reference documentation for go_logs v3 API.

## Table of Contents

- [Logger Interface](#logger-interface)
- [Creating Loggers](#creating-loggers)
- [Log Levels](#log-levels)
- [Structured Fields](#structured-fields)
- [Configuration Options](#configuration-options)
- [Formatters](#formatters)
- [Context Functions](#context-functions)
- [Hooks](#hooks)
- [File Rotation](#file-rotation)
- [Entry Type](#entry-type)
- [Global Logger](#global-logger)
- [v2 Legacy API](#v2-legacy-api)

---

## Logger Interface

The `Logger` interface is the core abstraction for all logging operations.

```go
type Logger interface {
    // Core logging methods
    Log(level Level, msg string, fields ...Field)
    LogCtx(ctx context.Context, level Level, msg string, fields ...Field)

    // Convenience methods for each level
    Trace(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)

    // Child logger creation
    With(fields ...Field) Logger

    // Configuration
    SetLevel(level Level)
    GetLevel() Level

    // Lifecycle
    Sync() error
}
```

### Log

Logs a message at the specified level with structured fields.

```go
func (l Logger) Log(level Level, msg string, fields ...Field)
```

**Parameters:**
- `level` - The severity level (TraceLevel, DebugLevel, etc.)
- `msg` - The log message
- `fields` - Zero or more structured fields

**Example:**

```go
logger.Log(go_logs.InfoLevel, "User logged in",
    go_logs.String("username", "john"),
    go_logs.String("ip", "192.168.1.1"),
)
```

### LogCtx

Logs a message with context support for distributed tracing.

```go
func (l Logger) LogCtx(ctx context.Context, level Level, msg string, fields ...Field)
```

**Parameters:**
- `ctx` - Context containing trace_id, span_id, etc.
- `level` - The severity level
- `msg` - The log message
- `fields` - Zero or more structured fields

**Example:**

```go
ctx := go_logs.WithTraceID(context.Background(), "trace-123")
logger.LogCtx(ctx, go_logs.InfoLevel, "Processing request")
```

### Trace

Logs at TraceLevel. Use for extremely detailed, high-volume logging.

```go
func (l Logger) Trace(msg string, fields ...Field)
```

**Example:**

```go
logger.Trace("Function entered",
    go_logs.String("function", "processData"),
    go_logs.Any("args", args),
)
```

### Debug

Logs at DebugLevel. Use for detailed diagnostic information.

```go
func (l Logger) Debug(msg string, fields ...Field)
```

**Example:**

```go
logger.Debug("Processing request",
    go_logs.String("endpoint", "/api/users"),
    go_logs.String("method", "GET"),
)
```

### Info

Logs at InfoLevel. Use for general informational messages.

```go
func (l Logger) Info(msg string, fields ...Field)
```

**Example:**

```go
logger.Info("Server started",
    go_logs.String("host", "localhost"),
    go_logs.Int("port", 8080),
)
```

### Warn

Logs at WarnLevel. Use for potentially harmful situations.

```go
func (l Logger) Warn(msg string, fields ...Field)
```

**Example:**

```go
logger.Warn("High memory usage",
    go_logs.Float64("usage_percent", 85.5),
    go_logs.Int64("bytes_used", 17179869184),
)
```

### Error

Logs at ErrorLevel. Use for error events.

```go
func (l Logger) Error(msg string, fields ...Field)
```

**Example:**

```go
err := errors.New("connection refused")
logger.Error("Database connection failed",
    go_logs.Err(err),
    go_logs.String("host", "db.example.com"),
)
```

### Fatal

Logs at FatalLevel and terminates the application with `os.Exit(1)`.

```go
func (l Logger) Fatal(msg string, fields ...Field)
```

**Example:**

```go
if config == nil {
    logger.Fatal("Cannot load configuration",
        go_logs.String("file", "/etc/app/config.yaml"),
    )
}
```

**Warning**: `Fatal` terminates the application. Use only for unrecoverable errors.

### With

Creates a child logger with pre-pended fields.

```go
func (l Logger) With(fields ...Field) Logger
```

**Returns:** A new Logger instance with inherited configuration and fields.

**Example:**

```go
baseLogger, _ := go_logs.New()

// Create child logger with request context
requestLogger := baseLogger.With(
    go_logs.String("request_id", "req-123"),
    go_logs.String("user_id", "user-456"),
)

// All logs from requestLogger include request_id and user_id
requestLogger.Info("Processing request")
requestLogger.Info("Request completed")
```

### SetLevel

Changes the minimum log level threshold at runtime.

```go
func (l Logger) SetLevel(level Level)
```

**Example:**

```go
// Enable debug logging dynamically
if debugMode {
    logger.SetLevel(go_logs.DebugLevel)
}
```

### GetLevel

Returns the current minimum log level threshold.

```go
func (l Logger) GetLevel() Level
```

**Example:**

```go
if logger.GetLevel() >= go_logs.DebugLevel {
    // Debug logging is enabled, do expensive logging
    logger.Debug("Detailed state", go_logs.Any("state", expensiveDump()))
}
```

### Sync

Flushes any buffered log entries. Call before application exit.

```go
func (l Logger) Sync() error
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("app.log", 100, 5),
)
defer logger.Sync() // Ensure all logs are written
```

---

## Creating Loggers

### New

Creates a new Logger with the given options.

```go
func New(opts ...Option) (Logger, error)
```

**Parameters:**
- `opts` - Zero or more Option functions

**Returns:**
- `Logger` - Configured logger instance
- `error` - Configuration error (e.g., invalid file path)

**Default Configuration:**
- Level: InfoLevel
- Output: os.Stdout
- Formatter: TextFormatter
- No hooks
- No redaction

**Example:**

```go
// Basic logger with defaults
logger, err := go_logs.New()
if err != nil {
    panic(err)
}

// Configured logger
logger, err := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithOutput(os.Stdout),
    go_logs.WithCaller(true),
)
```

---

## Log Levels

Log levels follow syslog-style numeric ordering for fast threshold filtering.

### Level Type

```go
type Level int
```

### Level Constants

```go
const (
    TraceLevel   Level = 10  // Extremely detailed
    DebugLevel   Level = 20  // Diagnostic information
    InfoLevel    Level = 30  // General operational messages
    SuccessLevel Level = 25  // Successful operations (v2 compatibility)
    WarnLevel    Level = 40  // Potentially harmful situations
    ErrorLevel   Level = 50  // Error events
    FatalLevel   Level = 60  // Critical errors (terminates app)
    SilentLevel  Level = 0   // Disables all logging
)
```

### ParseLevel

Converts a string to a Level.

```go
func ParseLevel(levelStr string) Level
```

**Supported values:** trace, debug, info, warn, warning, error, fatal, silent, none, disable

**Example:**

```go
level := go_logs.ParseLevel("debug") // Returns DebugLevel
level := go_logs.ParseLevel("WARNING") // Returns WarnLevel (case-insensitive)
level := go_logs.ParseLevel("unknown") // Returns InfoLevel (default)
```

### Level.String

Returns the string representation.

```go
func (l Level) String() string
```

**Example:**

```go
level := go_logs.InfoLevel
fmt.Println(level.String()) // Output: INFO
```

### Level.ShouldLog

Checks if this level should be logged given a threshold.

```go
func (l Level) ShouldLog(threshold Level) bool
```

**Example:**

```go
if entry.Level.ShouldLog(go_logs.ErrorLevel) {
    // Send to monitoring system
}
```

---

## Structured Fields

Fields are typed key-value pairs for structured logging. See [Structured Logging](Structured-Logging.md) for detailed documentation.

### Field Functions

```go
func String(key, val string) Field
func Int(key string, val int) Field
func Int64(key string, val int64) Field
func Float64(key string, val float64) Field
func Bool(key string, val bool) Field
func Err(err error) Field
func Any(key string, val interface{}) Field
```

### Field Methods

```go
func (f Field) Key() string
func (f Field) Type() FieldType
func (f Field) Value() interface{}
func (f Field) StringValue() string
func (f Field) IntValue() int
func (f Field) Int64Value() int64
func (f Field) Float64Value() float64
func (f Field) BoolValue() bool
func (f Field) ErrorValue() error
```

---

## Configuration Options

Options configure the logger using the functional options pattern.

### WithLevel

Sets the minimum log level threshold.

```go
func WithLevel(level Level) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel),
)
```

### WithOutput

Sets the output destination.

```go
func WithOutput(w io.Writer) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithOutput(os.Stdout),
)
```

### WithFormatter

Sets the log formatter.

```go
func WithFormatter(f Formatter) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### WithHooks

Adds hooks for custom log processing.

```go
func WithHooks(hooks ...Hook) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithHooks(myCustomHook),
)
```

### WithRedactor

Enables redaction of sensitive fields.

```go
func WithRedactor(keys ...string) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithRedactor("password", "token", "api_key"),
)
```

### WithCommonRedaction

Enables redaction of common sensitive fields.

```go
func WithCommonRedaction() Option
```

Redacts: password, passwd, pwd, token, api_key, apikey, api-key, secret, authorization, auth, cookie, session, credit_card, ssn, social_security

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithCommonRedaction(),
)
logger.Info("User login", go_logs.String("password", "secret123"))
// Output: password=***
```

### WithCaller

Enables caller information (file:line function).

```go
func WithCaller(enabled bool) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithCaller(true),
)
logger.Info("Hello")
// Output: [2026/02/28 10:30:00] INFO main.go:42 main.Hello Hello
```

### WithStackTrace

Enables stack trace capture.

```go
func WithStackTrace(enabled bool) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithStackTrace(true),
)
```

### WithStackTraceLevel

Sets the minimum level for automatic stack traces.

```go
func WithStackTraceLevel(level Level) Option
```

**Default:** ErrorLevel

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithStackTrace(true),
    go_logs.WithStackTraceLevel(go_logs.WarnLevel),
)
```

### WithCallerSkip

Sets the number of stack frames to skip for caller info.

```go
func WithCallerSkip(skip int) Option
```

**Default:** 2

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithCaller(true),
    go_logs.WithCallerSkip(3), // For wrapper functions
)
```

### WithMultiOutput

Enables output to multiple writers.

```go
func WithMultiOutput(writers ...io.Writer) Option
```

**Example:**

```go
file, _ := go_logs.NewRotatingFileWriter("app.log", 100, 5)
logger, _ := go_logs.New(
    go_logs.WithMultiOutput(file, os.Stdout),
)
```

### WithRotatingFile

Creates a rotating file writer.

```go
func WithRotatingFile(filename string, maxSizeMB int, maxBackups int) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
)
```

### WithRotatingFileEnhanced

Creates a rotating file writer with full configuration.

```go
func WithRotatingFileEnhanced(config RotatingFileConfig) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFileEnhanced(go_logs.RotatingFileConfig{
        Filename:     "/var/log/app.log",
        MaxSizeMB:    100,
        MaxBackups:   5,
        RotationType: go_logs.RotateDaily,
        Compress:     true,
        MaxAge:       30,
    }),
)
```

---

## Formatters

### Formatter Interface

```go
type Formatter interface {
    Format(entry *Entry) ([]byte, error)
}
```

### NewTextFormatter

Creates a TextFormatter with default configuration.

```go
func NewTextFormatter() *TextFormatter
```

### NewTextFormatterWithConfig

Creates a TextFormatter with custom configuration.

```go
func NewTextFormatterWithConfig(config FormatterConfig) *TextFormatter
```

### TextFormatter Methods

```go
func (f *TextFormatter) SetEnableColors(enabled bool)
func (f *TextFormatter) SetEnableTimestamp(enabled bool)
func (f *TextFormatter) SetEnableLevel(enabled bool)
func (f *TextFormatter) SetTimestampFormat(format string)
```

### NewJSONFormatter

Creates a JSONFormatter with default configuration.

```go
func NewJSONFormatter() *JSONFormatter
```

### NewJSONFormatterWithConfig

Creates a JSONFormatter with custom configuration.

```go
func NewJSONFormatterWithConfig(config FormatterConfig) *JSONFormatter
```

### JSONFormatter Methods

```go
func (f *JSONFormatter) SetEnableTimestamp(enabled bool)
func (f *JSONFormatter) SetEnableLevel(enabled bool)
func (f *JSONFormatter) SetTimestampFormat(format string)
```

---

## Context Functions

Functions for managing context values for distributed tracing.

### WithTraceID

Adds a trace ID to the context.

```go
func WithTraceID(ctx context.Context, traceID string) context.Context
```

### WithSpanID

Adds a span ID to the context.

```go
func WithSpanID(ctx context.Context, spanID string) context.Context
```

### WithRequestID

Adds a request ID to the context.

```go
func WithRequestID(ctx context.Context, requestID string) context.Context
```

### WithUserID

Adds a user ID to the context.

```go
func WithUserID(ctx context.Context, userID string) context.Context
```

### GetTraceID

Extracts the trace ID from context.

```go
func GetTraceID(ctx context.Context) string
```

### GetSpanID

Extracts the span ID from context.

```go
func GetSpanID(ctx context.Context) string
```

### GetRequestID

Extracts the request ID from context.

```go
func GetRequestID(ctx context.Context) string
```

### GetUserID

Extracts the user ID from context.

```go
func GetUserID(ctx context.Context) string
```

### ExtractFieldsFromContext

Extracts all context values as Fields.

```go
func ExtractFieldsFromContext(ctx context.Context) []Field
```

---

## Hooks

### Hook Interface

```go
type Hook interface {
    Run(entry *Entry) error
}
```

### HookFunc

Adapter type for using functions as hooks.

```go
type HookFunc func(entry *Entry) error

func (f HookFunc) Run(entry *Entry) error
```

### NewFuncHook

Creates a Hook from a function.

```go
func NewFuncHook(fn func(entry *Entry) error) Hook
```

**Example:**

```go
hook := go_logs.NewFuncHook(func(entry *go_logs.Entry) error {
    // Send errors to monitoring
    if entry.Level >= go_logs.ErrorLevel {
        monitoring.TrackError(entry.Message)
    }
    return nil
})

logger, _ := go_logs.New(
    go_logs.WithHooks(hook),
)
```

---

## File Rotation

### NewRotatingFileWriter

Creates a rotating file writer.

```go
func NewRotatingFileWriter(filename string, maxSizeMB int, maxBackups int) (*RotatingFileWriter, error)
```

**Parameters:**
- `filename` - Path to the log file
- `maxSizeMB` - Maximum size in MB before rotation
- `maxBackups` - Maximum number of backup files to keep

**Example:**

```go
writer, err := go_logs.NewRotatingFileWriter("app.log", 100, 5)
if err != nil {
    panic(err)
}
defer writer.Close()

logger, _ := go_logs.New(
    go_logs.WithOutput(writer),
)
```

### RotatingFileWriter Methods

```go
func (w *RotatingFileWriter) Write(p []byte) (n int, err error)
func (w *RotatingFileWriter) Rotate() error
func (w *RotatingFileWriter) Sync() error
func (w *RotatingFileWriter) Close() error
func (w *RotatingFileWriter) GetMaxSize() int64
func (w *RotatingFileWriter) GetMaxBackups() int
```

---

## Entry Type

Entry represents a complete log entry.

```go
type Entry struct {
    Level      Level
    Message    string
    Fields     []Field
    Timestamp  time.Time
    Caller     *CallerInfo
    StackTrace []byte
}
```

### Entry Methods

```go
func (e *Entry) String() string
func (e *Entry) Clone() *Entry
func (e *Entry) WithFields(fields ...Field) *Entry
func (e *Entry) WithLevel(level Level) *Entry
func (e *Entry) WithMessage(msg string) *Entry
func (e *Entry) WithTimestamp(ts time.Time) *Entry
func (e *Entry) HasField(key string) bool
func (e *Entry) GetField(key string) Field
func (e *Entry) GetFieldValue(key string) interface{}
func (e *Entry) FieldCount() int
func (e *Entry) GetFieldsByPrefix(prefix string) []Field
func (e *Entry) RemoveField(key string) bool
func (e *Entry) ReplaceFieldValue(key string, value interface{}) bool
func (e *Entry) FormatBasic() string
```

---

## Global Logger

### SetGlobal

Sets the global logger instance.

```go
func SetGlobal(logger Logger)
```

### GetGlobal

Gets the global logger instance.

```go
func GetGlobal() Logger
```

### Global Convenience Functions

```go
func Global() Logger                    // Returns global logger
func L() Logger                         // Shorthand for Global()
```

---

## v2 Legacy API

The v2 API is fully backward compatible and maintained alongside v3.

### Initialization

```go
func Init()
func Close()
```

### Logging Functions

```go
func TraceLog(msg string)
func DebugLog(msg string)
func InfoLog(msg string)
func SuccessLog(msg string)
func WarningLog(msg string)
func ErrorLog(msg string)
func FatalLog(msg string)
```

### Formatted Logging

```go
func Tracef(format string, args ...interface{})
func Debugf(format string, args ...interface{})
func Infof(format string, args ...interface{})
func Successf(format string, args ...interface{})
func Warningf(format string, args ...interface{})
func Errorf(format string, args ...interface{})
func Fatalf(format string, args ...interface{})
```

### Notification Functions

```go
func IsNotifierEnabled() bool
```

---

## See Also

- [Structured Logging](Structured-Logging.md) - Detailed field documentation
- [Formatters](Formatters.md) - Formatter configuration
- [Hooks](Hooks.md) - Hook system documentation
- [Configuration](Configuration.md) - Environment variables
