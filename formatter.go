package go_logs

// Formatter defines how to convert a log Entry to bytes for output.
//
// This interface enables different output formats:
//   - TextFormatter: Human-readable with colors for development
//   - JSONFormatter: Structured JSON for log aggregators (ELK, Loki, Datadog)
//
// Implementations must be thread-safe and efficient.
type Formatter interface {
	// Format converts a log Entry to bytes for writing to output.
	// Returns an error if formatting fails (e.g., JSON encoding error).
	//
	// The returned bytes should include a trailing newline if appropriate
	// for the format (text formatters typically include it, JSON formatters
	// typically do not).
	Format(entry *Entry) ([]byte, error)
}

// FormatterConfig contains configuration options for formatters.
// This allows customization of timestamp format, colors, etc.
type FormatterConfig struct {
	// EnableColors enables ANSI color codes in output (TextFormatter only)
	EnableColors bool

	// EnableTimestamp includes timestamp in output
	EnableTimestamp bool

	// EnableLevel includes log level in output
	EnableLevel bool

	// TimestampFormat is the format string for timestamps (e.g., "2006/01/02 15:04:05")
	// See https://pkg.go.dev/time#pkg-constants for reference
	TimestampFormat string
}
