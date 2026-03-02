# Migration from v2 to v3

This guide helps you migrate from go_logs v2 (legacy global functions) to v3 (modern Logger interface). The good news is that v2 code continues to work - you can migrate gradually or not at all.

## Key Differences

| Feature | v2 | v3 |
|---------|-----|-----|
| API Style | Global functions | Logger interface |
| Structured Fields | No | Yes (typed fields) |
| Child Loggers | No | Yes (`With()`) |
| Context Support | No | Yes (`LogCtx()`) |
| Multiple Instances | No | Yes |
| Hooks | No | Yes |
| Dependency Injection | No | Yes (interface-based) |
| Configuration | Environment only | Options + Environment |
| Initialization | `Init()` required | Auto-initializes |

## Backward Compatibility

**v2 code works without changes in v3:**

```go
// v2 code - still works in v3!
import "github.com/drossan/go_logs"

func main() {
    go_logs.Init()
    go_logs.InfoLog("Hello from v2")
    go_logs.ErrorLog("Something went wrong")
}
```

You can use v2 and v3 APIs together:

```go
import "github.com/drossan/go_logs"

func main() {
    // v2 API
    go_logs.InfoLog("Using v2 API")

    // v3 API
    logger, _ := go_logs.New()
    logger.Info("Using v3 API", go_logs.String("version", "3.0"))
}
```

## Migration Strategies

### Strategy 1: No Migration

Keep using v2 API. It's fully supported and maintained.

```go
// No changes needed
go_logs.Init()
go_logs.InfoLog("Application started")
```

**Pros:**
- No code changes
- Zero risk
- Works immediately

**Cons:**
- No structured logging
- No child loggers
- No context propagation
- No hooks

### Strategy 2: Gradual Migration

Migrate incrementally, starting with new code:

```go
// Old code - keep as is
func legacyFunction() {
    go_logs.InfoLog("Legacy function")
}

// New code - use v3
func newFunction() {
    logger, _ := go_logs.New()
    logger.Info("New function",
        go_logs.String("feature", "structured"),
    )
}
```

### Strategy 3: Full Migration

Completely migrate to v3 API for all new features.

## API Equivalence Table

### Initialization

| v2 | v3 |
|----|-----|
| `go_logs.Init()` | `logger, _ := go_logs.New()` |
| `go_logs.Close()` | `logger.Sync()` |

### Logging Functions

| v2 | v3 |
|----|-----|
| `go_logs.TraceLog(msg)` | `logger.Trace(msg)` |
| `go_logs.DebugLog(msg)` | `logger.Debug(msg)` |
| `go_logs.InfoLog(msg)` | `logger.Info(msg)` |
| `go_logs.SuccessLog(msg)` | `logger.Log(go_logs.SuccessLevel, msg)` |
| `go_logs.WarningLog(msg)` | `logger.Warn(msg)` |
| `go_logs.ErrorLog(msg)` | `logger.Error(msg)` |
| `go_logs.FatalLog(msg)` | `logger.Fatal(msg)` |

### Formatted Logging

| v2 | v3 |
|----|-----|
| `go_logs.Infof(format, args...)` | `logger.Info(fmt.Sprintf(format, args...))` |
| `go_logs.Errorf(format, args...)` | `logger.Error(fmt.Sprintf(format, args...))` |

### Configuration

| v2 (Environment) | v3 (Programmatic) |
|------------------|-------------------|
| `LOG_LEVEL=debug` | `go_logs.WithLevel(go_logs.DebugLevel)` |
| `LOG_FORMAT=json` | `go_logs.WithFormatter(go_logs.NewJSONFormatter())` |
| `SAVE_LOG_FILE=1` | `go_logs.WithRotatingFile(...)` |

## Migration Examples

### Basic Logger

**v2:**
```go
package main

import "github.com/drossan/go_logs"

func main() {
    go_logs.Init()
    go_logs.InfoLog("Application started")
    go_logs.ErrorLog("Something went wrong")
}
```

**v3:**
```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithOutput(os.Stdout),
    )
    logger.Info("Application started")
    logger.Error("Something went wrong")
}
```

### With Structured Fields

**v2 (concatenation):**
```go
go_logs.Infof("User %s logged in from %s", username, ip)
```

**v3 (structured):**
```go
logger.Info("User logged in",
    go_logs.String("username", username),
    go_logs.String("ip", ip),
)
```

### With File Output

**v2:**
```go
// Set environment variables
os.Setenv("SAVE_LOG_FILE", "1")
os.Setenv("LOG_FILE_NAME", "app.log")
os.Setenv("LOG_FILE_PATH", "/var/log")

go_logs.Init()
go_logs.InfoLog("Application started")
```

**v3:**
```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
)
defer logger.Sync()

logger.Info("Application started")
```

### With Error Handling

**v2:**
```go
err := doSomething()
if err != nil {
    go_logs.ErrorLog("Operation failed: " + err.Error())
}
```

**v3:**
```go
err := doSomething()
if err != nil {
    logger.Error("Operation failed",
        go_logs.Err(err),
        go_logs.String("operation", "doSomething"),
    )
}
```

### With Context

**v2 (no context support):**
```go
func handler(w http.ResponseWriter, r *http.Request) {
    // No way to propagate trace ID
    go_logs.InfoLog("Request received")
}
```

**v3:**
```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := go_logs.WithTraceID(r.Context(), r.Header.Get("X-Trace-ID"))

    logger.LogCtx(ctx, go_logs.InfoLevel, "Request received")
}
```

### With Child Loggers

**v2 (repetition):**
```go
func processRequest(requestID string) {
    go_logs.InfoLog("Request " + requestID + " started")
    go_logs.InfoLog("Request " + requestID + " processing")
    go_logs.InfoLog("Request " + requestID + " completed")
}
```

**v3:**
```go
func processRequest(requestID string) {
    reqLogger := logger.With(go_logs.String("request_id", requestID))

    reqLogger.Info("Request started")
    reqLogger.Info("Processing")
    reqLogger.Info("Request completed")
}
```

## Step-by-Step Migration

### Step 1: Identify Logging Code

Find all v2 logging calls:

```bash
grep -r "go_logs\.\(InfoLog\|ErrorLog\|DebugLog\|WarningLog\|FatalLog\|SuccessLog\)" .
```

### Step 2: Create v3 Logger

Create a logger instance:

```go
// In main.go or a dedicated package
var logger go_logs.Logger

func init() {
    var err error
    logger, err = go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    )
    if err != nil {
        panic(err)
    }
}
```

### Step 3: Replace Calls One at a Time

Start with simple calls:

```go
// Before
go_logs.InfoLog("Application started")

// After
logger.Info("Application started")
```

### Step 4: Add Structured Fields

Enhance with context:

```go
// Before
go_logs.InfoLog("User " + username + " logged in")

// After
logger.Info("User logged in",
    go_logs.String("username", username),
)
```

### Step 5: Add Child Loggers

Reduce repetition:

```go
// Before
func processOrder(orderID string) {
    go_logs.InfoLog("Processing order " + orderID)
    // ... many logs with orderID
}

// After
func processOrder(orderID string) {
    orderLogger := logger.With(go_logs.String("order_id", orderID))
    orderLogger.Info("Processing order")
}
```

### Step 6: Add Context Propagation

For HTTP handlers:

```go
// Before
func handler(w http.ResponseWriter, r *http.Request) {
    go_logs.InfoLog("Request received")
}

// After
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := go_logs.WithTraceID(r.Context(), r.Header.Get("X-Trace-ID"))
    logger.LogCtx(ctx, go_logs.InfoLevel, "Request received")
}
```

## Common Migration Patterns

### Service Pattern

Create a logger per service:

```go
type UserService struct {
    logger go_logs.Logger
}

func NewUserService(baseLogger go_logs.Logger) *UserService {
    return &UserService{
        logger: baseLogger.With(go_logs.String("service", "user")),
    }
}

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    s.logger.Info("Creating user", go_logs.String("email", req.Email))
    // ...
}
```

### Middleware Pattern

Create request-scoped loggers:

```go
func loggingMiddleware(logger go_logs.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            requestID := uuid.New().String()

            reqLogger := logger.With(
                go_logs.String("request_id", requestID),
                go_logs.String("method", r.Method),
                go_logs.String("path", r.URL.Path),
            )

            ctx := go_logs.WithRequestID(r.Context(), requestID)

            reqLogger.Info("Request started")
            next.ServeHTTP(w, r.WithContext(ctx))
            reqLogger.Info("Request completed")
        })
    }
}
```

## Troubleshooting

### "undefined: go_logs.New"

Make sure you're using v3:

```bash
go get github.com/drossan/go_logs@latest
```

### Both v2 and v3 Output

If you see double output, you might be using both APIs:

```go
// This logs twice!
go_logs.InfoLog("Message") // v2 logs to stdout
logger.Info("Message")      // v3 logs to stdout
```

**Solution**: Use only one API per log message.

### Missing Fields in v3

Make sure you're passing fields correctly:

```go
// Wrong - fields are ignored
logger.Info("Message", "key", "value")

// Correct - use typed field constructors
logger.Info("Message", go_logs.String("key", "value"))
```

### Init() Still Required?

No, v3 auto-initializes. But v2 still requires `Init()`:

```go
// v2 still needs Init()
go_logs.Init()
go_logs.InfoLog("v2 message")

// v3 doesn't need Init()
logger, _ := go_logs.New()
logger.Info("v3 message")
```

## Checklist

- [ ] Identify all v2 logging calls
- [ ] Create v3 logger instance
- [ ] Replace simple logging calls
- [ ] Add structured fields where appropriate
- [ ] Create child loggers for repeated context
- [ ] Add context propagation for HTTP handlers
- [ ] Update tests to use v3
- [ ] Remove `go_logs.Init()` calls (no longer needed for v3)
- [ ] Update documentation

## See Also

- [API Reference](API-Reference.md) - Complete v3 API
- [Getting Started](Getting-Started.md) - Quick start guide
- [Examples](Examples.md) - More v3 examples
