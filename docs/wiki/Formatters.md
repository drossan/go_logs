# Formatters

Formatters control how log entries are converted to bytes for output. go_logs provides two built-in formatters: TextFormatter for development and JSONFormatter for production.

## Formatter Interface

All formatters implement the `Formatter` interface:

```go
type Formatter interface {
    Format(entry *Entry) ([]byte, error)
}
```

The `Format` method receives an `Entry` and returns formatted bytes. Implementations must be thread-safe.

## TextFormatter

TextFormatter produces human-readable output with optional ANSI colors. This is the default formatter and is ideal for development.

### Output Format

```
[TIMESTAMP] LEVEL message key=value key2="value with spaces"
```

### Example Output

```
[2026/02/28 17:30:00] INFO Server started host=localhost port=8080
[2026/02/28 17:30:01] WARN High memory usage usage_percent=85.5
[2026/02/28 17:30:02] ERROR Database connection failed error=connection refused host=db.example.com
```

### Creating a TextFormatter

```go
// Default configuration
formatter := go_logs.NewTextFormatter()

// Custom configuration
formatter := go_logs.NewTextFormatterWithConfig(go_logs.FormatterConfig{
    EnableColors:    true,
    EnableTimestamp: true,
    EnableLevel:     true,
    TimestampFormat: "2006/01/02 15:04:05",
})
```

### Configuration Methods

```go
// Enable/disable ANSI colors
formatter.SetEnableColors(true)

// Enable/disable timestamp
formatter.SetEnableTimestamp(true)

// Enable/disable level
formatter.SetEnableLevel(true)

// Set timestamp format (Go reference time format)
formatter.SetTimestampFormat("2006-01-02 15:04:05.000")
```

### Default Configuration

| Option | Default |
|--------|---------|
| EnableColors | true |
| EnableTimestamp | true |
| EnableLevel | true |
| TimestampFormat | "2006/01/02 15:04:05" |

### Color Mapping

| Level | Color |
|-------|-------|
| Trace | Cyan |
| Debug | HiBlue (bright blue) |
| Info | Yellow |
| Warn | HiYellow (bright yellow) |
| Error | Red |
| Fatal | HiRed (bright red) |
| Success | Green |

### Usage Example

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    // Create TextFormatter with default settings
    formatter := go_logs.NewTextFormatter()

    logger, _ := go_logs.New(
        go_logs.WithFormatter(formatter),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Debug("Debug message")
    logger.Info("Info message")
    logger.Warn("Warning message")
    logger.Error("Error message")
}
```

### Disabling Colors

For non-TTY output (files, pipes), disable colors:

```go
formatter := go_logs.NewTextFormatter()
formatter.SetEnableColors(false)

logger, _ := go_logs.New(
    go_logs.WithFormatter(formatter),
)
```

Or check if output is a terminal:

```go
formatter := go_logs.NewTextFormatter()
formatter.SetEnableColors(isTerminal(os.Stdout.Fd()))

logger, _ := go_logs.New(
    go_logs.WithFormatter(formatter),
)
```

### Custom Timestamp Format

```go
formatter := go_logs.NewTextFormatter()
formatter.SetTimestampFormat("2006-01-02T15:04:05.000Z07:00")

// Output: [2026-02-28T17:30:00.000+01:00] INFO message
```

## JSONFormatter

JSONFormatter produces structured JSON output ideal for log aggregation systems like ELK, Loki, Datadog, and Splunk.

### Output Format

```json
{"timestamp":"2026-02-28T17:30:00Z","level":"INFO","message":"message","fields":{"key":"value"}}
```

### Output Structure

```go
type JSONLogEntry struct {
    Timestamp  string                 `json:"timestamp,omitempty"`
    Level      string                 `json:"level,omitempty"`
    Message    string                 `json:"message"`
    Fields     map[string]interface{} `json:"fields,omitempty"`
    Caller     string                 `json:"caller,omitempty"`
    CallerFunc string                 `json:"caller_func,omitempty"`
    StackTrace string                 `json:"stack_trace,omitempty"`
}
```

### Creating a JSONFormatter

```go
// Default configuration
formatter := go_logs.NewJSONFormatter()

// Custom configuration
formatter := go_logs.NewJSONFormatterWithConfig(go_logs.FormatterConfig{
    EnableTimestamp: true,
    EnableLevel:     true,
    TimestampFormat: time.RFC3339,
})
```

### Configuration Methods

```go
// Enable/disable timestamp
formatter.SetEnableTimestamp(true)

// Enable/disable level
formatter.SetEnableLevel(true)

// Set timestamp format (default: RFC3339)
formatter.SetTimestampFormat(time.RFC3339Nano)
```

### Default Configuration

| Option | Default |
|--------|---------|
| EnableTimestamp | true |
| EnableLevel | true |
| TimestampFormat | time.RFC3339 |

### Usage Example

```go
package main

import (
    "os"
    "github.com/drossan/go_logs"
)

func main() {
    logger, _ := go_logs.New(
        go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        go_logs.WithOutput(os.Stdout),
    )

    logger.Info("Server started",
        go_logs.String("host", "localhost"),
        go_logs.Int("port", 8080),
    )
}
```

**Output:**
```json
{"timestamp":"2026-02-28T17:30:00Z","level":"INFO","message":"Server started","fields":{"host":"localhost","port":8080}}
```

### With Caller Information

```go
logger, _ := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithCaller(true),
)

logger.Info("Processing request")
```

**Output:**
```json
{"timestamp":"2026-02-28T17:30:00Z","level":"INFO","message":"Processing request","caller":"main.go:42","caller_func":"main.processRequest"}
```

### With Stack Trace

```go
logger, _ := go_logs.New(
    go_logs.WithFormatter(go_logs.NewJSONFormatter()),
    go_logs.WithStackTrace(true),
)

logger.Error("Something went wrong", go_logs.Err(err))
```

**Output:**
```json
{"timestamp":"2026-02-28T17:30:00Z","level":"ERROR","message":"Something went wrong","fields":{"error":"connection refused"},"stack_trace":"goroutine 1 [running]:\nmain.main()\n\t/app/main.go:42 +0x123\n..."}
```

### Error Handling

Errors are formatted using their `Error()` method:

```go
err := errors.New("connection refused")
logger.Error("Failed", go_logs.Err(err))
```

**Output:**
```json
{"timestamp":"2026-02-28T17:30:00Z","level":"ERROR","message":"Failed","fields":{"error":"connection refused"}}
```

## Integration with Log Aggregators

### ELK Stack (Elasticsearch, Logstash, Kibana)

Configure Filebeat or Logstash to read JSON logs:

```yaml
# filebeat.yml
filebeat.inputs:
- type: log
  paths:
    - /var/log/app/*.log
  json.keys_under_root: true
  json.add_error_key: true
```

### Grafana Loki

Use Promtail or the Loki Docker driver:

```yaml
# promtail.yml
scrape_configs:
- job_name: app
  static_configs:
  - targets:
      - localhost
    labels:
      job: app
      __path__: /var/log/app/*.log
  pipeline_stages:
  - json:
      expressions:
        level: level
        message: message
```

### Datadog

Use the Datadog Agent with JSON parsing:

```yaml
# datadog.yaml
logs_enabled: true

# conf.d/app.d/conf.yaml
logs:
  - type: file
    path: /var/log/app/app.log
    service: myapp
    source: go
```

### Splunk

Configure the Universal Forwarder:

```ini
# inputs.conf
[monitor:///var/log/app]
disabled = false
index = main
sourcetype = _json
```

## Creating Custom Formatters

Implement the `Formatter` interface for custom output formats.

### Example: CSV Formatter

```go
package main

import (
    "encoding/csv"
    "bytes"
    "strings"
    "github.com/drossan/go_logs"
)

type CSVFormatter struct{}

func (f *CSVFormatter) Format(entry *go_logs.Entry) ([]byte, error) {
    var buf bytes.Buffer
    writer := csv.NewWriter(&buf)

    // Build record
    record := []string{
        entry.Timestamp.Format("2006-01-02 15:04:05"),
        entry.Level.String(),
        entry.Message,
    }

    // Add fields
    for _, field := range entry.Fields {
        record = append(record,
            field.Key() + "=" + formatValue(field.Value()),
        )
    }

    if err := writer.Write(record); err != nil {
        return nil, err
    }
    writer.Flush()

    return buf.Bytes(), nil
}

func formatValue(v interface{}) string {
    return strings.ReplaceAll(fmt.Sprintf("%v", v), "\"", "\"\"")
}

// Usage
func main() {
    logger, _ := go_logs.New(
        go_logs.WithFormatter(&CSVFormatter{}),
    )
    logger.Info("Hello", go_logs.String("name", "World"))
}
```

### Example: XML Formatter

```go
package main

import (
    "bytes"
    "encoding/xml"
    "github.com/drossan/go_logs"
)

type XMLEntry struct {
    XMLName   xml.Name `xml:"log"`
    Timestamp string   `xml:"timestamp"`
    Level     string   `xml:"level"`
    Message   string   `xml:"message"`
    Fields    []XMLField `xml:"field"`
}

type XMLField struct {
    Key   string `xml:"key,attr"`
    Value string `xml:",chardata"`
}

type XMLFormatter struct{}

func (f *XMLFormatter) Format(entry *go_logs.Entry) ([]byte, error) {
    xmlEntry := XMLEntry{
        Timestamp: entry.Timestamp.Format("2006-01-02T15:04:05Z"),
        Level:     entry.Level.String(),
        Message:   entry.Message,
    }

    for _, field := range entry.Fields {
        xmlEntry.Fields = append(xmlEntry.Fields, XMLField{
            Key:   field.Key(),
            Value: fmt.Sprintf("%v", field.Value()),
        })
    }

    data, err := xml.Marshal(xmlEntry)
    if err != nil {
        return nil, err
    }

    return append(data, '\n'), nil
}
```

### Example: Syslog Formatter

```go
package main

import (
    "bytes"
    "fmt"
    "github.com/drossan/go_logs"
)

type SyslogFormatter struct {
    AppName string
    Host    string
}

func (f *SyslogFormatter) Format(entry *go_logs.Entry) ([]byte, error) {
    var buf bytes.Buffer

    // RFC 5424 format
    // <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID STRUCTURED-DATA MSG
    priority := f.calculatePriority(entry.Level)
    timestamp := entry.Timestamp.Format("2006-01-02T15:04:05.000Z07:00")

    buf.WriteString(fmt.Sprintf("<%d>1 %s %s %s - - - ",
        priority,
        timestamp,
        f.Host,
        f.AppName,
    ))

    buf.WriteString(entry.Message)

    // Add fields as key=value pairs
    for _, field := range entry.Fields {
        buf.WriteString(fmt.Sprintf(" %s=%v", field.Key(), field.Value()))
    }

    buf.WriteString("\n")
    return buf.Bytes(), nil
}

func (f *SyslogFormatter) calculatePriority(level go_logs.Level) int {
    var facility int = 1 // user-level
    var severity int

    switch level {
    case go_logs.DebugLevel, go_logs.TraceLevel:
        severity = 7 // debug
    case go_logs.InfoLevel:
        severity = 6 // info
    case go_logs.WarnLevel:
        severity = 4 // warning
    case go_logs.ErrorLevel:
        severity = 3 // error
    case go_logs.FatalLevel:
        severity = 2 // critical
    default:
        severity = 6 // info
    }

    return facility*8 + severity
}
```

## Choosing a Formatter

| Use Case | Recommended Formatter |
|----------|----------------------|
| Development | TextFormatter with colors |
| Production | JSONFormatter |
| Log aggregation (ELK, Loki) | JSONFormatter |
| Direct file reading | TextFormatter (no colors) |
| Custom output | Custom Formatter |

## Performance Considerations

| Formatter | Typical Performance | Notes |
|-----------|--------------------|-------|
| TextFormatter | ~220 ns/op | Fast, minimal allocations |
| JSONFormatter | ~250 ns/op | Uses encoding/json, proper escaping |
| Custom | Varies | Profile for production use |

## Switching Formatters at Runtime

You can change formatters by creating a new logger:

```go
var logger go_logs.Logger

func init() {
    if os.Getenv("ENV") == "production" {
        logger, _ = go_logs.New(
            go_logs.WithFormatter(go_logs.NewJSONFormatter()),
        )
    } else {
        logger, _ = go_logs.New(
            go_logs.WithFormatter(go_logs.NewTextFormatter()),
        )
    }
}
```

## See Also

- [API Reference](API-Reference.md) - Formatter interface documentation
- [Configuration](Configuration.md) - LOG_FORMAT environment variable
- [File Rotation](File-Rotation.md) - Combining formatters with file output
