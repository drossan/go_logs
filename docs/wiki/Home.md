# go_logs v3

**[English](#english)** | **[Espanol](Home-es)**

---

<a name="english"></a>
## English

A modern, structured logging library for Go with zero external dependencies (except optional Slack integration).

[![Go Reference](https://pkg.go.dev/badge/github.com/drossan/go_logs.svg)](https://pkg.go.dev/github.com/drossan/go_logs)
[![Go Report Card](https://goreportcard.com/badge/github.com/drossan/go_logs)](https://goreportcard.com/report/github.com/drossan/go_logs)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Overview

**go_logs** is a production-ready logging library that provides structured logging with typed fields, multiple output formatters, child loggers, context propagation, and an extensible hook system. It follows Clean Architecture principles with clear separation between interfaces and implementations.

The library maintains **100% backward compatibility** between v2 (legacy global functions) and v3 (modern Logger interface), allowing gradual migration without breaking changes.

## Key Features

- **Structured Logging** - Type-safe fields (String, Int, Int64, Float64, Bool, Err, Any)
- **Multiple Formatters** - TextFormatter for development (colors), JSONFormatter for production (ELK, Loki, Datadog)
- **Child Loggers** - Create contextual loggers with inherited fields using `With()`
- **Context Propagation** - Automatic extraction of trace_id, span_id, request_id from context
- **Hook System** - Extensible architecture for custom processing (Slack, metrics, etc.)
- **File Rotation** - Built-in RotatingFileWriter with size-based rotation (zero dependencies)
- **Zero Allocations** - Field creation and level filtering are allocation-free
- **Thread-Safe** - All operations are safe for concurrent use
- **Syslog-style Levels** - Trace=10, Debug=20, Info=30, Warn=40, Error=50, Fatal=60
- **Sensitive Data Redaction** - Automatic masking of passwords, tokens, API keys
- **Caller Information** - Optional file:line and function name in log entries
- **Stack Traces** - Automatic capture for Error level and above

## Quick Comparison: v2 vs v3

| Feature | v2 (Legacy) | v3 (Modern) |
|---------|-------------|-------------|
| API Style | Global functions | Logger interface |
| Structured Fields | No | Yes (typed fields) |
| Child Loggers | No | Yes (`With()`) |
| Context Support | No | Yes (`LogCtx()`) |
| Multiple Instances | No | Yes |
| Hooks | No | Yes |
| Dependency Injection | No | Yes (interface-based) |
| Backward Compatible | - | Yes (100%) |

### v2 Example (Legacy - Still Supported)

```go
import "github.com/drossan/go_logs"

func main() {
    go_logs.Init() // Optional, auto-initializes
    go_logs.InfoLog("Application started")
    go_logs.ErrorLog("Something went wrong")
}
```

### v3 Example (Modern - Recommended)

```go
import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Create a configured logger
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Structured logging with typed fields
    logger.Info("Server started",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
    )

    // Child logger with inherited context
    reqLogger := logger.With(
        go_logs.String("request_id", "abc-123"),
        go_logs.String("user_id", "user-456"),
    )
    reqLogger.Info("Request received")
}
```

## Installation

```bash
go get github.com/drossan/go_logs@latest
```

See [Installation](Installation.md) for detailed instructions.

## Documentation Index

### Getting Started

- [Getting Started](Getting-Started.md) - Quick start guide with your first logger
- [Installation](Installation.md) - Installation and dependencies

### Core Concepts

- [API Reference](API-Reference.md) - Complete API documentation
- [Structured Logging](Structured-Logging.md) - Field types and best practices
- [Formatters](Formatters.md) - TextFormatter and JSONFormatter configuration

### Advanced Features

- [Context and Tracing](Context-and-Tracing.md) - Distributed tracing support
- [Hooks](Hooks.md) - Custom log processing and Slack integration
- [File Rotation](File-Rotation.md) - RotatingFileWriter configuration

### Configuration

- [Configuration](Configuration.md) - Environment variables and programmatic options

### Migration

- [Migration v2 to v3](Migration-v2-to-v3.md) - Step-by-step migration guide

### Examples and Performance

- [Examples](Examples.md) - Practical code examples for common scenarios
- [Performance](Performance.md) - Benchmarks and optimization tips

## Performance Highlights

| Operation | Performance |
|-----------|-------------|
| Fast-path filtering | 0.32 ns/op |
| Field creation | 0.34 ns/op, 0 allocs |
| TextFormatter | 220.6 ns/op |
| JSONFormatter | 249.3 ns/op |
| RotatingFileWriter | 16M msg/sec |

See [Performance](Performance.md) for detailed benchmarks.

## Quick Examples

### Basic Logger

```go
logger, _ := go_logs.New()
logger.Info("Hello, World!")
```

### Logger with File Rotation

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### Logger with Context Tracing

```go
func handler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
    ctx = go_logs.WithTraceID(ctx, r.Header.Get("X-Trace-ID"))
    logger.LogCtx(ctx, go_logs.InfoLevel, "Processing request")
}
```

### HTTP Middleware Example

```go
func loggingMiddleware(logger go_logs.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            reqID := generateRequestID()

            reqLogger := logger.With(
                go_logs.String("request_id", reqID),
                go_logs.String("method", r.Method),
                go_logs.String("path", r.URL.Path),
            )

            ctx := go_logs.WithRequestID(r.Context(), reqID)
            next.ServeHTTP(w, r.WithContext(ctx))

            reqLogger.Info("Request completed",
                go_logs.Int64("duration_ms", time.Since(start).Milliseconds()),
                go_logs.Int("status", 200),
            )
        })
    }
}
```

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting PRs.

## License

MIT License - see [LICENSE](../LICENSE) for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/drossan/go_logs/issues)
- **Documentation**: This wiki
- **Go Reference**: [pkg.go.dev](https://pkg.go.dev/github.com/drossan/go_logs)
