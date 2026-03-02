package go_logs

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"log"
)

// Infof logs an informational message with formatting.
//
// This function provides fmt.Sprintf-like formatting for informational messages.
// The formatted message is displayed in yellow color to the terminal.
// If file logging is enabled, the message is saved to the log file.
// If Slack notifications are enabled for INFO level, the message is sent to Slack.
//
// Parameters:
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   Infof("Server started on port %d", 8080)
//   Infof("User %s logged in from %s", username, ipAddress)
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_INFO_LOG: Enable Slack notifications (0 or 1)
func Infof(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	InfoLog(message)
}

// Errorf logs an error message with formatting.
//
// This function provides fmt.Sprintf-like formatting for error messages.
// The formatted message is displayed in red color to the terminal.
// If file logging is enabled, the message is saved to the log file.
// If Slack notifications are enabled for ERROR level, the message is sent to Slack.
//
// Parameters:
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   Errorf("Failed to connect to %s: %v", host, err)
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_ERROR_LOG: Enable Slack notifications (0 or 1)
func Errorf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	ErrorLog(message)
}

// Warningf logs a warning message with formatting.
//
// This function provides fmt.Sprintf-like formatting for warning messages.
// The formatted message is displayed in yellow color to the terminal.
// If file logging is enabled, the message is saved to the log file.
// If Slack notifications are enabled for WARNING level, the message is sent to Slack.
//
// Parameters:
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   Warningf("API version %s is deprecated, use %s", oldVersion, newVersion)
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_WARNING_LOG: Enable Slack notifications (0 or 1)
func Warningf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	WarningLog(message)
}

// Successf logs a success message with formatting.
//
// This function provides fmt.Sprintf-like formatting for success messages.
// The formatted message is displayed in green color to the terminal.
// If file logging is enabled, the message is saved to the log file.
// If Slack notifications are enabled for SUCCESS level, the message is sent to Slack.
//
// Parameters:
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   Successf("Processed %d records in %dms", count, duration)
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_SUCCESS_LOG: Enable Slack notifications (0 or 1)
func Successf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	SuccessLog(message)
}

// Fatalf logs a fatal message with formatting and terminates the program.
//
// This function provides fmt.Sprintf-like formatting for fatal messages.
// The formatted message is displayed in red color to the terminal and includes a bomb emoji (💣).
// If file logging is enabled, the message is saved to the log file.
// If Slack notifications are enabled for FATAL level, the message is sent to Slack.
// After logging, the program terminates by calling log.Fatal().
//
// Parameters:
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   Fatalf("Critical error: %v", err)
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_FATAL_LOG: Enable Slack notifications (0 or 1)
func Fatalf(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	FatalLog(message)
}

// InfoLogCtx logs an informational message with context support.
//
// This function accepts a context.Context parameter for distributed tracing and cancellation.
// Currently, the context is accepted but not yet used internally.
// Future versions will extract trace IDs from context and include them in log messages.
//
// Parameters:
//   ctx     - The context (can be used for tracing, cancellation, etc.)
//   message - The informational message to log
//
// Example:
//   ctx := context.Background()
//   InfoLogCtx(ctx, "Request processed successfully")
func InfoLogCtx(ctx context.Context, message string) {
	// TODO: Extract trace ID from context and include in log
	InfoLog(message)
}

// ErrorLogCtx logs an error message with context support.
//
// This function accepts a context.Context parameter for distributed tracing and cancellation.
// Currently, the context is accepted but not yet used internally.
// Future versions will extract trace IDs from context and include them in log messages.
//
// Parameters:
//   ctx     - The context (can be used for tracing, cancellation, etc.)
//   message - The error message to log
//
// Example:
//   ctx := context.Background()
//   ErrorLogCtx(ctx, "Failed to process request")
func ErrorLogCtx(ctx context.Context, message string) {
	// TODO: Extract trace ID from context and include in log
	ErrorLog(message)
}

// WarningLogCtx logs a warning message with context support.
//
// This function accepts a context.Context parameter for distributed tracing and cancellation.
// Currently, the context is accepted but not yet used internally.
// Future versions will extract trace IDs from context and include them in log messages.
//
// Parameters:
//   ctx     - The context (can be used for tracing, cancellation, etc.)
//   message - The warning message to log
//
// Example:
//   ctx := context.Background()
//   WarningLogCtx(ctx, "Deprecated API usage detected")
func WarningLogCtx(ctx context.Context, message string) {
	// TODO: Extract trace ID from context and include in log
	WarningLog(message)
}

// SuccessLogCtx logs a success message with context support.
//
// This function accepts a context.Context parameter for distributed tracing and cancellation.
// Currently, the context is accepted but not yet used internally.
// Future versions will extract trace IDs from context and include them in log messages.
//
// Parameters:
//   ctx     - The context (can be used for tracing, cancellation, etc.)
//   message - The success message to log
//
// Example:
//   ctx := context.Background()
//   SuccessLogCtx(ctx, "Operation completed successfully")
func SuccessLogCtx(ctx context.Context, message string) {
	// TODO: Extract trace ID from context and include in log
	SuccessLog(message)
}

// InfoLogCtxf logs an informational message with context and formatting.
//
// This function combines context support with fmt.Sprintf-like formatting.
// The context parameter is accepted for distributed tracing support.
//
// Parameters:
//   ctx    - The context (can be used for tracing, cancellation, etc.)
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   ctx := context.Background()
//   InfoLogCtxf(ctx, "User %s logged in from %s", username, ipAddress)
func InfoLogCtxf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	InfoLogCtx(ctx, message)
}

// ErrorLogCtxf logs an error message with context and formatting.
//
// This function combines context support with fmt.Sprintf-like formatting.
// The context parameter is accepted for distributed tracing support.
//
// Parameters:
//   ctx    - The context (can be used for tracing, cancellation, etc.)
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   ctx := context.Background()
//   ErrorLogCtxf(ctx, "Request failed: %v", err)
func ErrorLogCtxf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	ErrorLogCtx(ctx, message)
}

// WarningLogCtxf logs a warning message with context and formatting.
//
// This function combines context support with fmt.Sprintf-like formatting.
// The context parameter is accepted for distributed tracing support.
//
// Parameters:
//   ctx    - The context (can be used for tracing, cancellation, etc.)
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   ctx := context.Background()
//   WarningLogCtxf(ctx, "API version %s is deprecated", oldVersion)
func WarningLogCtxf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	WarningLogCtx(ctx, message)
}

// SuccessLogCtxf logs a success message with context and formatting.
//
// This function combines context support with fmt.Sprintf-like formatting.
// The context parameter is accepted for distributed tracing support.
//
// Parameters:
//   ctx    - The context (can be used for tracing, cancellation, etc.)
//   format - The format string (follows fmt.Sprintf syntax)
//   args   - Arguments for the format string
//
// Example:
//   ctx := context.Background()
//   SuccessLogCtxf(ctx, "Processed %d records", count)
func SuccessLogCtxf(ctx context.Context, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	SuccessLogCtx(ctx, message)
}

// WarningLog logs a warning message.
//
// The message is displayed in yellow color to the terminal.
// If file logging is enabled (SAVE_LOG_FILE=1), the message is saved to the log file.
// If Slack notifications are enabled for WARNING level, the message is sent to Slack.
//
// Note: This function was previously missing from the API, even though the
// notificationLogWarning configuration variable was loaded. It has been added
// to provide complete logging level support.
//
// Parameters:
//   message - The warning message to log
//
// Example:
//   WarningLog("Database connection pool nearly exhausted")
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_WARNING_LOG: Enable Slack notifications (0 or 1)
func WarningLog(message string) {
	color.Set(color.FgYellow)
	log.Println(message)
	color.Unset()
	saveLog(message, "WARNING")
}
