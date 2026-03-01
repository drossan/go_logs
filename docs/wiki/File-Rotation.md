# File Rotation

go_logs provides a built-in `RotatingFileWriter` for automatic log file rotation by size. This implementation has zero external dependencies and uses only the Go standard library.

## Overview

`RotatingFileWriter` automatically rotates log files when they reach a specified size, creating backup files with incrementing suffixes.

### Rotation Behavior

When rotation occurs:
1. The current file is closed
2. Existing backups are renamed (`.3` -> `.4`, `.2` -> `.3`, `.1` -> `.2`)
3. The current file is renamed to `.1`
4. A new empty file is created
5. Old backups beyond `maxBackups` are deleted

### Example File Layout

With `maxBackups=3`:
```
app.log        # Current log file
app.log.1      # Most recent backup
app.log.2      # Second most recent
app.log.3      # Oldest backup (third rotation)
```

## Creating a RotatingFileWriter

### NewRotatingFileWriter

Creates a rotating file writer with size-based rotation.

```go
func NewRotatingFileWriter(filename string, maxSizeMB int, maxBackups int) (*RotatingFileWriter, error)
```

**Parameters:**
- `filename` - Path to the log file (e.g., `/var/log/app.log`)
- `maxSizeMB` - Maximum size in megabytes before rotation
- `maxBackups` - Maximum number of backup files to keep

**Example:**

```go
writer, err := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
if err != nil {
    panic(err)
}
defer writer.Close()

logger, _ := go_logs.New(
    go_logs.WithOutput(writer),
)
```

### WithRotatingFile Option

Convenience option for creating a logger with file rotation:

```go
func WithRotatingFile(filename string, maxSizeMB int, maxBackups int) Option
```

**Example:**

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
defer logger.Sync()
```

## Configuration

### Basic Configuration

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile(
        "/var/log/myapp/app.log", // Log file path
        100,                       // Max 100 MB per file
        5,                         // Keep 5 backup files
    ),
)
```

### With JSON Formatter

For production, combine with JSONFormatter:

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithLevel(go_logs.InfoLevel),
)
```

### With Multi-Output

Log to both file and console:

```go
fileWriter, _ := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)

logger, _ := go_logs.New(
    go_logs.WithMultiOutput(fileWriter, os.Stdout),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

## RotatingFileWriter Methods

### Write

Implements `io.Writer`. Writes bytes to the log file, rotating if necessary.

```go
func (w *RotatingFileWriter) Write(p []byte) (n int, err error)
```

### Rotate

Manually triggers rotation. Useful for signal handling (e.g., SIGHUP).

```go
func (w *RotatingFileWriter) Rotate() error
```

**Example:**

```go
// Handle SIGHUP for log rotation
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGHUP)

go func() {
    for range sigChan {
        if err := writer.Rotate(); err != nil {
            log.Printf("rotation failed: %v", err)
        }
    }
}()
```

### Sync

Flushes buffered data to disk.

```go
func (w *RotatingFileWriter) Sync() error
```

### Close

Flushes and closes the file. Safe to call multiple times.

```go
func (w *RotatingFileWriter) Close() error
```

### GetMaxSize

Returns the maximum file size in bytes.

```go
func (w *RotatingFileWriter) GetMaxSize() int64
```

### GetMaxBackups

Returns the maximum number of backups.

```go
func (w *RotatingFileWriter) GetMaxBackups() int
```

## Environment Variables

Configure rotation via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SAVE_LOG_FILE` | Enable file logging | `0` |
| `LOG_FILE_NAME` | Log file name | `log.txt` |
| `LOG_FILE_PATH` | Log file directory | current directory |
| `LOG_MAX_SIZE` | Max size in MB | `100` |
| `LOG_MAX_BACKUPS` | Max backup files | `5` |

**Example:**

```bash
export SAVE_LOG_FILE=1
export LOG_FILE_NAME=app.log
export LOG_FILE_PATH=/var/log/myapp
export LOG_MAX_SIZE=100
export LOG_MAX_BACKUPS=5
```

```go
// Logger picks up environment variables
logger, _ := go_logs.New()
```

## Enhanced Rotation (v3.2+)

For advanced rotation features (time-based, compression), use `RotatingFileConfig`:

```go
type RotatingFileConfig struct {
    Filename     string        // Log file path
    MaxSizeMB    int           // Max size before rotation
    MaxBackups   int           // Max backup files to keep
    MaxAge       int           // Max days to keep (0 = no limit)
    RotationType RotationType  // Size or time-based rotation
    Compress     bool          // Compress rotated files
}
```

### Rotation Types

```go
const (
    RotateSize   RotationType = iota  // Rotate by size (default)
    RotateDaily                        // Rotate daily at midnight
    RotateHourly                       // Rotate hourly
)
```

### Example: Daily Rotation with Compression

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFileEnhanced(go_logs.RotatingFileConfig{
        Filename:     "/var/log/app.log",
        MaxSizeMB:    100,
        MaxBackups:   30,
        MaxAge:       7,              // Keep 7 days
        RotationType: go_logs.RotateDaily,
        Compress:     true,           // Compress old files
    }),
)
```

## Complete Example

```go
package main

import (
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/drossan/go_logs"
)

func main() {
    // Create rotating file writer
    writer, err := go_logs.NewRotatingFileWriter(
        "/var/log/myapp/app.log",
        100,  // 100 MB max size
        5,    // Keep 5 backups
    )
    if err != nil {
        panic(err)
    }
    defer writer.Close()

    // Create logger with JSON formatter
    logger, err := go_logs.New(
        go_logs.WithOutput(writer),
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithLevel(go_logs.InfoLevel),
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    // Handle SIGHUP for manual rotation
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGHUP)

    go func() {
        for range sigChan {
            logger.Info("Received SIGHUP, rotating log file")
            if err := writer.Rotate(); err != nil {
                logger.Error("Rotation failed", go_logs.Err(err))
            }
        }
    }()

    // Application logging
    logger.Info("Application started",
        go_logs.String("version", "1.0.0"),
        go_logs.Int("pid", os.Getpid()),
    )

    // Simulate application work
    for i := 0; i < 1000; i++ {
        logger.Info("Processing request",
            go_logs.Int("request_id", i),
            go_logs.String("path", "/api/users"),
        )
        time.Sleep(100 * time.Millisecond)
    }

    logger.Info("Application shutting down")
}
```

## Best Practices

### 1. Always Call Sync Before Exit

Ensure buffered data is written:

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("app.log", 100, 5),
)
defer logger.Sync() // Important!
```

### 2. Use Absolute Paths

Use absolute paths to avoid confusion:

```go
// Good
go_logs.WithRotatingFile("/var/log/myapp/app.log", 100, 5)

// Risky - depends on working directory
go_logs.WithRotatingFile("logs/app.log", 100, 5)
```

### 3. Set Appropriate File Permissions

The writer uses 0600 (owner read/write) for security. Ensure the directory exists:

```go
// Create log directory
os.MkdirAll("/var/log/myapp", 0755)

logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/myapp/app.log", 100, 5),
)
```

### 4. Handle Rotation Signals

Support manual rotation via SIGHUP:

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGHUP)

go func() {
    for range sigChan {
        writer.Rotate()
    }
}()
```

### 5. Use JSON for Production

Combine rotation with JSON formatter:

```go
logger, _ := go_logs.New(
    go_logs.WithRotatingFile("/var/log/app.log", 100, 5),
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### 6. Size Recommendations

| Log Volume | Max Size | Max Backups | Disk Usage |
|------------|----------|-------------|------------|
| Low (< 1 GB/day) | 100 MB | 5 | 600 MB |
| Medium (1-10 GB/day) | 100 MB | 20 | 2.1 GB |
| High (> 10 GB/day) | 500 MB | 30 | 15.5 GB |

## Troubleshooting

### Permission Denied

```
Error: open /var/log/app.log: permission denied
```

**Solution**: Ensure the application has write permissions to the directory:

```bash
sudo mkdir -p /var/log/myapp
sudo chown $USER:$USER /var/log/myapp
```

### Directory Does Not Exist

```
Error: open /var/log/myapp/app.log: no such file or directory
```

**Solution**: The writer creates the directory automatically, but ensure the parent path exists:

```go
os.MkdirAll("/var/log/myapp", 0755)
```

### Rotation Not Occurring

**Causes**:
- `maxSizeMB` is too large for your log volume
- Logs are not being written frequently enough

**Solution**: Reduce `maxSizeMB` for testing:

```go
writer, _ := go_logs.NewRotatingFileWriter("app.log", 1, 5) // 1 MB for testing
```

### Out of Disk Space

**Solution**: Reduce `maxBackups` or enable compression:

```go
go_logs.WithRotatingFileEnhanced(go_logs.RotatingFileConfig{
    Filename:     "/var/log/app.log",
    MaxSizeMB:    100,
    MaxBackups:   5,   // Reduce from 20
    Compress:     true, // Enable compression
})
```

## Comparison with External Tools

| Feature | RotatingFileWriter | logrotate |
|---------|-------------------|-----------|
| Dependencies | None | External tool |
| Configuration | Programmatic | Config file |
| Compression | Built-in (enhanced) | Yes |
| Time-based rotation | Built-in (enhanced) | Yes |
| External signals | SIGHUP | SIGHUP |
| Process management | Internal | Cron/systemd |

## See Also

- [Configuration](Configuration.md) - Environment variables
- [Formatters](Formatters.md) - JSON formatter for production
- [Examples](Examples.md) - More file rotation examples
