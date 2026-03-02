# Hooks

Hooks provide an extension point for processing log entries before they are written. Use hooks to send logs to external systems, collect metrics, trigger alerts, or implement custom processing logic.

## Hook Interface

All hooks implement the `Hook` interface:

```go
type Hook interface {
    Run(entry *Entry) error
}
```

The `Run` method is called for every log entry that passes the level filter. If the hook returns an error, it is logged but does not prevent other hooks or output from proceeding.

## Entry Type

Hooks receive an `Entry` with all log information:

```go
type Entry struct {
    Level      Level
    Message    string
    Fields     []Field
    Timestamp  time.Time
    Caller     *CallerInfo
    StackTrace []byte
}
```

### Entry Methods

```go
// Check for a field
entry.HasField("user_id")

// Get a field value
entry.GetFieldValue("user_id")

// Clone the entry for modification
cloned := entry.Clone()

// Add more fields
modified := entry.WithFields(go_logs.String("extra", "value"))

// Check level
if entry.Level >= go_logs.ErrorLevel {
    // This is an error or worse
}
```

## Creating Hooks

### Using HookFunc

The simplest way to create a hook is using `HookFunc` or `NewFuncHook`:

```go
hook := go_logs.NewFuncHook(func(entry *go_logs.Entry) error {
    fmt.Printf("Log: [%s] %s\n", entry.Level, entry.Message)
    return nil
})

logger, _ := go_logs.New(
    go_logs.WithHooks(hook),
)
```

### Implementing the Hook Interface

For stateful hooks, implement the interface directly:

```go
type MetricsHook struct {
    counter *prometheus.CounterVec
}

func NewMetricsHook() *MetricsHook {
    return &MetricsHook{
        counter: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "log_entries_total",
                Help: "Total number of log entries",
            },
            []string{"level"},
        ),
    }
}

func (h *MetricsHook) Run(entry *go_logs.Entry) error {
    h.counter.WithLabelValues(entry.Level.String()).Inc()
    return nil
}

// Usage
logger, _ := go_logs.New(
    go_logs.WithHooks(NewMetricsHook()),
)
```

## Built-in Hook Patterns

### Metrics Collection Hook

Track log counts by level:

```go
type MetricsHook struct {
    counts map[go_logs.Level]int64
    mu     sync.RWMutex
}

func NewMetricsHook() *MetricsHook {
    return &MetricsHook{
        counts: make(map[go_logs.Level]int64),
    }
}

func (h *MetricsHook) Run(entry *go_logs.Entry) error {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.counts[entry.Level]++
    return nil
}

func (h *MetricsHook) GetCount(level go_logs.Level) int64 {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.counts[level]
}

// Usage
metricsHook := NewMetricsHook()
logger, _ := go_logs.New(go_logs.WithHooks(metricsHook))

// Later, check metrics
fmt.Printf("Error count: %d\n", metricsHook.GetCount(go_logs.ErrorLevel))
```

### Error Alerting Hook

Send alerts for errors:

```go
type AlertHook struct {
    notifier Notifier
    levels   map[go_logs.Level]bool
}

func NewAlertHook(notifier Notifier) *AlertHook {
    return &AlertHook{
        notifier: notifier,
        levels: map[go_logs.Level]bool{
            go_logs.ErrorLevel: true,
            go_logs.FatalLevel: true,
        },
    }
}

func (h *AlertHook) Run(entry *go_logs.Entry) error {
    if !h.levels[entry.Level] {
        return nil // Skip non-error levels
    }

    // Build alert message
    msg := fmt.Sprintf("[%s] %s", entry.Level, entry.Message)

    // Add fields
    for _, field := range entry.Fields {
        msg += fmt.Sprintf("\n  %s: %v", field.Key(), field.Value())
    }

    // Send alert asynchronously
    go h.notifier.Send(msg)
    return nil
}
```

### Filtering Hook

Filter logs based on content:

```go
type FilterHook struct {
    blockedKeys map[string]bool
}

func NewFilterHook(blockedKeys ...string) *FilterHook {
    h := &FilterHook{
        blockedKeys: make(map[string]bool),
    }
    for _, key := range blockedKeys {
        h.blockedKeys[key] = true
    }
    return h
}

func (h *FilterHook) Run(entry *go_logs.Entry) error {
    // Remove blocked fields
    for i := len(entry.Fields) - 1; i >= 0; i-- {
        if h.blockedKeys[entry.Fields[i].Key()] {
            entry.Fields = append(entry.Fields[:i], entry.Fields[i+1:]...)
        }
    }
    return nil
}
```

### Enrichment Hook

Add default fields to all entries:

```go
type EnrichmentHook struct {
    hostname string
    version  string
    env      string
}

func NewEnrichmentHook() *EnrichmentHook {
    hostname, _ := os.Hostname()
    return &EnrichmentHook{
        hostname: hostname,
        version:  os.Getenv("APP_VERSION"),
        env:      os.Getenv("APP_ENV"),
    }
}

func (h *EnrichmentHook) Run(entry *go_logs.Entry) error {
    // Add fields to the entry
    entry.Fields = append(entry.Fields,
        go_logs.String("hostname", h.hostname),
        go_logs.String("version", h.version),
        go_logs.String("environment", h.env),
    )
    return nil
}
```

### Rate Limiting Hook

Prevent log flooding:

```go
type RateLimitHook struct {
    limiter *rate.Limiter
}

func NewRateLimitHook(rateLimit int) *RateLimitHook {
    return &RateLimitHook{
        limiter: rate.NewLimiter(rate.Limit(rateLimit), rateLimit),
    }
}

func (h *RateLimitHook) Run(entry *go_logs.Entry) error {
    if !h.limiter.Allow() {
        return fmt.Errorf("rate limit exceeded")
    }
    return nil
}
```

### Sampling Hook

Sample high-volume logs:

```go
type SamplingHook struct {
    sampleRate int // Log 1 in N entries
    counter    int64
    mu         sync.Mutex
}

func NewSamplingHook(sampleRate int) *SamplingHook {
    return &SamplingHook{sampleRate: sampleRate}
}

func (h *SamplingHook) Run(entry *go_logs.Entry) error {
    // Always log errors and above
    if entry.Level >= go_logs.ErrorLevel {
        return nil
    }

    h.mu.Lock()
    defer h.mu.Unlock()

    h.counter++
    if h.counter%int64(h.sampleRate) != 0 {
        return fmt.Errorf("sampled out")
    }
    return nil
}
```

## Slack Hook

Send error logs to Slack:

```go
package hooks

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/drossan/go_logs"
)

type SlackHook struct {
    webhookURL string
    channel    string
    username   string
    levels     map[go_logs.Level]bool
}

type SlackMessage struct {
    Channel     string         `json:"channel"`
    Username    string         `json:"username"`
    Text        string         `json:"text"`
    Attachments []Attachment   `json:"attachments,omitempty"`
}

type Attachment struct {
    Color  string `json:"color"`
    Title  string `json:"title"`
    Text   string `json:"text"`
    Fields []Field `json:"fields,omitempty"`
}

type Field struct {
    Title string `json:"title"`
    Value string `json:"value"`
    Short bool   `json:"short"`
}

func NewSlackHook(webhookURL, channel, username string) *SlackHook {
    return &SlackHook{
        webhookURL: webhookURL,
        channel:    channel,
        username:   username,
        levels: map[go_logs.Level]bool{
            go_logs.ErrorLevel: true,
            go_logs.FatalLevel: true,
        },
    }
}

func (h *SlackHook) Run(entry *go_logs.Entry) error {
    if !h.levels[entry.Level] {
        return nil
    }

    // Build Slack message
    msg := SlackMessage{
        Channel:  h.channel,
        Username: h.username,
        Attachments: []Attachment{
            {
                Color:  h.getColor(entry.Level),
                Title:  fmt.Sprintf("[%s] %s", entry.Level, entry.Message),
                Fields: h.buildFields(entry),
            },
        },
    }

    // Send to Slack
    return h.send(msg)
}

func (h *SlackHook) getColor(level go_logs.Level) string {
    switch level {
    case go_logs.ErrorLevel:
        return "danger"
    case go_logs.WarnLevel:
        return "warning"
    default:
        return "#439FE0"
    }
}

func (h *SlackHook) buildFields(entry *go_logs.Entry) []Field {
    fields := make([]Field, 0, len(entry.Fields))
    for _, f := range entry.Fields {
        fields = append(fields, Field{
            Title: f.Key(),
            Value: fmt.Sprintf("%v", f.Value()),
            Short: len(fmt.Sprintf("%v", f.Value())) < 20,
        })
    }
    return fields
}

func (h *SlackHook) send(msg SlackMessage) error {
    body, err := json.Marshal(msg)
    if err != nil {
        return err
    }

    resp, err := http.Post(h.webhookURL, "application/json", bytes.NewReader(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("slack returned status %d", resp.StatusCode)
    }

    return nil
}

// Usage
func main() {
    slackHook := NewSlackHook(
        "https://hooks.slack.com/services/XXX/YYY/ZZZ",
        "#alerts",
        "Logger",
    )

    logger, _ := go_logs.New(
        go_logs.WithHooks(slackHook),
    )

    logger.Error("Database connection failed", go_logs.Err(err))
}
```

## Multiple Hooks

Add multiple hooks to a logger:

```go
logger, _ := go_logs.New(
    go_logs.WithHooks(
        NewMetricsHook(),
        NewAlertHook(notifier),
        NewEnrichmentHook(),
    ),
)
```

Hooks are executed in the order they are added.

## Hook Error Handling

If a hook returns an error, it is logged but does not stop other hooks or output:

```go
func (h *MyHook) Run(entry *go_logs.Entry) error {
    if err := h.process(entry); err != nil {
        // Log the error but don't fail
        log.Printf("hook error: %v", err)
        return err
    }
    return nil
}
```

## Conditional Hooks

Hooks can conditionally process entries:

```go
type ConditionalHook struct {
    condition func(*go_logs.Entry) bool
    hook      go_logs.Hook
}

func NewConditionalHook(condition func(*go_logs.Entry) bool, hook go_logs.Hook) *ConditionalHook {
    return &ConditionalHook{condition: condition, hook: hook}
}

func (h *ConditionalHook) Run(entry *go_logs.Entry) error {
    if h.condition(entry) {
        return h.hook.Run(entry)
    }
    return nil
}

// Usage: Only send errors from production to Slack
slackHook := NewSlackHook(...)
conditionalHook := NewConditionalHook(
    func(e *go_logs.Entry) bool {
        return e.Level >= go_logs.ErrorLevel && os.Getenv("ENV") == "production"
    },
    slackHook,
)
```

## Async Hooks

For slow operations (network calls), run hooks asynchronously:

```go
type AsyncHook struct {
    hook   go_logs.Hook
    buffer chan *go_logs.Entry
}

func NewAsyncHook(hook go_logs.Hook, bufferSize int) *AsyncHook {
    h := &AsyncHook{
        hook:   hook,
        buffer: make(chan *go_logs.Entry, bufferSize),
    }
    go h.process()
    return h
}

func (h *AsyncHook) Run(entry *go_logs.Entry) error {
    // Clone entry to avoid race conditions
    select {
    case h.buffer <- entry.Clone():
    default:
        // Buffer full, drop entry
    }
    return nil
}

func (h *AsyncHook) process() {
    for entry := range h.buffer {
        h.hook.Run(entry)
    }
}

// Usage
logger, _ := go_logs.New(
    go_logs.WithHooks(NewAsyncHook(NewSlackHook(...), 100)),
)
```

## Best Practices

### 1. Keep Hooks Fast

Hooks run synchronously (unless you make them async). Avoid slow operations:

```go
// Bad: Blocking HTTP call
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

### 2. Handle Errors Gracefully

Don't let hook errors crash your application:

```go
func (h *Hook) Run(entry *go_logs.Entry) error {
    if err := h.process(entry); err != nil {
        // Log but don't propagate
        fmt.Fprintf(os.Stderr, "hook error: %v\n", err)
        return nil // or return err to log it
    }
    return nil
}
```

### 3. Clone Entries for Modification

If you modify entries or use them asynchronously, clone them:

```go
func (h *Hook) Run(entry *go_logs.Entry) error {
    // Clone for async use
    cloned := entry.Clone()

    go func() {
        // Safe to use cloned
        h.process(cloned)
    }()

    return nil
}
```

### 4. Filter Early

Check conditions before expensive operations:

```go
func (h *SlackHook) Run(entry *go_logs.Entry) error {
    // Skip non-errors immediately
    if entry.Level < go_logs.ErrorLevel {
        return nil
    }

    // Only now do expensive work
    return h.sendToSlack(entry)
}
```

### 5. Use Rate Limiting

Protect external services from log floods:

```go
type RateLimitedHook struct {
    hook   go_logs.Hook
    limiter *rate.Limiter
}

func (h *RateLimitedHook) Run(entry *go_logs.Entry) error {
    if !h.limiter.Allow() {
        return nil // Skip this entry
    }
    return h.hook.Run(entry)
}
```

## See Also

- [API Reference](API-Reference.md) - Hook interface documentation
- [Examples](Examples.md) - More hook examples
- [Configuration](Configuration.md) - Configuring hooks via options
