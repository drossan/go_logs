# Structured Logging

Structured logging adds context to your log messages using typed key-value pairs called fields. This makes logs searchable, filterable, and easier to analyze in log aggregation systems.

## What is Structured Logging?

Traditional logging uses plain text messages:

```
[2026/02/28 10:30:00] INFO User john logged in from 192.168.1.1
```

Structured logging separates the message from the data:

```
[2026/02/28 10:30:00] INFO User logged in username=john ip=192.168.1.1
```

In JSON format, this becomes even more powerful:

```json
{
  "timestamp": "2026-02-28T10:30:00Z",
  "level": "INFO",
  "message": "User logged in",
  "fields": {
    "username": "john",
    "ip": "192.168.1.1"
  }
}
```

This structure allows you to:
- Search by field values (`username:john`)
- Filter by field presence (`exists:ip`)
- Aggregate and analyze data (count logins by username)

## Field Types

go_logs provides typed field constructors for different data types:

### String

For text values like usernames, hostnames, IDs, etc.

```go
go_logs.String(key, value string) Field
```

**Example:**

```go
logger.Info("User logged in",
    go_logs.String("username", "john"),
    go_logs.String("email", "john@example.com"),
    go_logs.String("ip", "192.168.1.1"),
)
```

**Output (Text):**
```
[2026/02/28 10:30:00] INFO User logged in username=john email=john@example.com ip=192.168.1.1
```

**Output (JSON):**
```json
{"timestamp":"2026-02-28T10:30:00Z","level":"INFO","message":"User logged in","fields":{"username":"john","email":"john@example.com","ip":"192.168.1.1"}}
```

### Int

For integer values like ports, counts, status codes, etc.

```go
go_logs.Int(key string, value int) Field
```

**Example:**

```go
logger.Info("HTTP request",
    go_logs.Int("status_code", 200),
    go_logs.Int("response_size", 1024),
    go_logs.Int("port", 8080),
)
```

### Int64

For large integers like file sizes, timestamps, database IDs, etc.

```go
go_logs.Int64(key string, value int64) Field
```

**Example:**

```go
logger.Info("File processed",
    go_logs.Int64("bytes", 1536299520),
    go_logs.Int64("file_id", 9876543210),
    go_logs.Int64("timestamp_ns", time.Now().UnixNano()),
)
```

### Float64

For floating-point values like percentages, measurements, rates, etc.

```go
go_logs.Float64(key string, value float64) Field
```

**Example:**

```go
logger.Info("System metrics",
    go_logs.Float64("cpu_percent", 45.7),
    go_logs.Float64("memory_gb", 1.23),
    go_logs.Float64("request_rate", 1250.5),
)
```

### Bool

For boolean flags and states.

```go
go_logs.Bool(key string, value bool) Field
```

**Example:**

```go
logger.Info("Feature flags",
    go_logs.Bool("debug_mode", true),
    go_logs.Bool("cache_enabled", false),
    go_logs.Bool("maintenance_mode", false),
)
```

### Err

For error values. The key is automatically set to `"error"`.

```go
go_logs.Err(err error) Field
```

**Example:**

```go
err := errors.New("connection refused")
logger.Error("Database connection failed",
    go_logs.Err(err),
    go_logs.String("host", "db.example.com"),
)
```

**Output:**
```
[2026/02/28 10:30:00] ERROR Database connection failed error=connection refused host=db.example.com
```

**Note:** The error's `Error()` method is called to get the string representation.

### Any

For arbitrary values that don't fit other types. Uses `fmt.Sprintf("%v")` for formatting.

```go
go_logs.Any(key string, value interface{}) Field
```

**Example:**

```go
type User struct {
    ID   int
    Name string
}

user := User{ID: 123, Name: "John"}
logger.Info("User created",
    go_logs.Any("user", user),
)
```

**Output:**
```
[2026/02/28 10:30:00] INFO User created user={123 John}
```

**Note:** For complex types, consider using JSON serialization or custom formatting.

## Field Type Reference

| Function | Go Type | JSON Type | Use Case |
|----------|---------|-----------|----------|
| `String` | string | string | Text values (names, IDs, URLs) |
| `Int` | int | number | Small integers (counts, ports) |
| `Int64` | int64 | number | Large integers (file sizes, timestamps) |
| `Float64` | float64 | number | Decimals (percentages, measurements) |
| `Bool` | bool | boolean | Flags and states |
| `Err` | error | string | Error values (key is "error") |
| `Any` | interface{} | varies | Arbitrary values |

## Field Methods

Fields provide methods for accessing their values:

```go
field := go_logs.String("username", "john")

field.Key()          // "username"
field.Type()         // StringType
field.Value()        // "john" (interface{})
field.StringValue()  // "john" (typed accessor)
```

### Typed Accessors

```go
// Returns the typed value, or zero value if type doesn't match
field.StringValue()   // string
field.IntValue()      // int
field.Int64Value()    // int64
field.Float64Value()  // float64
field.BoolValue()     // bool
field.ErrorValue()    // error
```

## Best Practices

### Use Consistent Field Names

Use the same field name for the same concept across your application:

```go
// Good: Consistent naming
logger.Info("User logged in", go_logs.String("user_id", "123"))
logger.Info("Order created", go_logs.String("user_id", "123"))
logger.Info("Payment processed", go_logs.String("user_id", "123"))

// Bad: Inconsistent naming
logger.Info("User logged in", go_logs.String("user_id", "123"))
logger.Info("Order created", go_logs.String("uid", "123"))
logger.Info("Payment processed", go_logs.String("customer_id", "123"))
```

### Use Descriptive Keys

Choose clear, descriptive field names:

```go
// Good
logger.Info("Request completed",
    go_logs.Int("response_time_ms", 45),
    go_logs.Int("status_code", 200),
)

// Bad
logger.Info("Request completed",
    go_logs.Int("time", 45),
    go_logs.Int("code", 200),
)
```

### Use Appropriate Types

Match field types to the data:

```go
// Good: Use Int64 for large numbers
logger.Info("File processed",
    go_logs.Int64("bytes", 1536299520),
)

// Bad: Use String for numbers (loses numeric indexing)
logger.Info("File processed",
    go_logs.String("bytes", "1536299520"),
)
```

### Use Child Loggers for Context

Create child loggers with common fields to avoid repetition:

```go
// Instead of repeating fields:
logger.Info("Request started", go_logs.String("request_id", "abc-123"))
logger.Info("Processing", go_logs.String("request_id", "abc-123"))
logger.Info("Request completed", go_logs.String("request_id", "abc-123"))

// Use a child logger:
reqLogger := logger.With(go_logs.String("request_id", "abc-123"))
reqLogger.Info("Request started")
reqLogger.Info("Processing")
reqLogger.Info("Request completed")
```

### Include Error Context

When logging errors, include relevant context:

```go
err := db.Connect()
if err != nil {
    logger.Error("Database connection failed",
        go_logs.Err(err),
        go_logs.String("host", cfg.DB.Host),
        go_logs.Int("port", cfg.DB.Port),
        go_logs.String("database", cfg.DB.Name),
    )
}
```

### Use Time Durations

Log durations as integers (milliseconds) for easy filtering:

```go
start := time.Now()
// ... do work ...
duration := time.Since(start)

logger.Info("Request completed",
    go_logs.Int64("duration_ms", duration.Milliseconds()),
)
```

## Common Patterns

### HTTP Request Logging

```go
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    requestID := r.Header.Get("X-Request-ID")

    // Create request-scoped logger
    reqLogger := h.logger.With(
        go_logs.String("request_id", requestID),
        go_logs.String("method", r.Method),
        go_logs.String("path", r.URL.Path),
        go_logs.String("remote_addr", r.RemoteAddr),
    )

    reqLogger.Info("Request started")

    // ... handle request ...

    reqLogger.Info("Request completed",
        go_logs.Int("status", statusCode),
        go_logs.Int64("duration_ms", time.Since(start).Milliseconds()),
        go_logs.Int("response_bytes", responseSize),
    )
}
```

### Database Operation Logging

```go
func (db *Database) Query(ctx context.Context, query string, args ...interface{}) {
    start := time.Now()

    logger.Debug("Database query started",
        go_logs.String("query", query),
        go_logs.Any("args", args),
    )

    result, err := db.exec(ctx, query, args...)

    duration := time.Since(start)
    logger.Debug("Database query completed",
        go_logs.String("query", query),
        go_logs.Int64("duration_ms", duration.Milliseconds()),
        go_logs.Int("rows_affected", result.RowsAffected),
    )

    if err != nil {
        logger.Error("Database query failed",
            go_logs.Err(err),
            go_logs.String("query", query),
        )
    }

    return result, err
}
```

### Worker Job Logging

```go
func (w *Worker) ProcessJob(ctx context.Context, job *Job) error {
    jobLogger := w.logger.With(
        go_logs.String("job_id", job.ID),
        go_logs.String("job_type", job.Type),
        go_logs.Int("attempt", job.Attempt),
    )

    jobLogger.Info("Job started")

    if err := job.Execute(ctx); err != nil {
        jobLogger.Error("Job failed",
            go_logs.Err(err),
            go_logs.Bool("will_retry", job.Attempt < job.MaxAttempts),
        )
        return err
    }

    jobLogger.Info("Job completed",
        go_logs.Int64("duration_ms", time.Since(start).Milliseconds()),
    )
    return nil
}
```

### Error Enrichment

```go
func (s *Service) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    logger.Info("Creating user",
        go_logs.String("email", req.Email),
    )

    user, err := s.repo.Create(ctx, req)
    if err != nil {
        logger.Error("Failed to create user",
            go_logs.Err(err),
            go_logs.String("email", req.Email),
            go_logs.String("error_type", fmt.Sprintf("%T", err)),
        )
        return nil, err
    }

    logger.Info("User created",
        go_logs.String("user_id", user.ID),
        go_logs.String("email", user.Email),
    )
    return user, nil
}
```

## Field Naming Conventions

Follow these conventions for consistent field names:

| Concept | Recommended Field Name | Type |
|---------|------------------------|------|
| User identifier | `user_id` | String |
| Request identifier | `request_id` | String |
| Trace identifier | `trace_id` | String |
| Span identifier | `span_id` | String |
| HTTP method | `method` | String |
| HTTP path | `path` | String |
| HTTP status | `status_code` | Int |
| Response time | `duration_ms` | Int64 |
| Error | `error` | Err |
| Database host | `db_host` | String |
| Database name | `database` | String |
| File path | `file_path` | String |
| File size | `bytes` | Int64 |

## See Also

- [API Reference](API-Reference.md) - Complete field function reference
- [Context and Tracing](Context-and-Tracing.md) - Automatic context fields
- [Examples](Examples.md) - More practical examples
