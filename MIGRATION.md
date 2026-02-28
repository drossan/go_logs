# go_logs v3 Migration Guide

This guide helps you migrate from `go_logs` v2 to v3. The v3 release introduces significant improvements while maintaining **100% backward compatibility** with v2.

## Table of Contents

- [What's New in v3](#whats-new-in-v3)
- [Migration Strategy](#migration-strategy)
- [Quick Start](#quick-start)
- [Step-by-Step Migration](#step-by-step-migration)
- [API Reference](#api-reference)
- [Configuration Changes](#configuration-changes)
- [Examples](#examples)

---

## What's New in v3

### Major Features

1. **Structured Logging** - Type-safe fields instead of printf-style formatting
2. **Child Loggers** - Pre-configured loggers with inherited fields
3. **Context Support** - Automatic trace_id extraction from context
4. **Dual Formatters** - TextFormatter (dev) and JSONFormatter (prod)
5. **File Rotation** - Automatic log rotation by size with backup management
6. **Extensible Hooks** - Pluggable system for Slack, Sentry, metrics, etc.
7. **Data Redaction** - Automatic masking of sensitive fields (passwords, tokens)
8. **Runtime Level Control** - Change log level without restarting
9. **Multiple Instances** - Create independent loggers with different configs

### Backward Compatibility

**All v2 code continues to work without changes!**

```go
// v2 code - works in v3 without modification
go_logs.Init()
defer go_logs.Close()

go_logs.InfoLog("Server started")
go_logs.Errorf("Connection failed: %v", err)
```

---

## Migration Strategy

We recommend a **3-phase gradual migration**:

### Phase 1: Drop-in Replacement (No Code Changes)
- Update dependency to v3
- Run existing tests - they should pass
- Deploy and monitor

### Phase 2: Adopt New Features Incrementally
- Use v3 API for new code
- Migrate critical paths to structured logging
- Add child loggers for request tracing
- Enable JSON formatter for production

### Phase 3: Full Migration (Optional)
- Replace all v2 functions with v3 API
- Use structured fields throughout
- Remove unused v2 functions

---

## Quick Start

### For New Projects

Use v3 API from the start:

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    // Create logger with structured fields
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    )

    // Log with structured fields
    logger.Info("Server started",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
        go_logs.Bool("debug", false),
    )
}
```

### For Existing Projects

No immediate changes required - v2 code works as-is:

```go
// This continues to work in v3
go_logs.Init()
defer go_logs.Close()

go_logs.InfoLog("Application started")
go_logs.Errorf("Error: %v", err)
```

---

## Step-by-Step Migration

### Step 1: Update Dependency

```bash
# Update to v3
go get -u github.com/drossan/go_logs@v3
go mod tidy
```

### Step 2: Verify Existing Tests

```bash
# Run your existing test suite
go test ./...

# All v2 tests should pass without changes
```

### Step 3: Migrate to v3 API (Gradual)

#### 3.1 Replace Printf-Style Logging

**Before (v2):**
```go
go_logs.Infof("User %s (ID: %d) logged in from %s", username, userID, ip)
```

**After (v3):**
```go
logger.Info("user_logged_in",
    go_logs.String("username", username),
    go_logs.Int("user_id", userID),
    go_logs.String("ip", ip),
)
```

**Benefits:**
- Structured fields are queryable in log aggregators (ELK, Splunk)
- Type-safe (compile-time checks)
- No format string errors

#### 3.2 Use Child Loggers for Context

**Before (v2):**
```go
go_logs.Infof("Processing request for user %s", username)
// ... many lines of code ...
go_logs.Infof("Request completed successfully")
```

**After (v3):**
```go
// Create child logger with request context
reqLogger := logger.With(
    go_logs.String("request_id", reqID),
    go_logs.String("user_id", userID),
)

reqLogger.Info("Processing request")
// ... many lines of code ...
reqLogger.Info("Request completed successfully")
```

**Benefits:**
- All logs automatically include request_id and user_id
- No need to pass context manually
- Better traceability

#### 3.3 Use Context for Distributed Tracing

**Before (v2):**
```go
// Context accepted but ignored
go_logs.InfoLogCtx(ctx, "Processing request")
```

**After (v3):**
```go
// Add trace_id to context
ctx = go_logs.WithTraceID(ctx, "trace-abc-123")

// Trace_id automatically extracted and logged
logger.LogCtx(ctx, go_logs.InfoLevel, "Processing request")
// Output: INFO Processing request trace_id=trace-abc-123
```

#### 3.4 Enable JSON Formatter for Production

**Before (v2):**
```go
// Only text format available
go_logs.InfoLog("Server started")
// Output: [INFO] Server started
```

**After (v3):**
```go
// Set environment variable
export LOG_FORMAT=json

// Or configure programmatically
logger, _ := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)

logger.Info("Server started", go_logs.Int("port", 8080))
// Output: {"level":"INFO","message":"Server started","timestamp":"2025-02-28T12:00:00Z","fields":{"port":8080}}
```

#### 3.5 Use Rotating File Writer

**Before (v2):**
```bash
# File grows indefinitely
export SAVE_LOG_FILE=1
export LOG_FILE_NAME=app.log
```

**After (v3):**
```bash
# Automatic rotation by size
export SAVE_LOG_FILE=1
export LOG_FILE_NAME=app.log
export LOG_MAX_SIZE=100      # Rotate at 100MB
export LOG_MAX_BACKUPS=3     # Keep 3 backups
```

**Result:**
- `app.log` - Current log file
- `app.log.1` - First backup
- `app.log.2` - Second backup
- `app.log.3` - Third backup
- Older backups automatically deleted

#### 3.6 Enable Data Redaction

**Before (v2):**
```go
// Risk: Passwords logged in plain text
go_logs.Infof("User login: %s, password: %s", username, password)
```

**After (v3):**
```go
// Enable redaction
logger, _ := go_logs.New(
    go_logs.WithCommonRedaction(),  // Masks password, token, secret, etc.
)

logger.Info("User login",
    go_logs.String("username", username),
    go_logs.String("password", password),  // Automatically masked
)
// Output: INFO User login username=john password=***
```

---

## API Reference

### v2 API (Backward Compatible)

All v2 functions continue to work:

```go
// Basic logging
go_logs.InfoLog(message)
go_logs.ErrorLog(message)
go_logs.WarningLog(message)
go_logs.SuccessLog(message)
go_logs.FatalLog(message)  // Exits

// Formatted logging
go_logs.Infof(format, args...)
go_logs.Errorf(format, args...)
go_logs.Warningf(format, args...)
go_logs.Successf(format, args...)
go_logs.Fatalf(format, args...)  // Exits

// Context support (was no-op, now extracts trace_id)
go_logs.InfoLogCtx(ctx, message)
go_logs.ErrorLogCtx(ctx, message)
go_logs.InfoLogCtxf(ctx, format, args...)

// Initialization
go_logs.Init()   // Load config from env vars
go_logs.Close()  // Flush and close files
```

### v3 API (New)

#### Creating Loggers

```go
// Create logger with options
logger, err := go_logs.New(
    go_logs.WithLevel(go_logs.InfoLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithOutput(os.Stdout),
    go_logs.WithHooks(myHook),
    go_logs.WithRedactor("password", "token"),
)
```

#### Structured Logging

```go
// Field constructors
go_logs.String(key, value)
go_logs.Int(key, value)
go_logs.Int64(key, value)
go_logs.Float64(key, value)
go_logs.Bool(key, value)
go_logs.Err(err)           // Shorthand for errors
go_logs.Any(key, value)    // Arbitrary types

// Logging methods
logger.Trace(msg, fields...)
logger.Debug(msg, fields...)
logger.Info(msg, fields...)
logger.Warn(msg, fields...)
logger.Error(msg, fields...)
logger.Fatal(msg, fields...)  // Exits

// With level
logger.Log(go_logs.ErrorLevel, msg, fields...)
```

#### Child Loggers

```go
// Create child logger with inherited fields
childLogger := logger.With(
    go_logs.String("service", "api"),
    go_logs.String("version", "1.0"),
)

// All logs include service and version
childLogger.Info("Request received")
```

#### Context Support

```go
// Add trace_id to context
ctx := go_logs.WithTraceID(ctx, "trace-abc-123")
ctx = go_logs.WithRequestID(ctx, "req-xyz-789")

// Log with context - extracts trace_id automatically
logger.LogCtx(ctx, go_logs.InfoLevel, "Processing")
```

#### Runtime Configuration

```go
// Change log level at runtime
logger.SetLevel(go_logs.DebugLevel)
currentLevel := logger.GetLevel()

// Sync buffers
logger.Sync()
```

---

## Configuration Changes

### Environment Variables

#### v2 Variables (Still Supported)

```bash
# File logging
SAVE_LOG_FILE=1
LOG_FILE_NAME=app.log
LOG_FILE_PATH=/var/log

# Slack notifications (legacy)
NOTIFICATIONS_SLACK_ENABLED=1
NOTIFICATION_FATAL_LOG=1
NOTIFICATION_ERROR_LOG=1
NOTIFICATION_WARNING_LOG=1
NOTIFICATION_INFO_LOG=1
NOTIFICATION_SUCCESS_LOG=1

# Slack credentials
SLACK_TOKEN=xoxb-...
SLACK_CHANNEL_ID=C1234567890
```

#### v3 New Variables

```bash
# Log level (replaces NOTIFICATION_*_LOG)
LOG_LEVEL=info          # trace, debug, info, warn, error, fatal, silent

# Output format
LOG_FORMAT=json         # text (default) or json

# File rotation
LOG_MAX_SIZE=100        # MB (default: 100)
LOG_MAX_BACKUPS=3       # Number of backups (default: 3)

# Data redaction
LOG_REDACT_KEYS=password,token,secret,api_key
```

### Backward Compatibility

The legacy notification system takes precedence when both are configured:

```bash
# Legacy system (NOTIFICATION_*_LOG) takes precedence
export NOTIFICATION_ERROR_LOG=1
export LOG_LEVEL=info

# Result: Only ERROR logs sent to Slack (legacy behavior)

# To use new system, don't set NOTIFICATION_*_LOG
export LOG_LEVEL=error

# Result: ERROR and FATAL logs sent to Slack (numeric threshold)
```

---

## Examples

### Example 1: Web Server with Structured Logging

```go
package main

import (
    "context"
    "net/http"
    "os"

    "github.com/drossan/go_logs"
)

func main() {
    // Initialize logger
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    )

    // Create child logger for HTTP server
    httpLogger := logger.With(
        go_logs.String("service", "http"),
        go_logs.String("version", "1.0"),
    )

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Add trace_id to context
        ctx := go_logs.WithTraceID(r.Context(), generateTraceID())

        // Log request with trace_id
        httpLogger.LogCtx(ctx, go_logs.InfoLevel, "Request received",
            go_logs.String("method", r.Method),
            go_logs.String("path", r.URL.Path),
        )

        w.Write([]byte("OK"))
    })

    httpLogger.Info("Server started",
        go_logs.String("addr", ":8080"),
    )

    http.ListenAndServe(":8080", nil)
}

func generateTraceID() string {
    return "trace-" + randomString()
}
```

### Example 2: Multiple Loggers

```go
package main

import (
    "os"

    "github.com/drossan/go_logs"
)

func main() {
    // Access logger - JSON format to file
    accessLogger, _ := go_logs.New(
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(createFile("access.log")),
        go_logs.WithLevel(go_logs.InfoLevel),
    )

    // Error logger - Text format to stderr
    errorLogger, _ := go_logs.New(
        go_logs.WithFormatter(go_logs.NewTextFormatter()),
        go_logs.WithOutput(os.Stderr),
        go_logs.WithLevel(go_logs.ErrorLevel),
    )

    // Use different loggers for different purposes
    accessLogger.Info("API request",
        go_logs.String("endpoint", "/api/users"),
        go_logs.Int("status", 200),
    )

    errorLogger.Error("Database connection failed",
        go_logs.String("host", "db.example.com"),
        go_logs.Err(err),
    )
}
```

### Example 3: Migration from v2 to v3

**Before (v2):**
```go
package main

import "github.com/drossan/go_logs"

func handleRequest(userID int, username string) {
    go_logs.Infof("User %s (ID: %d) logged in", username, userID)

    if err := processUser(userID); err != nil {
        go_logs.Errorf("Failed to process user %d: %v", userID, err)
        return
    }

    go_logs.Infof("User %d processed successfully", userID)
}
```

**After (v3) - Step 1 (Drop-in):**
```go
package main

import "github.com/drossan/go_logs"

func handleRequest(userID int, username string) {
    // No code changes - works as-is
    go_logs.Infof("User %s (ID: %d) logged in", username, userID)

    if err := processUser(userID); err != nil {
        go_logs.Errorf("Failed to process user %d: %v", userID, err)
        return
    }

    go_logs.Infof("User %d processed successfully", userID)
}
```

**After (v3) - Step 2 (Gradual):**
```go
package main

import "github.com/drossan/go_logs"

var logger, _ = go_logs.New(
    go_logs.WithLevel(go_logs.InfoLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)

func handleRequest(userID int, username string) {
    // Mix of v2 and v3 API
    logger.Info("user_logged_in",
        go_logs.String("username", username),
        go_logs.Int("user_id", userID),
    )

    if err := processUser(userID); err != nil {
        // Still using v2 for error handling
        go_logs.Errorf("Failed to process user %d: %v", userID, err)
        return
    }

    logger.Info("user_processed",
        go_logs.Int("user_id", userID),
    )
}
```

**After (v3) - Step 3 (Full):**
```go
package main

import "github.com/drossan/go_logs"

var logger, _ = go_logs.New(
    go_logs.WithLevel(go_logs.InfoLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)

func handleRequest(userID int, username string) {
    // Fully migrated to v3
    logger.Info("user_logged_in",
        go_logs.String("username", username),
        go_logs.Int("user_id", userID),
    )

    if err := processUser(userID); err != nil {
        logger.Error("user_processing_failed",
            go_logs.Int("user_id", userID),
            go_logs.Err(err),
        )
        return
    }

    logger.Info("user_processed",
        go_logs.Int("user_id", userID),
    )
}
```

---

## Troubleshooting

### Issue: Logs not appearing

**v2:**
```bash
# Ensure Init() is called
go_logs.Init()
```

**v3:**
```bash
# Ensure logger is created
logger, _ := go_logs.New()
# Or use Init() for default logger
go_logs.Init()
```

### Issue: Context not being used

**v2:**
```go
// Context was accepted but ignored
go_logs.InfoLogCtx(ctx, "message")  // No trace_id in output
```

**v3:**
```go
// Must add trace_id to context first
ctx = go_logs.WithTraceID(ctx, "trace-123")
logger.LogCtx(ctx, go_logs.InfoLevel, "message")  // Includes trace_id
```

### Issue: Wrong log level

**Check configuration:**
```bash
# v2: Check NOTIFICATION_*_LOG vars
echo $NOTIFICATION_ERROR_LOG

# v3: Check LOG_LEVEL
echo $LOG_LEVEL
```

### Issue: File not rotating

**Check configuration:**
```bash
# Ensure LOG_MAX_SIZE is set
echo $LOG_MAX_SIZE  # Should be > 0

# Check file permissions
ls -la /path/to/log.log
```

---

## Best Practices

### 1. Use Structured Fields

❌ **Bad:**
```go
logger.Infof("User %s logged in from %s at %s", user, ip, time)
```

✅ **Good:**
```go
logger.Info("user_logged_in",
    go_logs.String("username", user),
    go_logs.String("ip", ip),
    go_logs.String("timestamp", time.Format(time.RFC3339)),
)
```

### 2. Use Child Loggers for Context

❌ **Bad:**
```go
logger.Info("Request", go_logs.String("request_id", reqID))
logger.Info("Processing", go_logs.String("request_id", reqID))
logger.Info("Complete", go_logs.String("request_id", reqID))
```

✅ **Good:**
```go
reqLogger := logger.With(go_logs.String("request_id", reqID))
reqLogger.Info("Request")
reqLogger.Info("Processing")
reqLogger.Info("Complete")
```

### 3. Enable Redaction for Sensitive Data

❌ **Bad:**
```go
logger.Info("Login",
    go_logs.String("username", user),
    go_logs.String("password", pass),  // Password in plain text!
)
```

✅ **Good:**
```go
logger, _ := go_logs.New(go_logs.WithCommonRedaction())
logger.Info("Login",
    go_logs.String("username", user),
    go_logs.String("password", pass),  // Automatically masked as ***
)
```

### 4. Use JSON Formatter in Production

❌ **Bad:**
```go
# Text format hard to parse
[INFO] 2025-02-28 12:00:00 User john logged in from 192.168.1.1
```

✅ **Good:**
```bash
export LOG_FORMAT=json
# {"level":"INFO","message":"User logged in","fields":{"username":"john","ip":"192.168.1.1"}}
```

### 5. Use Error Field for Errors

❌ **Bad:**
```go
logger.Error("Database failed",
    go_logs.String("error", err.Error()),  // Loses error context
)
```

✅ **Good:**
```go
logger.Error("Database failed",
    go_logs.Err(err),  // Preserves stack trace and context
)
```

---

## FAQ

### Q: Do I need to change my code to use v3?

**A:** No! v3 is 100% backward compatible with v2. Your existing code will work without changes.

### Q: Should I migrate to the v3 API?

**A:** We recommend gradual migration:
- Keep v2 code for existing functionality
- Use v3 API for new features
- Migrate critical paths to structured logging
- Migrate remaining code over time

### Q: Will v2 functions be removed?

**A:** No. v2 functions are permanently supported for backward compatibility.

### Q: Can I mix v2 and v3 APIs?

**A:** Yes! You can use both APIs in the same application:
```go
// v2 API
go_logs.InfoLog("Legacy code")

// v3 API
logger.Info("New code", go_logs.String("key", "value"))
```

### Q: How do I enable JSON logging?

**A:** Set environment variable:
```bash
export LOG_FORMAT=json
```

Or configure programmatically:
```go
logger, _ := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### Q: How does file rotation work?

**A:** Configure via environment variables:
```bash
export LOG_MAX_SIZE=100    # Rotate at 100MB
export LOG_MAX_BACKUPS=3   # Keep 3 backups
```

Rotation happens automatically when file size exceeds LOG_MAX_SIZE.

### Q: What's the performance impact of v3?

**A:** v3 is designed for high performance:
- Structured fields: < 10ns per field
- Fast-path filtering: < 5ns
- No significant overhead vs v2 for simple cases
- Benchmarks show comparable performance

---

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/drossan/go_logs/issues
- Documentation: https://pkg.go.dev/github.com/drossan/go_logs
- Migration examples: See `example_v3_test.go`

---

**Happy logging! 🚀**
