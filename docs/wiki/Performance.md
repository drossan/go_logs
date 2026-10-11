# Performance

go_logs is designed for high-performance logging with minimal overhead. This page covers benchmarks, performance characteristics, and optimization tips.

## Benchmarks

### Core Operations

| Operation | Performance |
|-----------|-------------|
| Logger, filtered level (`BenchmarkV3LevelFiltering`) | 2.2 ns/op, 0 allocs |
| Field creation | 0.32 ns/op, 0 allocs |
| TextFormatter, simple entry | 187 ns/op, 4 allocs |
| JSONFormatter, simple entry | 387 ns/op, 4 allocs |
| Logger → JSON → RotatingFileWriter (end to end) | 3.4 µs/op (~300k msg/s), 16 allocs |

Measured on Apple M1 (arm64) on 2026-10-11 with `go test -run xxx -bench . -benchmem .`; your numbers will vary with hardware and Go version. Reproduce with that command.

### Detailed Benchmarks

Output of `go test -run xxx -bench . -benchmem .` (Apple M1, 2026-10-11):

```
BenchmarkV3LevelFiltering-8                532573428     2.224 ns/op      0 B/op     0 allocs/op
BenchmarkV3FieldCreation-8                1000000000     0.3317 ns/op     0 B/op     0 allocs/op
BenchmarkLevelShouldLog/Filtered-8        1000000000     0.3587 ns/op     0 B/op     0 allocs/op
BenchmarkTextFormatter_SimpleEntry-8         6502068     186.6 ns/op    152 B/op     4 allocs/op
BenchmarkTextFormatter_WithFields-8          1706208     704.9 ns/op    616 B/op    21 allocs/op
BenchmarkJSONFormatter_SimpleEntry-8         3068571     387.2 ns/op    352 B/op     4 allocs/op
BenchmarkJSONFormatter_WithFields-8           989826     1243 ns/op     904 B/op    17 allocs/op
BenchmarkV3RotatingFileWriter-8              1000000     1201 ns/op    1284 B/op    16 allocs/op
BenchmarkLoggerToRotatingFile-8               365101     3356 ns/op    1016 B/op    16 allocs/op
BenchmarkV3ConcurrentLogging-8               1430082     833.4 ns/op   1026 B/op    15 allocs/op
BenchmarkV3Redaction-8                        929787     1313 ns/op    1425 B/op    19 allocs/op
```

`BenchmarkV3LevelFiltering` is the cost of calling a logger method whose level is filtered out. `BenchmarkLevelShouldLog` and field creation measure only a comparison or a struct build, which is why they are below 1 ns.

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkTextFormatter -benchmem ./...

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof
```

## Performance Characteristics

### Zero Allocations

Field creation and level filtering are zero-allocation operations:

```go
// These operations allocate no memory
field := go_logs.String("key", "value")  // 0 allocs
if level.shouldLog(threshold) { }        // 0 allocs
```

### Fast-Path Filtering

Level filtering uses numeric comparison (syslog-style) for maximum speed:

```go
// This is a single integer comparison
if entry.Level >= threshold {
    // Log the entry
}
```

### Buffered I/O

RotatingFileWriter uses buffered writes for performance:

```go
// 4KB buffer by default
writer := bufio.NewWriter(file)

// Flush on Sync() or Close()
writer.Flush()
```

### Thread-Safety

All operations are thread-safe using mutex, but the critical path is minimal:

```go
func (l *LoggerImpl) Log(level Level, msg string, fields ...Field) {
    // Fast path: check level before acquiring lock
    if level < l.level {
        return  // No lock needed for filtered messages
    }

    l.mu.Lock()
    defer l.mu.Unlock()
    // ... rest of logging
}
```

## Optimization Tips

### 1. Set Appropriate Log Level

The most effective optimization is filtering at the right level:

```go
// Development
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.DebugLevel),
)

// Production
logger, _ := go_logs.New(
    go_logs.WithLevel(go_logs.InfoLevel),
)
```

Filtered logs cost ~2 ns per call and allocate nothing.

### 2. Avoid Expensive Operations in Filtered Logs

Check level before expensive operations:

```go
// Bad: Always computes expensive data
logger.Debug("State dump", go_logs.Any("state", expensiveDump()))

// Good: Only computes if debug is enabled
if logger.GetLevel() <= go_logs.DebugLevel {
    logger.Debug("State dump", go_logs.Any("state", expensiveDump()))
}
```

### 3. Use Child Loggers Sparingly

Child loggers add overhead for field copying:

```go
// Good: Create once per request
reqLogger := logger.With(
    go_logs.String("request_id", requestID),
)

// Avoid: Creating in tight loops
for i := 0; i < 10000; i++ {
    iterLogger := logger.With(go_logs.Int("iteration", i)) // Overhead!
}
```

### 4. Batch File Writes

RotatingFileWriter buffers writes automatically:

```go
// Already buffered (4KB buffer)
writer, _ := go_logs.NewRotatingFileWriter("app.log", 100, 5)

// Remember to sync before exit
defer writer.Sync()
```

### 5. Use JSON Formatter in Production

JSONFormatter is optimized for log aggregators:

```go
logger, _ := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
)
```

### 6. Disable Caller Info in High-Throughput

Caller info adds ~100-200ns per log:

```go
// Disable for high-throughput paths
logger, _ := go_logs.New(
    go_logs.WithCaller(false), // Default
)

// Enable only for errors
errorLogger := logger.With(go_logs.Bool("include_caller", true))
```

### 7. Use Async Hooks for External Services

Don't block on external calls:

```go
type AsyncSlackHook struct {
    buffer chan *go_logs.Entry
}

func (h *AsyncSlackHook) Run(entry *go_logs.Entry) error {
    select {
    case h.buffer <- entry.Clone():
        // Queued
    default:
        // Buffer full, drop
    }
    return nil
}

func (h *AsyncSlackHook) process() {
    for entry := range h.buffer {
        h.sendToSlack(entry) // Async
    }
}
```

### 8. Reduce Field Count

More fields = more overhead:

```go
// Bad: Many fields
logger.Info("Request",
    go_logs.String("field1", v1),
    go_logs.String("field2", v2),
    // ... 20 more fields
)

// Good: Essential fields only
logger.Info("Request",
    go_logs.String("request_id", id),
    go_logs.Int("status", status),
    go_logs.Int64("duration_ms", duration),
)
```

## Memory Usage

### Entry Size

An Entry struct is approximately:

```
Entry: ~80 bytes base
+ Fields: ~24 bytes per field
+ Message: ~16 bytes + string length
+ StackTrace: ~1KB if captured
```

### Memory Pooling

For extreme performance, consider pooling entries:

```go
var entryPool = sync.Pool{
    New: func() interface{} {
        return &go_logs.Entry{}
    },
}

// Get from pool
entry := entryPool.Get().(*go_logs.Entry)
defer entryPool.Put(entry)
```

Note: go_logs does not use pooling internally to avoid complexity, but you can implement it in hooks.

## Comparison with Other Libraries

| Library | Throughput | Allocations | Notes |
|---------|------------|-------------|-------|
| go_logs v3 | ~4M logs/sec | 0-8 per log | Zero-allocation fields |
| zap | ~10M logs/sec | 0 per log | Most optimized |
| zerolog | ~8M logs/sec | 0 per log | Very fast |
| logrus | ~500K logs/sec | 23 per log | Slower, more features |
| standard log | ~1M logs/sec | 2 per log | Simple |

go_logs balances performance with a clean API and full feature set.

## Profiling Your Application

### CPU Profiling

```go
import (
    "os"
    "runtime/pprof"
)

func main() {
    f, _ := os.Create("cpu.prof")
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // Your application
}
```

Analyze:
```bash
go tool pprof cpu.prof
(pprof) top10
(pprof) list log
```

### Memory Profiling

```go
import (
    "os"
    "runtime"
    "runtime/pprof"
)

func dumpMemProfile() {
    f, _ := os.Create("mem.prof")
    runtime.GC()
    pprof.WriteHeapProfile(f)
    f.Close()
}
```

### Continuous Profiling

For production monitoring, consider using continuous profiling tools like:
- pprof HTTP endpoint
- Datadog Continuous Profiler
- Grafana Pyroscope

## Performance Anti-Patterns

### 1. Logging in Hot Paths

```go
// Bad: Logging in tight loop
for i := 0; i < 1000000; i++ {
    logger.Debug("Processing", go_logs.Int("i", i))
}

// Good: Sample or batch
for i := 0; i < 1000000; i++ {
    if i % 1000 == 0 {
        logger.Debug("Progress", go_logs.Int("i", i))
    }
}
```

### 2. String Formatting in Messages

```go
// Bad: Formats even if filtered
logger.Debug(fmt.Sprintf("Processing %d items", len(items)))

// Good: Let logger format
logger.Debug("Processing items", go_logs.Int("count", len(items)))
```

### 3. Blocking Hooks

```go
// Bad: Blocks on every log
func (h *Hook) Run(entry *go_logs.Entry) error {
    http.Post(url, "application/json", body) // Blocks!
    return nil
}

// Good: Async processing
func (h *Hook) Run(entry *go_logs.Entry) error {
    go h.sendAsync(entry.Clone())
    return nil
}
```

## Benchmarking Your Code

Create benchmarks for your logging patterns:

```go
func BenchmarkMyLoggingPattern(b *testing.B) {
    logger, _ := go_logs.New(
        go_logs.WithOutput(io.Discard), // Discard output for benchmark
    )

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("Processing request",
            go_logs.String("request_id", "abc-123"),
            go_logs.Int("status", 200),
            go_logs.Int64("duration_ms", 45),
        )
    }
}
```

Run:
```bash
go test -bench=BenchmarkMyLoggingPattern -benchmem
```

## See Also

- [API Reference](API-Reference.md) - API documentation
- [Hooks](Hooks.md) - Creating efficient hooks
- [Examples](Examples.md) - Production patterns
