package go_logs

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"log"
)

// Issue #13: Formatted message functions (fmt.Sprintf-like)
// These provide more flexibility than the basic string-only functions

// Infof logs an informational message with formatting
// Issue #13: Adds formatted logging support for INFO level
func Infof(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	InfoLog(message)
}

// Errorf logs an error message with formatting
// Issue #13: Adds formatted logging support for ERROR level
func Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	ErrorLog(message)
}

// Warningf logs a warning message with formatting
// Issue #13 + #14: Adds formatted logging support for WARNING level
func Warningf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	WarningLog(message)
}

// Successf logs a success message with formatting
// Issue #13: Adds formatted logging support for SUCCESS level
func Successf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	SuccessLog(message)
}

// Fatalf logs a fatal message with formatting and terminates the program
// Issue #13: Adds formatted logging support for FATAL level
func Fatalf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	FatalLog(message)
}

// Issue #11: Context support for distributed tracing and cancellation
// These functions accept context.Context for better observability

// InfoLogCtx logs an informational message with context support
// Issue #11: Adds context support for INFO level
// The context can be used for distributed tracing, though currently not used internally
func InfoLogCtx(ctx context.Context, message string) {
	// TODO: Extract trace ID from context and include in log
	InfoLog(message)
}

// ErrorLogCtx logs an error message with context support
// Issue #11: Adds context support for ERROR level
func ErrorLogCtx(ctx context.Context, message string) {
	// TODO: Extract trace ID from context and include in log
	ErrorLog(message)
}

// WarningLogCtx logs a warning message with context support
// Issue #11 + #14: Adds context support for WARNING level
func WarningLogCtx(ctx context.Context, message string) {
	// TODO: Extract trace ID from context and include in log
	WarningLog(message)
}

// SuccessLogCtx logs a success message with context support
// Issue #11: Adds context support for SUCCESS level
func SuccessLogCtx(ctx context.Context, message string) {
	// TODO: Extract trace ID from context and include in log
	SuccessLog(message)
}

// Issue #11 + #13: Combined context + formatting support

// InfoLogCtxf logs an informational message with context and formatting
// Issue #11 + #13: Combines context support with formatted logging
func InfoLogCtxf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	InfoLogCtx(ctx, message)
}

// ErrorLogCtxf logs an error message with context and formatting
// Issue #11 + #13: Combines context support with formatted logging
func ErrorLogCtxf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	ErrorLogCtx(ctx, message)
}

// WarningLogCtxf logs a warning message with context and formatting
// Issue #11 + #13 + #14: Combines context support with formatted logging
func WarningLogCtxf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	WarningLogCtx(ctx, message)
}

// SuccessLogCtxf logs a success message with context and formatting
// Issue #11 + #13: Combines context support with formatted logging
func SuccessLogCtxf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	SuccessLogCtx(ctx, message)
}

// Issue #14: Implement WarningLog function
// WarningLog logs a warning message with yellow color
// Previously, notificationLogWarning was loaded but never used
func WarningLog(message string) {
	color.Set(color.FgYellow)
	log.Println(message)
	color.Unset()
	saveLog(message, "WARNING")
}
