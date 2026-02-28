package go_logs

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// TextFormatter formats log entries as human-readable text with optional colors.
// This is the default formatter for development environments.
//
// Output format: [TIMESTAMP] LEVEL message key=value key2=value2
//
// Example:
//
//	[2026/02/28 17:30:00] INFO connection established host=db.example.com port=5432
type TextFormatter struct {
	config      FormatterConfig
	levelColors map[Level]*color.Color
}

// NewTextFormatter creates a new TextFormatter with default configuration.
//
// Default configuration:
//   - EnableColors: true
//   - EnableTimestamp: true
//   - EnableLevel: true
//   - TimestampFormat: "2006/01/02 15:04:05"
//
// Color mapping by level:
//   - Trace: Cyan
//   - Debug: HiBlue
//   - Info: Yellow
//   - Warn: HiYellow
//   - Error: Red
//   - Fatal: HiRed
//   - Success: Green
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		config: FormatterConfig{
			EnableColors:    true,
			EnableTimestamp: true,
			EnableLevel:     true,
			TimestampFormat: "2006/01/02 15:04:05",
		},
		levelColors: map[Level]*color.Color{
			TraceLevel:   color.New(color.FgCyan),
			DebugLevel:   color.New(color.FgHiBlue),
			InfoLevel:    color.New(color.FgYellow),
			WarnLevel:    color.New(color.FgHiYellow),
			ErrorLevel:   color.New(color.FgRed),
			FatalLevel:   color.New(color.FgHiRed),
			SuccessLevel: color.New(color.FgGreen),
		},
	}
}

// NewTextFormatterWithConfig creates a new TextFormatter with custom configuration.
// This allows disabling colors, timestamps, or custom timestamp formats.
func NewTextFormatterWithConfig(config FormatterConfig) *TextFormatter {
	tf := &TextFormatter{
		config:      config,
		levelColors: make(map[Level]*color.Color),
	}

	// Initialize color mapping
	tf.levelColors[TraceLevel] = color.New(color.FgCyan)
	tf.levelColors[DebugLevel] = color.New(color.FgHiBlue)
	tf.levelColors[InfoLevel] = color.New(color.FgYellow)
	tf.levelColors[WarnLevel] = color.New(color.FgHiYellow)
	tf.levelColors[ErrorLevel] = color.New(color.FgRed)
	tf.levelColors[FatalLevel] = color.New(color.FgHiRed)
	tf.levelColors[SuccessLevel] = color.New(color.FgGreen)

	return tf
}

// Format converts a log Entry to formatted text bytes.
// Returns bytes with a trailing newline.
func (f *TextFormatter) Format(entry *Entry) ([]byte, error) {
	var buf bytes.Buffer

	// Timestamp
	if f.config.EnableTimestamp {
		buf.WriteString("[")
		buf.WriteString(entry.Timestamp.Format(f.config.TimestampFormat))
		buf.WriteString("] ")
	}

	// Level with color
	if f.config.EnableLevel {
		if f.config.EnableColors {
			if levelColor, ok := f.levelColors[entry.Level]; ok {
				levelColor.Fprint(&buf, entry.Level.String())
			} else {
				buf.WriteString(entry.Level.String())
			}
		} else {
			buf.WriteString(entry.Level.String())
		}
		buf.WriteString(" ")
	}

	// Message
	buf.WriteString(entry.Message)

	// Fields
	if len(entry.Fields) > 0 {
		buf.WriteString(" ")
		f.formatFields(&buf, entry.Fields)
	}

	buf.WriteString("\n")
	return buf.Bytes(), nil
}

// formatFields formats structured fields as key=value pairs.
// Special handling for different field types.
func (f *TextFormatter) formatFields(buf *bytes.Buffer, fields []Field) {
	var parts []string
	for _, field := range fields {
		switch field.Type() {
		case StringType:
			// Quote strings if they contain spaces
			val := fmt.Sprintf("%v", field.Value())
			if strings.Contains(val, " ") {
				parts = append(parts, fmt.Sprintf("%s=\"%s\"", field.Key(), val))
			} else {
				parts = append(parts, fmt.Sprintf("%s=%s", field.Key(), val))
			}
		case ErrorType:
			// Format errors specially
			if err, ok := field.Value().(error); ok && err != nil {
				parts = append(parts, fmt.Sprintf("%s=%s", field.Key(), err.Error()))
			} else {
				parts = append(parts, fmt.Sprintf("%s=%v", field.Key(), field.Value()))
			}
		default:
			// Default formatting for other types
			parts = append(parts, fmt.Sprintf("%s=%v", field.Key(), field.Value()))
		}
	}
	buf.WriteString(strings.Join(parts, " "))
}

// SetEnableColors enables or disables colors in the output.
// This is useful for non-TTY output (files, pipes).
func (f *TextFormatter) SetEnableColors(enabled bool) {
	f.config.EnableColors = enabled
}

// SetEnableTimestamp enables or disables timestamp output.
func (f *TextFormatter) SetEnableTimestamp(enabled bool) {
	f.config.EnableTimestamp = enabled
}

// SetEnableLevel enables or disables level output.
func (f *TextFormatter) SetEnableLevel(enabled bool) {
	f.config.EnableLevel = enabled
}

// SetTimestampFormat sets the timestamp format string.
// See https://pkg.go.dev/time#pkg-constants for reference.
func (f *TextFormatter) SetTimestampFormat(format string) {
	f.config.TimestampFormat = format
}
