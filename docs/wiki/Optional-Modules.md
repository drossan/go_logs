# Optional Modules (Hybrid Architecture)

go_logs v3.5 introduces a **hybrid architecture** with the core in the main package and optional features as submodules.

## Architecture Overview

```
go_logs/
├── Core (always included)
│   ├── logger.go, logger_impl.go
│   ├── metrics.go              ← NEW: Zero-overhead metrics
│   └── ... (other core files)
│
├── async/                       ← Opt-in
│   └── async.go                ← Non-blocking logging
│
├── http/                        ← Opt-in
│   └── dynamic_level.go        ← HTTP log level control
│
└── signal/                      ← Opt-in
    └── signal.go               ← SIGHUP handler
```

## Metrics (Core - Always Enabled)

Zero-overhead logging statistics available on every logger:

```go
logger, _ := go_logs.New()
metrics := logger.GetMetrics()

// Available metrics
fmt.Printf("Total logs: %d\n", metrics.Total())
fmt.Printf("Info logs: %d\n", metrics.Count(go_logs.InfoLevel))
fmt.Printf("Errors: %d\n", metrics.Count(go_logs.ErrorLevel))
fmt.Printf("Dropped: %d\n", metrics.Dropped())

// Get a snapshot for monitoring
snapshot := metrics.Snapshot()
// snapshot.Total
// snapshot.ByLevel[go_logs.InfoLevel]

// Reset for testing
metrics.Reset()
```

### Thread-Safety

All metrics operations use atomic operations, making them safe for concurrent use with zero lock overhead.

### Integration with Monitoring

```go
// Expose metrics to Prometheus
http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    metrics := logger.GetMetrics()
    snapshot := metrics.Snapshot()

    fmt.Fprintf(w, "# HELP go_logs_total Total log entries\n")
    fmt.Fprintf(w, "# TYPE go_logs_total counter\n")
    fmt.Fprintf(w, "go_logs_total %d\n", snapshot.Total)

    for level, count := range snapshot.ByLevel {
        fmt.Fprintf(w, "go_logs_by_level{level=\"%s\"} %d\n", level.String(), count)
    }

    fmt.Fprintf(w, "go_logs_dropped %d\n", metrics.Dropped())
})
```

## Async Logging (Submodule - Opt-in)

Non-blocking logging for high-throughput applications:

```go
import "github.com/drossan/go_logs/async"

// Create base sync logger
syncLogger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

// Wrap with async (buffer size 1000)
asyncLogger := async.Wrap(syncLogger, 1000)
defer asyncLogger.Sync()

// Non-blocking logging
asyncLogger.Info("Server started", go_logs.Int("port", 8080))
```

### Configuration

```go
asyncLogger := async.WrapWithConfig(syncLogger, async.Config{
    BufferSize:      10000,              // Buffer capacity
    ShutdownTimeout: 10 * time.Second,   // Max wait on Sync()
})
```

### Behavior

- **Non-blocking**: Log calls return immediately
- **Drop-on-overflow**: If buffer is full, logs are dropped (counted in metrics)
- **Graceful shutdown**: `Sync()` waits for all pending logs
- **Shared metrics**: Dropped logs increment the shared metrics counter

### When to Use

- High-throughput applications (>10k logs/sec)
- Logging to slow outputs (network, remote services)
- When you can tolerate occasional log loss for performance

### When NOT to Use

- Financial/audit logging where every log must be persisted
- Low-throughput applications (<1k logs/sec)
- When debugging issues where logs might be lost

## Dynamic Level via HTTP (Submodule - Opt-in)

Change log levels at runtime via HTTP endpoints:

```go
import httplogs "github.com/drossan/go_logs/http"

logger, _ := go_logs.New(go_logs.WithLevel(go_logs.InfoLevel))

handler := httplogs.NewDynamicLevelHandler(logger, httplogs.Config{
    Endpoint:  "/debug/level",
    AuthToken: "secret-token",
    RateLimit: 10,  // requests per second
    AllowedIPs: []string{"10.0.0.0/8", "192.168.1.100"},
})

http.Handle("/debug/", handler)
```

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/debug/level` | Get current log level |
| PUT | `/debug/level` | Set log level |
| GET | `/debug/level/metrics` | Get logging metrics |

### Usage Examples

```bash
# Get current level
curl -H "Authorization: Bearer secret-token" http://localhost:8080/debug/level
# Response: {"level": "INFO", "timestamp": "2026-03-01T10:00:00Z"}

# Set to debug
curl -X PUT -H "Authorization: Bearer secret-token" \
  -H "Content-Type: application/json" \
  -d '{"level": "debug"}' \
  http://localhost:8080/debug/level
# Response: {"level": "DEBUG", "timestamp": "2026-03-01T10:01:00Z"}

# Get metrics
curl -H "Authorization: Bearer secret-token" http://localhost:8080/debug/level/metrics
# Response: {"total": 1234, "by_level": {"INFO": 1000, "ERROR": 234}, "dropped": 0}
```

### Security Features

- **Bearer token authentication**: Required if `AuthToken` is set
- **Rate limiting**: Prevents abuse (requests per second)
- **IP whitelisting**: Exact IPs or CIDR ranges

### Use Cases

- Debug production issues without redeploying
- Temporarily increase logging during incidents
- Reduce logging during peak hours

## SIGHUP Handler (Submodule - Opt-in)

Rotate logs when receiving system signals:

```go
import "github.com/drossan/go_logs/signal"

writer, _ := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
logger, _ := go_logs.New(go_logs.WithOutput(writer))

// Create handler for SIGHUP
handler := signal.NewSIGHUPHandler(signal.WrapRotator(writer))
handler.Register()
defer handler.Stop()

// Logs will rotate when SIGHUP is received
```

### Integration with logrotate

Create `/etc/logrotate.d/myapp`:

```
/var/log/app.log {
    daily
    rotate 30
    compress
    missingok
    notifempty
    postrotate
        kill -HUP $(cat /var/run/myapp.pid)
    endscript
}
```

### Custom Signals

```go
// Handle multiple signals
handler := signal.NewSIGHUPHandler(rotator, syscall.SIGUSR1, syscall.SIGUSR2)
```

## Summary

| Feature | Package | Overhead | Use Case |
|---------|---------|----------|----------|
| **Metrics** | Core | Zero (atomic) | Monitoring & observability |
| **Async** | `async/` | Goroutine + channel | High throughput |
| **HTTP** | `http/` | HTTP endpoint | Debugging production |
| **Signal** | `signal/` | Signal handler | logrotate integration |
