package go_logs

import (
	"github.com/fatih/color"
	"log"
)

// FatalLog logs a fatal message and terminates the program.
//
// The message is displayed in red color to the terminal and includes a bomb emoji (💣).
// If file logging is enabled (SAVE_LOG_FILE=1), the message is saved to the log file.
// If Slack notifications are enabled for FATAL level, the message is sent to Slack.
//
// After logging, the program terminates by calling log.Fatal().
//
// Parameters:
//   message - The fatal message to log
//
// Example:
//   FatalLog("Database connection failed")
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_FATAL_LOG: Enable Slack notifications for fatal logs (0 or 1)
func FatalLog(message string) {
	color.Set(color.FgRed)
	saveLog(" 💣 "+message, "FATAL")
	color.Unset()
	log.Fatal(" 💣  " + message)
}

// ErrorLog logs an error message.
//
// The message is displayed in red color to the terminal.
// If file logging is enabled (SAVE_LOG_FILE=1), the message is saved to the log file.
// If Slack notifications are enabled for ERROR level, the message is sent to Slack.
//
// Parameters:
//   message - The error message to log
//
// Example:
//   ErrorLog("Failed to connect to database")
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_ERROR_LOG: Enable Slack notifications for error logs (0 or 1)
func ErrorLog(message string) {
	color.Set(color.FgRed)
	log.Println(message)
	color.Unset()
	saveLog(message, "ERROR")
}

// InfoLog logs an informational message.
//
// The message is displayed in yellow color to the terminal.
// If file logging is enabled (SAVE_LOG_FILE=1), the message is saved to the log file.
// If Slack notifications are enabled for INFO level, the message is sent to Slack.
//
// Parameters:
//   message - The informational message to log
//
// Example:
//   InfoLog("Server started on port 8080")
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_INFO_LOG: Enable Slack notifications for info logs (0 or 1)
func InfoLog(message string) {
	color.Set(color.FgYellow)
	log.Println(message)
	color.Unset()
	saveLog(message, "INFO")
}

// SuccessLog logs a success message.
//
// The message is displayed in green color to the terminal.
// If file logging is enabled (SAVE_LOG_FILE=1), the message is saved to the log file.
// If Slack notifications are enabled for SUCCESS level, the message is sent to Slack.
//
// Parameters:
//   message - The success message to log
//
// Example:
//   SuccessLog("Operation completed successfully")
//
// Environment variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1)
//   - NOTIFICATION_SUCCESS_LOG: Enable Slack notifications for success logs (0 or 1)
func SuccessLog(message string) {
	color.Set(color.FgGreen)
	log.Println(message)
	color.Unset()
	saveLog(message, "SUCCESS")
}
