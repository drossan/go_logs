# Context and Tracing

go_logs provides built-in support for distributed tracing through context propagation. This allows you to correlate logs across multiple services and operations using trace IDs, span IDs, and request IDs.

## Overview

Distributed tracing is essential for debugging and monitoring microservices. When a request flows through multiple services, each service logs with the same trace ID, allowing you to:

1. **Correlate logs** across services
2. **Trace request flow** from start to finish
3. **Identify bottlenecks** and failures
4. **Debug issues** in distributed systems

## Context Keys

go_logs uses the following context keys for tracing:

| Key | Description | Example Value |
|-----|-------------|---------------|
| `trace_id` | Unique identifier for the entire request flow | `abc-123-def-456` |
| `span_id` | Identifier for a specific operation | `span-xyz-789` |
| `request_id` | Identifier scoped to a single service | `req-456` |
| `user_id` | User associated with the request | `user-123` |

## Adding Context Values

### WithTraceID

Adds a trace ID to the context.

```go
func WithTraceID(ctx context.Context, traceID string) context.Context
```

**Example:**

```go
ctx := go_logs.WithTraceID(context.Background(), "trace-abc-123")
```

### WithSpanID

Adds a span ID to the context.

```go
func WithSpanID(ctx context.Context, spanID string) context.Context
```

**Example:**

```go
ctx := go_logs.WithSpanID(ctx, "span-xyz-789")
```

### WithRequestID

Adds a request ID to the context.

```go
func WithRequestID(ctx context.Context, requestID string) context.Context
```

**Example:**

```go
ctx := go_logs.WithRequestID(ctx, "req-456")
```

### WithUserID

Adds a user ID to the context.

```go
func WithUserID(ctx context.Context, userID string) context.Context
```

**Example:**

```go
ctx := go_logs.WithUserID(ctx, "user-123")
```

## Extracting Context Values

### GetTraceID

Extracts the trace ID from context.

```go
func GetTraceID(ctx context.Context) string
```

Returns an empty string if not set.

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

Extracts all context values as structured fields.

```go
func ExtractFieldsFromContext(ctx context.Context) []Field
```

**Example:**

```go
fields := go_logs.ExtractFieldsFromContext(ctx)
// fields may contain: String("trace_id", "..."), String("span_id", "..."), etc.
```

## Logging with Context

### LogCtx

Logs a message with context support. Automatically extracts trace information.

```go
func (l Logger) LogCtx(ctx context.Context, level Level, msg string, fields ...Field)
```

**Example:**

```go
ctx := go_logs.WithTraceID(context.Background(), "trace-123")
ctx = go_logs.WithSpanID(ctx, "span-456")

logger.LogCtx(ctx, go_logs.InfoLevel, "Processing request")
```

**Output:**
```
[2026/02/28 10:30:00] INFO Processing request trace_id=trace-123 span_id=span-456
```

## HTTP Handler Example

Here's a complete example showing context propagation in HTTP handlers:

```go
package main

import (
    "context"
    "net/http"
    "os"

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

    // Wrap with tracing middleware
    wrappedMux := tracingMiddleware(mux)

    logger.Info("Server starting", go_logs.Int("port", 8080))
    http.ListenAndServe(":8080", wrappedMux)
}

// tracingMiddleware extracts or generates trace IDs
func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract trace ID from request header or generate new one
        traceID := r.Header.Get("X-Trace-ID")
        if traceID == "" {
            traceID = uuid.New().String()
        }

        // Add trace ID to context
        ctx := go_logs.WithTraceID(r.Context(), traceID)

        // Add request ID (unique per request)
        requestID := uuid.New().String()
        ctx = go_logs.WithRequestID(ctx, requestID)

        // Add to response headers for downstream services
        w.Header().Set("X-Trace-ID", traceID)
        w.Header().Set("X-Request-ID", requestID)

        // Call next handler with enriched context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// usersHandler handles user requests
func usersHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    logger.LogCtx(ctx, go_logs.InfoLevel, "Processing request",
        go_logs.String("method", r.Method),
        go_logs.String("path", r.URL.Path),
    )

    // Call service layer with context
    users, err := getUsers(ctx)
    if err != nil {
        logger.LogCtx(ctx, go_logs.ErrorLevel, "Failed to get users",
            go_logs.Err(err),
        )
        http.Error(w, "Internal Server Error", 500)
        return
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request completed",
        go_logs.Int("user_count", len(users)),
    )

    // Write response...
}

// getUsers simulates a database call with context propagation
func getUsers(ctx context.Context) ([]User, error) {
    // Log with the same trace context
    logger.LogCtx(ctx, go_logs.DebugLevel, "Querying database")

    // Simulate calling another service with context
    err := callExternalService(ctx)
    if err != nil {
        return nil, err
    }

    return []User{{ID: 1, Name: "John"}}, nil
}

// callExternalService shows propagating context to external calls
func callExternalService(ctx context.Context) error {
    // Extract trace ID to send to external service
    traceID := go_logs.GetTraceID(ctx)

    req, _ := http.NewRequest("GET", "http://external-service/api", nil)
    req.Header.Set("X-Trace-ID", traceID)

    logger.LogCtx(ctx, go_logs.DebugLevel, "Calling external service",
        go_logs.String("url", "http://external-service/api"),
    )

    // Make request...
    return nil
}
```

## Microservices Example

When a request flows through multiple services, propagate the trace ID:

### Service A (API Gateway)

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Generate or extract trace ID
    traceID := r.Header.Get("X-Trace-ID")
    if traceID == "" {
        traceID = uuid.New().String()
    }

    ctx := go_logs.WithTraceID(r.Context(), traceID)

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request received")

    // Call Service B
    resp, err := callServiceB(ctx)
    if err != nil {
        logger.LogCtx(ctx, go_logs.ErrorLevel, "Service B failed",
            go_logs.Err(err),
        )
        http.Error(w, "Error", 500)
        return
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request completed")
}

func callServiceB(ctx context.Context) (*http.Response, error) {
    traceID := go_logs.GetTraceID(ctx)

    req, _ := http.NewRequest("GET", "http://service-b/api", nil)
    req.Header.Set("X-Trace-ID", traceID)

    logger.LogCtx(ctx, go_logs.DebugLevel, "Calling Service B")
    return http.DefaultClient.Do(req)
}
```

### Service B

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Extract trace ID from incoming request
    traceID := r.Header.Get("X-Trace-ID")
    ctx := go_logs.WithTraceID(r.Context(), traceID)

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request received in Service B")

    // Process request...

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request completed in Service B")
}
```

### Correlated Logs

Service A:
```json
{"timestamp":"2026-02-28T10:30:00Z","level":"INFO","message":"Request received","fields":{"trace_id":"abc-123"}}
{"timestamp":"2026-02-28T10:30:01Z","level":"DEBUG","message":"Calling Service B","fields":{"trace_id":"abc-123"}}
{"timestamp":"2026-02-28T10:30:02Z","level":"INFO","message":"Request completed","fields":{"trace_id":"abc-123"}}
```

Service B:
```json
{"timestamp":"2026-02-28T10:30:01Z","level":"INFO","message":"Request received in Service B","fields":{"trace_id":"abc-123"}}
{"timestamp":"2026-02-28T10:30:02Z","level":"INFO","message":"Request completed in Service B","fields":{"trace_id":"abc-123"}}
```

All logs with `trace_id=abc-123` can be correlated to trace the full request flow.

## OpenTelemetry Integration

go_logs can work alongside OpenTelemetry for distributed tracing:

```go
package main

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
    "github.com/drossan/go_logs"
)

func processRequest(ctx context.Context) {
    // Get OpenTelemetry span
    span := trace.SpanFromContext(ctx)
    if span.SpanContext().IsValid() {
        // Extract trace and span IDs from OpenTelemetry
        traceID := span.SpanContext().TraceID().String()
        spanID := span.SpanContext().SpanID().String()

        // Add to go_logs context
        ctx = go_logs.WithTraceID(ctx, traceID)
        ctx = go_logs.WithSpanID(ctx, spanID)
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "Processing request")
}
```

### OpenTelemetry Helper

Create a helper to bridge OpenTelemetry and go_logs:

```go
func ContextWithTracing(ctx context.Context) context.Context {
    span := trace.SpanFromContext(ctx)
    if !span.SpanContext().IsValid() {
        return ctx
    }

    ctx = go_logs.WithTraceID(ctx, span.SpanContext().TraceID().String())
    ctx = go_logs.WithSpanID(ctx, span.SpanContext().SpanID().String())

    return ctx
}

// Usage
func handler(ctx context.Context) {
    ctx = ContextWithTracing(ctx)
    logger.LogCtx(ctx, go_logs.InfoLevel, "Processing")
}
```

## gRPC Interceptor Example

Propagate tracing through gRPC calls:

```go
package main

import (
    "context"

    "github.com/drossan/go_logs"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor extracts trace IDs from gRPC metadata
func UnaryServerInterceptor(logger go_logs.Logger) grpc.UnaryServerInterceptor {
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

// UnaryClientInterceptor injects trace IDs into gRPC metadata
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
        // Inject trace ID into outgoing metadata
        if traceID := go_logs.GetTraceID(ctx); traceID != "" {
            md := metadata.Pairs("x-trace-id", traceID)
            ctx = metadata.NewOutgoingContext(ctx, md)
        }

        return invoker(ctx, method, req, reply, cc, opts...)
    }
}
```

## Message Queue Integration

Propagate tracing through message queues (e.g., Kafka, RabbitMQ):

```go
// Producer - inject trace ID into message headers
func publishMessage(ctx context.Context, topic string, message []byte) error {
    traceID := go_logs.GetTraceID(ctx)

    headers := map[string]string{
        "X-Trace-ID": traceID,
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "Publishing message",
        go_logs.String("topic", topic),
    )

    // Publish with headers...
    return nil
}

// Consumer - extract trace ID from message headers
func consumeMessage(ctx context.Context, headers map[string]string, message []byte) {
    // Extract trace ID from headers
    if traceID, ok := headers["X-Trace-ID"]; ok {
        ctx = go_logs.WithTraceID(ctx, traceID)
    }

    logger.LogCtx(ctx, go_logs.InfoLevel, "Processing message")

    // Process message...
}
```

## Best Practices

### 1. Always Propagate Context

Pass context through all function calls:

```go
// Good
func processOrder(ctx context.Context, order *Order) error {
    logger.LogCtx(ctx, go_logs.InfoLevel, "Processing order")
    return validateOrder(ctx, order)
}

// Bad - loses context
func processOrder(order *Order) error {
    logger.Info("Processing order") // No trace info!
    return nil
}
```

### 2. Use Middleware for HTTP

Always use middleware to set up tracing context:

```go
func main() {
    mux := http.NewServeMux()
    // ... register handlers ...

    // Wrap with tracing middleware
    http.ListenAndServe(":8080", tracingMiddleware(mux))
}
```

### 3. Include Trace ID in Responses

Return trace IDs in response headers for debugging:

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    traceID := go_logs.GetTraceID(ctx)

    w.Header().Set("X-Trace-ID", traceID)
    // ... handle request ...
}
```

### 4. Generate Trace IDs at Entry Points

Generate trace IDs at service boundaries:

```go
func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := r.Header.Get("X-Trace-ID")
        if traceID == "" {
            // Generate new trace ID if not provided
            traceID = uuid.New().String()
        }
        ctx := go_logs.WithTraceID(r.Context(), traceID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## See Also

- [API Reference](API-Reference.md) - Context function documentation
- [Examples](Examples.md) - More context examples
- [Hooks](Hooks.md) - Creating hooks that use context
