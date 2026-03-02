# Getting Started

This guide will help you get up and running with go_logs in just a few minutes.

## Prerequisites

- **Go 1.21+** - The library uses modern Go features
- A project with Go modules enabled

## Quick Start

### 1. Install the Package

```bash
go get github.com/drossan/go_logs@latest
```

### 2. Create Your First Logger

Create a new file `main.go`:

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Create a logger with default configuration
    logger, err := go_logs.New()
    if err != nil {
        panic(err)
    }

    // Basic logging
    logger.Info("Hello, go_logs!")
}
```

### 3. Run Your Program

```bash
go run main.go
```

**Expected output:**
```
[2026/02/28 10:30:00] INFO Hello, go_logs!
```

## Your First Structured Log

Structured logging adds context to your messages with typed fields:

```go
package main

import (
    "os"
    "errors"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
    )

    // Log with structured fields
    logger.Info("Server started",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
        go_logs.String("environment", "development"),
    )

    // Log an error
    err := errors.New("connection refused")
    logger.Error("Failed to connect to database",
        go_logs.Err(err),
        go_logs.String("host", "db.example.com"),
        go_logs.Int("port", 5432),
    )
}
```

**Expected output:**
```
[2026/02/28 10:30:00] INFO Server started host=localhost port=8080 environment=development
[2026/02/28 10:30:01] ERROR Failed to connect to database error=connection refused host=db.example.com port=5432
```

## Using Child Loggers

Child loggers inherit fields from their parent, reducing repetition:

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    // Base logger with service information
    baseLogger, _ := go_logs.New()

    // Create a child logger with request context
    requestLogger := baseLogger.With(
        go_logs.String("request_id", "req-abc-123"),
        go_logs.String("user_id", "user-456"),
    )

    // All logs from requestLogger include request_id and user_id
    requestLogger.Info("Request received")
    requestLogger.Info("Processing payment")
    requestLogger.Info("Request completed")
}
```

**Expected output:**
```
[2026/02/28 10:30:00] INFO Request received request_id=req-abc-123 user_id=user-456
[2026/02/28 10:30:01] INFO Processing payment request_id=req-abc-123 user_id=user-456
[2026/02/28 10:30:02] INFO Request completed request_id=req-abc-123 user_id=user-456
```

## Switching to JSON Format

For production environments, use JSON format for compatibility with log aggregators:

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Info("Server started",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
    )
}
```

**Expected output:**
```json
{"timestamp":"2026-02-28T10:30:00Z","level":"INFO","message":"Server started","fields":{"host":"localhost","port":8080}}
```

## Writing to a File with Rotation

For persistent logging with automatic rotation:

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithRotatingFile("/var/log/myapp/app.log", 100, 5),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    )
    defer logger.Sync() // Flush before exit

    logger.Info("Application started")
}
```

This creates:
- `/var/log/myapp/app.log` - Current log file
- `/var/log/myapp/app.log.1` - Most recent backup
- Up to 5 backup files, rotating when the file reaches 100MB

## Using Context for Distributed Tracing

Pass trace IDs through your application for request correlation:

```go
package main

import (
    "context"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New()

    // Add trace information to context
    ctx := context.Background()
    ctx = go_logs.WithTraceID(ctx, "trace-abc-123")
    ctx = go_logs.WithSpanID(ctx, "span-xyz-789")

    // Log with context - trace_id and span_id are automatically included
    logger.LogCtx(ctx, go_logs.InfoLevel, "Processing request",
        go_logs.String("operation", "create_user"),
    )
}
```

**Expected output:**
```
[2026/02/28 10:30:00] INFO Processing request trace_id=trace-abc-123 span_id=span-xyz-789 operation=create_user
```

## Using Environment Variables

Configure logging without changing code:

```bash
# Set log level
export LOG_LEVEL=debug

# Set format (text or json)
export LOG_FORMAT=json

# Enable file logging
export SAVE_LOG_FILE=1
export LOG_FILE_NAME=app.log
export LOG_FILE_PATH=/var/log/myapp

# Configure rotation
export LOG_MAX_SIZE=100
export LOG_MAX_BACKUPS=5
```

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    // Logger automatically picks up environment variables
    logger, _ := go_logs.New()

    logger.Debug("This will be logged if LOG_LEVEL=debug")
    logger.Info("Application started")
}
```

## Complete Example: HTTP Server

Here's a complete example showing best practices:

```go
package main

import (
    "context"
    "net/http"
    "os"
    "time"

    "github.com/drossan/go_logs"
    "github.com/google/uuid"
)

var logger go_logs.Logger

func main() {
    var err error

    // Configure logger for production
    logger, err = go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
        go_logs.WithCaller(true), // Include file:line
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Create HTTP server with logging middleware
    mux := http.NewServeMux()
    mux.HandleFunc("/api/hello", helloHandler)

    wrappedMux := loggingMiddleware(mux)

    logger.Info("Server starting",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
    )

    if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
        logger.Fatal("Server failed", go_logs.Err(err))
    }
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    logger.Info("Hello handler called")
    w.Write([]byte("Hello, World!"))
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := uuid.New().String()

        // Create request-scoped logger
        reqLogger := logger.With(
            go_logs.String("request_id", requestID),
            go_logs.String("method", r.Method),
            go_logs.String("path", r.URL.Path),
            go_logs.String("remote_addr", r.RemoteAddr),
        )

        // Add request ID to context
        ctx := go_logs.WithRequestID(r.Context(), requestID)

        // Log request start
        reqLogger.Info("Request started")

        // Call next handler
        next.ServeHTTP(w, r.WithContext(ctx))

        // Log request completion
        duration := time.Since(start)
        reqLogger.Info("Request completed",
            go_logs.Int64("duration_ms", duration.Milliseconds()),
        )
    })
}
```

## Next Steps

Now that you have go_logs running, explore these topics:

- [API Reference](API-Reference.md) - Complete API documentation
- [Structured Logging](Structured-Logging.md) - All field types and best practices
- [Formatters](Formatters.md) - Customize output format
- [Context and Tracing](Context-and-Tracing.md) - Distributed tracing
- [Hooks](Hooks.md) - Custom log processing
- [File Rotation](File-Rotation.md) - Production file logging
- [Configuration](Configuration.md) - All configuration options
- [Examples](Examples.md) - More practical examples

## Common Issues

### No Output Appears

Make sure the log level is set appropriately:

```go
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel), // Enable debug and above
)
```

### Colors Not Working in Terminal

Colors are enabled by default. If you see escape codes, your terminal may not support ANSI colors. Disable them:

```go
formatter := go_logs.NewTextFormatter()
formatter.SetEnableColors(false)
logger, _ := go_logs.New(
    go_logs.WithFormatter(formatter),
)
```

### Logs Not Appearing in File

Ensure `Sync()` is called before exit:

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("app.log", 100, 5),
)
defer logger.Sync() // Important!
```
