# Configuration

go_logs can be configured through environment variables, programmatic options, or a combination of both.

## Table of Contents

- [Environment Variables](#environment-variables)
- [Programmatic Configuration](#programmatic-configuration)
- [Configuration by Environment](#configuration-by-environment)
- [Sensitive Data Redaction](#sensitive-data-redaction)
- [Complete Examples](#complete-examples)

---

## Environment Variables

### Core Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `LOG_LEVEL` | Minimum log level | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | Output format | `text` | `text`, `json` |

### File Logging (v2 Compatibility)

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `SAVE_LOG_FILE` | Enable file logging | `0` | `1` |
| `LOG_FILE_NAME` | Log file name | `log.txt` | `app.log` |
| `LOG_FILE_PATH` | Log file directory | current dir | `/var/log/myapp` |

### File Rotation

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `LOG_MAX_SIZE` | Max file size in MB | `100` | `100` |
| `LOG_MAX_BACKUPS` | Max backup files | `5` | `10` |

### Slack Notifications (v2)

| Variable | Description | Default |
|----------|-------------|---------|
| `NOTIFICATIONS_SLACK_ENABLED` | Enable Slack notifications | `0` |
| `SLACK_TOKEN` | Slack bot token | - |
| `SLACK_CHANNEL_ID` | Slack channel ID | - |
| `NOTIFICATION_FATAL_LOG` | Send fatal logs to Slack | `0` |
| `NOTIFICATION_ERROR_LOG` | Send error logs to Slack | `0` |
| `NOTIFICATION_WARNING_LOG` | Send warning logs to Slack | `0` |
| `NOTIFICATION_INFO_LOG` | Send info logs to Slack | `0` |
| `NOTIFICATION_SUCCESS_LOG` | Send success logs to Slack | `0` |

### Log Levels Reference

| Level | Value | Description |
|-------|-------|-------------|
| `trace` | 10 | Extremely detailed |
| `debug` | 20 | Diagnostic information |
| `info` | 30 | General operational |
| `warn`, `warning` | 40 | Potential issues |
| `error` | 50 | Errors |
| `fatal` | 60 | Critical errors |
| `silent`, `none`, `disable` | 0 | Disable all logging |

### Setting Environment Variables

**Linux/macOS:**
```bash
export LOG_LEVEL=debug
export LOG_FORMAT=json
export SAVE_LOG_FILE=1
export LOG_FILE_NAME=app.log
export LOG_FILE_PATH=/var/log/myapp
```

**Windows (PowerShell):**
```powershell
$env:LOG_LEVEL = "debug"
$env:LOG_FORMAT = "json"
```

**Docker Compose:**
```yaml
services:
  app:
    image: myapp
    environment:
      - LOG_LEVEL=info
      - LOG_FORMAT=json
      - SAVE_LOG_FILE=1
      - LOG_FILE_NAME=app.log
      - LOG_FILE_PATH=/var/log
```

**Kubernetes:**
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: myapp
spec:
  containers:
  - name: app
    image: myapp
    env:
    - name: LOG_LEVEL
      value: "info"
    - name: LOG_FORMAT
      value: "json"
```

---

## Programmatic Configuration

### Option Pattern

Configure the logger using functional options:

```go
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.InfoLevel),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithOutput(os.Stdout),
    go_logs.WithCaller(true),
    go_logs.WithCommonRedaction(),
)
```

### Available Options

#### WithLevel

Sets the minimum log level.

```go
go_logs.WithLevel(go_logs.InfoLevel)
```

#### WithOutput

Sets the output destination.

```go
go_logs.WithOutput(os.Stdout)
go_logs.WithOutput(file) // Any io.Writer
```

#### WithFormatter

Sets the log formatter.

```go
go_logs.WithFormatter(go_logs.NewTextFormatter())
go_logs.WithFormatter(go_logs.NewJSONFormatter())
```

#### WithHooks

Adds hooks for custom processing.

```go
go_logs.WithHooks(myHook1, myHook2)
```

#### WithCaller

Enables caller information (file:line function).

```go
go_logs.WithCaller(true)
```

**Output:**
```
[2026/02/28 10:30:00] INFO main.go:42 main.processRequest Processing request
```

#### WithStackTrace

Enables stack trace capture.

```go
go_logs.WithStackTrace(true)
go_logs.WithStackTraceLevel(go_logs.WarnLevel) // Capture for Warn and above
```

#### WithCallerSkip

Adjusts stack frame skipping for wrapper functions.

```go
go_logs.WithCallerSkip(3)
```

#### WithRedactor

Enables redaction of specified fields.

```go
go_logs.WithRedactor("password", "token", "api_key")
```

#### WithCommonRedaction

Enables redaction of common sensitive fields.

```go
go_logs.WithCommonRedaction()
```

Redacts: password, passwd, pwd, token, api_key, apikey, api-key, secret, authorization, auth, cookie, session, credit_card, ssn, social_security

#### WithMultiOutput

Output to multiple writers.

```go
file, _ := go_logs.NewRotatingFileWriter("app.log", 100, 5)
go_logs.WithMultiOutput(file, os.Stdout)
```

#### WithRotatingFile

Creates a rotating file writer.

```go
go_logs.WithRotatingFile("/var/log/app.log", 100, 5)
```

#### WithRotatingFileEnhanced

Creates a rotating file with full configuration.

```go
go_logs.WithRotatingFileEnhanced(go_logs.RotatingFileConfig{
    Filename:     "/var/log/app.log",
    MaxSizeMB:    100,
    MaxBackups:   5,
    RotationType: go_logs.RotateDaily,
    Compress:     true,
    MaxAge:       30,
})
```

---

## Configuration by Environment

### Development

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func NewLogger() go_logs.Logger {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.DebugLevel),
        go_logs.WithFormatter(go_logs.NewTextFormatter()), // Colors
        go_logs.WithOutput(os.Stdout),
        go_logs.WithCaller(true), // Show file:line
    )
    return logger
}
```

### Staging

```go
func NewLogger() go_logs.Logger {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithRotatingFile("/var/log/app.log", 50, 10),
    )
    return logger
}
```

### Production

```go
func NewLogger() go_logs.Logger {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithRotatingFileEnhanced(go_logs.RotatingFileConfig{
            Filename:     "/var/log/app.log",
            MaxSizeMB:    100,
            MaxBackups:   30,
            RotationType: go_logs.RotateDaily,
            Compress:     true,
            MaxAge:       7,
        }),
        go_logs.WithCommonRedaction(),
        go_logs.WithStackTraceLevel(go_logs.ErrorLevel),
    )
    return logger
}
```

### Environment-Based Factory

```go
func NewLogger() go_logs.Logger {
    env := os.Getenv("ENV")

    switch env {
    case "production":
        return newProductionLogger()
    case "staging":
        return newStagingLogger()
    default:
        return newDevelopmentLogger()
    }
}

func newDevelopmentLogger() go_logs.Logger {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.DebugLevel),
        go_logs.WithFormatter(go_logs.NewTextFormatter()),
        go_logs.WithCaller(true),
    )
    return logger
}

func newProductionLogger() go_logs.Logger {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithRotatingFile("/var/log/app.log", 100, 30),
        go_logs.WithCommonRedaction(),
    )
    return logger
}
```

---

## Sensitive Data Redaction

### Built-in Redaction

Use `WithCommonRedaction()` for common sensitive fields:

```go
logger, _ := go_logs.New(
    go_logs.WithCommonRedaction(),
)

logger.Info("User login",
    go_logs.String("username", "john"),
    go_logs.String("password", "secret123"), // Will be redacted
    go_logs.String("token", "abc-123"),       // Will be redacted
)
```

**Output:**
```
[2026/02/28 10:30:00] INFO User login username=john password=*** token=***
```

### Custom Redaction

Specify custom fields to redact:

```go
logger, _ := go_logs.New(
    go_logs.WithRedactor("password", "token", "credit_card", "ssn"),
)
```

### Redacted Fields List

Common fields redacted by `WithCommonRedaction()`:

| Category | Fields |
|----------|--------|
| Passwords | password, passwd, pwd |
| Tokens | token, api_key, apikey, api-key |
| Secrets | secret, authorization, auth |
| Session | cookie, session |
| Financial | credit_card, ssn, social_security |

---

## Complete Examples

### Simple Console Logger

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Info("Application started")
}
```

### File Logger with Rotation

```go
package main

import (
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    )
    defer logger.Sync()

    logger.Info("Application started")
}
```

### Multi-Output Logger

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    file, _ := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)

    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithMultiOutput(file, os.Stdout),
    )
    defer logger.Sync()

    logger.Info("Application started")
}
```

### Logger with Hooks

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Create metrics hook
    metricsHook := go_logs.NewFuncHook(func(entry *go_logs.Entry) error {
        if entry.Level >= go_logs.ErrorLevel {
            // Increment error counter
            errorCounter.Inc()
        }
        return nil
    })

    logger, _ := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
        go_logs.WithHooks(metricsHook),
    )

    logger.Info("Application started")
}
```

### Full Production Configuration

```go
package main

import (
    "os"

    "github.com/drossan/go_logs"
)

func main() {
    // Create file writer with rotation
    fileWriter, err := go_logs.NewRotatingFileWriter("/var/log/myapp/app.log", 100, 30)
    if err != nil {
        panic(err)
    }

    // Create alert hook for errors
    alertHook := go_logs.NewFuncHook(func(entry *go_logs.Entry) error {
        if entry.Level >= go_logs.ErrorLevel {
            // Send alert asynchronously
            go sendAlert(entry.Message, entry.Fields)
        }
        return nil
    })

    // Create logger with full configuration
    logger, err := go_logs.New(
        go_logs.WithLevel(go_logs.InfoLevel),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithMultiOutput(fileWriter, os.Stdout),
        go_logs.WithCaller(true),
        go_logs.WithStackTraceLevel(go_logs.ErrorLevel),
        go_logs.WithCommonRedaction(),
        go_logs.WithHooks(alertHook),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Info("Application started",
        go_logs.String("version", "1.0.0"),
        go_logs.String("environment", os.Getenv("ENV")),
    )

    // Application logic...
}

func sendAlert(message string, fields []go_logs.Field) {
    // Implement alerting logic (Slack, PagerDuty, etc.)
}
```

---

## Runtime Configuration

### Changing Log Level at Runtime

```go
logger, _ := go_logs.New()

// Start with info level
logger.Info("Starting with info level")

// Change to debug level
logger.SetLevel(go_logs.DebugLevel)
logger.Debug("Now debug is enabled")

// Check current level
if logger.GetLevel() >= go_logs.DebugLevel {
    // Debug is enabled
}
```

### Environment Variable Reloading

For dynamic configuration, watch for changes:

```go
func watchConfig(logger go_logs.Logger) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        level := os.Getenv("LOG_LEVEL")
        if level != "" {
            logger.SetLevel(go_logs.ParseLevel(level))
        }
    }
}
```

---

## See Also

- [API Reference](API-Reference.md) - All configuration options
- [Formatters](Formatters.md) - Formatter configuration
- [File Rotation](File-Rotation.md) - Rotation configuration
- [Hooks](Hooks.md) - Hook configuration
