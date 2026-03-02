# Examples

This page contains practical examples for common use cases with go_logs v3.

## Table of Contents

- [Basic Logger](#basic-logger)
- [Logger with Rotation](#logger-with-rotation)
- [Logger with Slack Notifications](#logger-with-slack-notifications)
- [HTTP Middleware](#http-middleware)
- [gRPC Interceptor](#grpc-interceptor)
- [Worker Queue Logging](#worker-queue-logging)
- [Microservices](#microservices)
- [Testing with Logs](#testing-with-logs)

---

## Basic Logger

### Simple Console Logger

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithOutput(os.Stdout),
    )
    if err != nil {
        panic(err)
    }

    logger.Info("Application started")
    logger.Debug("This won't be logged (level is Info)")
    logger.Error("Something went wrong", go_logs.String("component", "main"))
}
```

### Logger with Structured Fields

```go
package main

import (
    "errors"
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New()

    // String fields
    logger.Info("User action",
        go_logs.String("user_id", "123"),
        go_logs.String("action", "login"),
    )

    // Numeric fields
    logger.Info("Request completed",
        go_logs.Int("status_code", 200),
        go_logs.Int64("duration_ms", 45),
    )

    // Error field
    err := errors.New("connection refused")
    logger.Error("Database error",
        go_logs.Err(err),
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 5432),
    )

    // Boolean field
    logger.Info("Feature check",
        go_logs.Bool("feature_enabled", true),
    )

    // Float field
    logger.Info("Performance metrics",
        go_logs.Float64("cpu_percent", 45.7),
        go_logs.Float64("memory_gb", 1.23),
    )
}
```

---

## Logger with Rotation

### File Rotation with JSON Format

```go
package main

import (
    "os"
    "time"
    "github.com/drossan/go_logs"
)

func main() {
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithRotatingFile("/var/log/myapp/app.log", 100, 5),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("Application started",
        go_logs.String("version", "1.0.0"),
        go_logs.Int("pid", os.Getpid()),
    )

    // Simulate application work
    for i := 0; i < 1000; i++ {
        logger.Info("Processing request",
            go_logs.Int("request_id", i),
        )
        time.Sleep(100 * time.Millisecond)
    }

    logger.Info("Application shutting down")
}
```

### Console and File Output

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Create rotating file writer
    fileWriter, err := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
    if err != nil {
        panic(err)
    }

    // Create logger with multi-output
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithMultiOutput(fileWriter, os.Stdout),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Logs go to both file and console
    logger.Info("Application started")
}
```

---

## Logger with Slack Notifications

### Slack Hook for Errors

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "os"

    "github.com/drossan/go_logs"
)

// SlackHook sends error logs to Slack
type SlackHook struct {
    webhookURL string
    channel    string
}

func NewSlackHook(webhookURL, channel string) *SlackHook {
    return &SlackHook{
        webhookURL: webhookURL,
        channel:    channel,
    }
}

func (h *SlackHook) Run(entry *go_logs.Entry) error {
    // Only send errors and above
    if entry.Level < go_logs.ErrorLevel {
        return nil
    }

    message := map[string]interface{}{
        "channel": h.channel,
        "text":    fmt.Sprintf("[%s] %s", entry.Level, entry.Message),
        "attachments": []map[string]string{
            {
                "color": "danger",
                "text":  h.formatFields(entry.Fields),
            },
        },
    }

    body, _ := json.Marshal(message)
    resp, err := http.Post(h.webhookURL, "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    return nil
}

func (h *SlackHook) formatFields(fields []go_logs.Field) string {
    var result string
    for _, f := range fields {
        result += fmt.Sprintf("%s: %v\n", f.Key(), f.Value())
    }
    return result
}

func main() {
    slackHook := NewSlackHook(
        os.Getenv("SLACK_WEBHOOK_URL"),
        "#alerts",
    )

    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithHooks(slackHook),
    )

    logger.Info("Application started") // Not sent to Slack
    logger.Error("Database connection failed", go_logs.Err(fmt.Errorf("connection refused"))) // Sent to Slack
}
```

---

## HTTP Middleware

### Request Logging Middleware

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
    logger, err = go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", usersHandler)
    mux.HandleFunc("/api/orders", ordersHandler)

    // Wrap with logging middleware
    wrappedMux := loggingMiddleware(mux)

    logger.Info("Server starting", go_logs.Int("port", 8080))
    http.ListenAndServe(":8080", wrappedMux)
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Create request-scoped logger
        reqLogger := logger.With(
            go_logs.String("request_id", requestID),
            go_logs.String("method", r.Method),
            go_logs.String("path", r.URL.Path),
            go_logs.String("remote_addr", r.RemoteAddr),
        )

        // Add trace ID from header or generate
        traceID := r.Header.Get("X-Trace-ID")
        if traceID != "" {
            ctx := go_logs.WithTraceID(r.Context(), traceID)
            r = r.WithContext(ctx)
        }

        // Log request start
        reqLogger.Info("Request started")

        // Wrap response writer to capture status
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        // Call next handler
        next.ServeHTTP(wrapped, r)

        // Log request completion
        duration := time.Since(start)
        reqLogger.Info("Request completed",
            go_logs.Int("status", wrapped.statusCode),
            go_logs.Int64("duration_ms", duration.Milliseconds()),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    logger.Info("Users handler called")
    w.Write([]byte(`{"users": []}`))
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
    logger.Info("Orders handler called")
    w.Write([]byte(`{"orders": []}`))
}
```

### Context Propagation in Handlers

```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // Add user ID to context after authentication
    userID := authenticate(r)
    ctx = go_logs.WithUserID(ctx, userID)

    // Log with context
    logger.LogCtx(ctx, go_logs.InfoLevel, "Processing user request")

    // Call service with context
    user, err := userService.Get(ctx, userID)
    if err != nil {
        logger.LogCtx(ctx, go_logs.ErrorLevel, "Failed to get user",
            go_logs.Err(err),
        )
        http.Error(w, "Internal error", 500)
        return
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "User retrieved",
        go_logs.String("username", user.Username),
    )
}
```

---

## gRPC Interceptor

### Server Interceptor

```go
package main

import (
    "context"

    "github.com/drossan/go_logs"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

var logger go_logs.Logger

// UnaryServerInterceptor logs gRPC requests
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Extract trace ID from metadata
        if md, ok := metadata.FromIncomingContext(ctx); ok {
            if traceIDs := md.Get("x-trace-id"); len(traceIDs) > 0 {
                ctx = go_logs.WithTraceID(ctx, traceIDs[0])
            }
        }

        logger.LogCtx(ctx, go_logs.InfoLevel, "gRPC call started",
            go_logs.String("method", info.FullMethod),
        )

        resp, err := handler(ctx, req)

        if err != nil {
            logger.LogCtx(ctx, go_logs.ErrorLevel, "gRPC call failed",
                go_logs.Err(err),
            )
        } else {
            logger.LogCtx(ctx, go_logs.InfoLevel, "gRPC call completed")
        }

        return resp, err
    }
}

func main() {
    logger, _ = go_logs.New(
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    )

    server := grpc.NewServer(
        grpc.UnaryInterceptor(UnaryServerInterceptor()),
    )

    // Register services...
}
```

### Client Interceptor

```go
// UnaryClientInterceptor injects trace IDs into outgoing calls
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        // Inject trace ID into outgoing metadata
        if traceID := go_logs.GetTraceID(ctx); traceID != "" {
            md := metadata.Pairs("x-trace-id", traceID)
            ctx = metadata.NewOutgoingContext(ctx, md)
        }

        logger.LogCtx(ctx, go_logs.DebugLevel, "gRPC call",
            go_logs.String("method", method),
        )

        return invoker(ctx, method, req, reply, cc, opts...)
    }
}
```

---

## Worker Queue Logging

### Job Processor with Context

```go
package main

import (
    "context"
    "os"
    "time"

    "github.com/drossan/go_logs"
)

type Job struct {
    ID   string
    Type string
    Data interface{}
}

type Worker struct {
    logger go_logs.Logger
    jobs   chan Job
}

func NewWorker(baseLogger go_logs.Logger) *Worker {
    return &Worker{
        logger: baseLogger.With(go_logs.String("component", "worker")),
        jobs:   make(chan Job, 100),
    }
}

func (w *Worker) Start(ctx context.Context) {
    w.logger.Info("Worker started")

    for {
        select {
        case <-ctx.Done():
            w.logger.Info("Worker shutting down")
            return
        case job := <-w.jobs:
            w.processJob(ctx, job)
        }
    }
}

func (w *Worker) processJob(ctx context.Context, job Job) {
    start := time.Now()

    // Create job-scoped logger
    jobLogger := w.logger.With(
        go_logs.String("job_id", job.ID),
        go_logs.String("job_type", job.Type),
    )

    jobLogger.Info("Job started")

    err := w.executeJob(ctx, job)

    duration := time.Since(start)

    if err != nil {
        jobLogger.Error("Job failed",
            go_logs.Err(err),
            go_logs.Int64("duration_ms", duration.Milliseconds()),
        )
        return
    }

    jobLogger.Info("Job completed",
        go_logs.Int64("duration_ms", duration.Milliseconds()),
    )
}

func (w *Worker) executeJob(ctx context.Context, job Job) error {
    // Simulate work
    time.Sleep(100 * time.Millisecond)
    return nil
}

func main() {
    logger, _ := go_logs.New(
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )
    defer logger.Sync()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    worker := NewWorker(logger)
    go worker.Start(ctx)

    // Submit jobs
    for i := 0; i < 10; i++ {
        worker.jobs <- Job{
            ID:   fmt.Sprintf("job-%d", i),
            Type: "process",
            Data: i,
        }
    }

    time.Sleep(2 * time.Second)
}
```

---

## Microservices

### Service Logger Factory

```go
package logging

import (
    "os"
    "github.com/drossan/go_logs"
)

var baseLogger go_logs.Logger

func Init(serviceName string) {
    var err error
    baseLogger, err = go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
        go_logs.WithCommonRedaction(),
    )
    if err != nil {
        panic(err)
    }

    baseLogger = baseLogger.With(
        go_logs.String("service", serviceName),
    )
}

func Logger() go_logs.Logger {
    return baseLogger
}

func ServiceLogger(serviceName string) go_logs.Logger {
    return baseLogger.With(go_logs.String("service", serviceName))
}
```

### Cross-Service Request Flow

**API Gateway:**

```go
package main

import (
    "net/http"
    "github.com/google/uuid"
    "github.com/drossan/go_logs"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
    traceID := uuid.New().String()
    ctx := go_logs.WithTraceID(r.Context(), traceID)

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request received",
        go_logs.String("path", r.URL.Path),
    )

    // Call downstream service with trace ID
    req, _ := http.NewRequest("GET", "http://user-service/api", nil)
    req.Header.Set("X-Trace-ID", traceID)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        logger.LogCtx(ctx, go_logs.ErrorLevel, "Downstream call failed",
            go_logs.Err(err),
        )
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request completed")
}
```

**User Service:**

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    traceID := r.Header.Get("X-Trace-ID")
    ctx := go_logs.WithTraceID(r.Context(), traceID)

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request received in user-service")

    // Process request...

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request completed in user-service")
}
```

---

## Testing with Logs

### Capture Logs in Tests

```go
package main

import (
    "bytes"
    "testing"

    "github.com/drossan/go_logs"
)

func TestSomething(t *testing.T) {
    // Create buffer to capture logs
    var buf bytes.Buffer

    logger, _ := go_logs.New(
        go_logs.WithOutput(&buf),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    )

    // Run code that logs
    processOrder(logger, "order-123")

    // Check logs contain expected data
    output := buf.String()
    if !bytes.Contains(output, []byte("order-123")) {
        t.Error("Expected order ID in logs")
    }
}
```

### Test Helper

```go
package testing

import (
    "bytes"
    "testing"

    "github.com/drossan/go_logs"
)

// TestLogger creates a logger that captures output for testing
func TestLogger(t *testing.T) (go_logs.Logger, *bytes.Buffer) {
    var buf bytes.Buffer
    logger, err := go_logs.New(
        go_logs.WithOutput(&buf),
        go_logs.WithLevel(go_logs.DebugLevel),
    )
    if err != nil {
        t.Fatalf("Failed to create test logger: %v", err)
    }
    return logger, &buf
}

// Usage
func TestMyFunction(t *testing.T) {
    logger, buf := TestLogger(t)

    result := myFunction(logger)

    if !bytes.Contains(buf.Bytes(), []byte("expected message")) {
        t.Error("Expected log message not found")
    }
}
```

### Mock Logger for Unit Tests

```go
package main

import (
    "testing"
    "github.com/drossan/go_logs"
)

// MockLogger captures log calls for verification
type MockLogger struct {
    entries []go_logs.Entry
}

func (m *MockLogger) Log(level go_logs.Level, msg string, fields ...go_logs.Field) {
    m.entries = append(m.entries, go_logs.Entry{
        Level:   level,
        Message: msg,
        Fields:  fields,
    })
}

func (m *MockLogger) LogCtx(ctx context.Context, level go_logs.Level, msg string, fields ...go_logs.Field) {
    m.Log(level, msg, fields...)
}

func (m *MockLogger) Trace(msg string, fields ...go_logs.Field) { m.Log(go_logs.TraceLevel, msg, fields...) }
func (m *MockLogger) Debug(msg string, fields ...go_logs.Field) { m.Log(go_logs.DebugLevel, msg, fields...) }
func (m *MockLogger) Info(msg string, fields ...go_logs.Field)  { m.Log(go_logs.InfoLevel, msg, fields...) }
func (m *MockLogger) Warn(msg string, fields ...go_logs.Field)  { m.Log(go_logs.WarnLevel, msg, fields...) }
func (m *MockLogger) Error(msg string, fields ...go_logs.Field) { m.Log(go_logs.ErrorLevel, msg, fields...) }
func (m *MockLogger) Fatal(msg string, fields ...go_logs.Field) { m.Log(go_logs.FatalLevel, msg, fields...) }
func (m *MockLogger) With(fields ...go_logs.Field) go_logs.Logger { return m }
func (m *MockLogger) SetLevel(level go_logs.Level)                {}
func (m *MockLogger) GetLevel() go_logs.Level                     { return go_logs.DebugLevel }
func (m *MockLogger) Sync() error                                 { return nil }

func TestWithMockLogger(t *testing.T) {
    mock := &MockLogger{}

    processOrder(mock, "order-123")

    // Verify logs
    if len(mock.entries) != 2 {
        t.Errorf("Expected 2 log entries, got %d", len(mock.entries))
    }

    if mock.entries[0].Message != "Order processing started" {
        t.Errorf("Unexpected message: %s", mock.entries[0].Message)
    }
}
```

---

## See Also

- [API Reference](API-Reference.md) - Complete API documentation
- [Configuration](Configuration.md) - Configuration options
- [Context and Tracing](Context-and-Tracing.md) - Distributed tracing
- [Hooks](Hooks.md) - Custom log processing
