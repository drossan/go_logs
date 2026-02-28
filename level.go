package go_logs

import "strings"

// Level represents the log level (syslog-style: Trace=10, Debug=20, Info=30, etc.)
// This numeric representation enables fast-path threshold-based filtering.
type Level int

const (
	// TraceLevel represents extremely detailed, high-volume logging (Level 10)
	TraceLevel Level = 10

	// DebugLevel represents detailed diagnostic information for troubleshooting (Level 20)
	DebugLevel Level = 20

	// InfoLevel represents general informational messages (Level 30)
	InfoLevel Level = 30

	// WarnLevel represents warning messages for potentially harmful situations (Level 40)
	WarnLevel Level = 40

	// ErrorLevel represents error events that might still allow the application to continue (Level 50)
	ErrorLevel Level = 50

	// FatalLevel represents critical errors that will terminate the application (Level 60)
	FatalLevel Level = 60

	// SuccessLevel represents successful operations (between Debug and Info, Level 25)
	// This level is maintained for v2 backward compatibility
	SuccessLevel Level = 25

	// SilentLevel disables all logging (Level 0)
	SilentLevel Level = 0
)

// shouldLog implements fast-path filtering for log levels.
// Returns true if this level should be logged given the threshold.
// This is a critical performance optimization and must be < 5ns.
//
// Syslog-style filtering: log if message level >= threshold level.
// For example, if threshold is WarnLevel (40), then ErrorLevel (50) and FatalLevel (60)
// will be logged, but InfoLevel (30) and below will be filtered out.
func (l Level) shouldLog(threshold Level) bool {
	return l >= threshold
}

// ShouldLog returns true if this level should be logged given the threshold.
// This is the public version of shouldLog for use in hooks and other external code.
//
// Syslog-style filtering: log if message level >= threshold level.
//
// Parameters:
//   threshold - The minimum level threshold
//
// Returns:
//   true if this level meets or exceeds the threshold, false otherwise
//
// Example:
//
//	if entry.Level.ShouldLog(ErrorLevel) {
//	    // Send to monitoring system
//	}
func (l Level) ShouldLog(threshold Level) bool {
	return l.shouldLog(threshold)
}

// String returns the string representation of the log level.
// Returns "UNKNOWN" for undefined levels.
func (l Level) String() string {
	switch l {
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	case SuccessLevel:
		return "SUCCESS"
	case SilentLevel:
		return "SILENT"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel converts a string to a Level.
// Case-insensitive. Returns InfoLevel as default for unknown strings.
//
// Supported values: trace, debug, info, warn, warning, error, fatal, silent, none, disable
func ParseLevel(levelStr string) Level {
	switch strings.ToUpper(levelStr) {
	case "TRACE":
		return TraceLevel
	case "DEBUG":
		return DebugLevel
	case "INFO":
		return InfoLevel
	case "WARN", "WARNING":
		return WarnLevel
	case "ERROR":
		return ErrorLevel
	case "FATAL":
		return FatalLevel
	case "SUCCESS":
		return SuccessLevel
	case "SILENT", "NONE", "DISABLE":
		return SilentLevel
	default:
		return InfoLevel // Default level
	}
}
