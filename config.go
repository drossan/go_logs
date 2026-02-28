package go_logs

import (
	"bufio"
	"github.com/drossan/go_logs/adapters"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var isInit bool

var saveLogFile bool
var logFileName string
var logFilePath string

var notificationsEnabled bool

var notificationLogFatal bool
var notificationLogError bool
var notificationLogWarning bool
var notificationLogInfo bool
var notificationLogSuccess bool

// Issue #1 Fix: Protect notificationSettings with mutex for concurrent access
var (
	notificationSettings      map[string]bool
	notificationSettingsMutex sync.RWMutex
)

var notifier *adapters.SlackNotifier

// Issue #9 Fix: Persistent file with buffering for performance
var (
	logFile     *os.File
	logWriter   *bufio.Writer
	logFileOnce sync.Once
	logFileMu   sync.Mutex
)

// Numeric log levels (syslog-style)
// These constants define the numeric values for each log level,
// allowing threshold-based logging similar to standard loggers.
const (
	LevelTrace  = 10  // Trace: Extremely detailed, high-volume information
	LevelDebug  = 20  // Debug: Detailed diagnostic information for troubleshooting
	LevelInfo   = 30  // Info: General operational messages
	LevelWarn   = 40  // Warning: Potential issues or non-critical situations
	LevelError  = 50  // Error: Operational errors that need attention
	LevelFatal  = 60  // Fatal: Application crashes or critical errors
	LevelSilent = 0   // Silent: Disables all logging
)

// logLevel stores the configured logging threshold
// Messages with level >= logLevel will be logged
var logLevel int // Default: 0 (disabled if not set)

// useLegacySystem tracks whether the old notification system is configured
// If true, legacy system takes precedence over LOG_LEVEL for backward compatibility
var useLegacySystem bool

// Init initializes the go_logs package with configuration from environment variables.
//
// This function must be called before using any logging functions. It reads configuration
// from environment variables and sets up file logging and Slack notifications if enabled.
//
// Environment Variables:
//   - SAVE_LOG_FILE: Enable file logging (0 or 1, default: 0)
//   - LOG_FILE_NAME: Name of the log file (default: "log.txt")
//   - LOG_FILE_PATH: Directory path for log files (default: current directory)
//   - NOTIFICATIONS_SLACK_ENABLED: Enable Slack notifications (0 or 1, default: 0)
//   - NOTIFICATION_FATAL_LOG: Send fatal logs to Slack (0 or 1)
//   - NOTIFICATION_ERROR_LOG: Send error logs to Slack (0 or 1)
//   - NOTIFICATION_WARNING_LOG: Send warning logs to Slack (0 or 1)
//   - NOTIFICATION_INFO_LOG: Send info logs to Slack (0 or 1)
//   - NOTIFICATION_SUCCESS_LOG: Send success logs to Slack (0 or 1)
//   - SLACK_TOKEN: Slack bot token for notifications
//   - SLACK_CHANNEL_ID: Slack channel ID for notifications
//
// Example:
//   // Set environment variables before calling Init()
//   os.Setenv("SAVE_LOG_FILE", "1")
//   os.Setenv("LOG_FILE_NAME", "app.log")
//   go_logs.Init()
//   go_logs.InfoLog("Application started")
//
// Note: Init() can be called multiple times safely, but subsequent calls may not
// reinitialize components that are already set up (like the persistent log file).
func Init() {
	var err error // Issue #3 Fix: Local error variable instead of global

	isInit = true

	saveLogFile, err = strconv.ParseBool(os.Getenv("SAVE_LOG_FILE"))
	if err != nil {
		log.Fatalf("Error parsing SAVE_LOG_FILE: %v", err)
	}

	if saveLogFile {
		logFileName = os.Getenv("LOG_FILE_NAME")

		if logFileName == "" {
			logFileName = "log.txt"
		}

		logFilePath = os.Getenv("LOG_FILE_PATH")
		openLogFile()
	}

	loadNotificationsConfig()

	notificationsEnabled, err = strconv.ParseBool(os.Getenv("NOTIFICATIONS_SLACK_ENABLED"))
	if err != nil {
		log.Fatalf("Error parsing NOTIFICATIONS_SLACK_ENABLED: %v", err)
	}

	if notificationsEnabled {
		loadSlackConfig()
	}

	loadLogLevel()
}

// getNumericLevel converts a log level string to its numeric value
// Supports both lowercase and uppercase level names
func getNumericLevel(level string) int {
	switch level {
	case "FATAL":
		return LevelFatal
	case "ERROR":
		return LevelError
	case "WARNING", "WARN":
		return LevelWarn
	case "INFO":
		return LevelInfo
	case "DEBUG":
		return LevelDebug
	case "TRACE":
		return LevelTrace
	default:
		return LevelSilent
	}
}

// loadLogLevel loads the LOG_LEVEL environment variable and sets the logging threshold
// Supports: trace, debug, info, warn, error, fatal, silent (case-insensitive)
func loadLogLevel() {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		return // Not configured, will use old system
	}

	levelStr = strings.ToLower(levelStr)
	switch levelStr {
	case "trace":
		logLevel = LevelTrace
	case "debug":
		logLevel = LevelDebug
	case "info":
		logLevel = LevelInfo
	case "warn", "warning":
		logLevel = LevelWarn
	case "error":
		logLevel = LevelError
	case "fatal":
		logLevel = LevelFatal
	case "silent", "none", "disable":
		logLevel = LevelSilent
	default:
		log.Printf("Warning: Unknown LOG_LEVEL '%s', using info (30)", levelStr)
		logLevel = LevelInfo
	}
}

// loadLogFormat loads the LOG_FORMAT environment variable and returns the appropriate formatter.
// Supports: text, json (case-insensitive, default: text)
//
// This is used by the v3 API to configure the default formatter based on environment.
func loadLogFormat() Formatter {
	format := os.Getenv("LOG_FORMAT")
	switch strings.ToLower(format) {
	case "json":
		return NewJSONFormatter()
	case "text", "":
		return NewTextFormatter()
	default:
		log.Printf("Warning: Unknown LOG_FORMAT '%s', using text format", format)
		return NewTextFormatter()
	}
}

func loadNotificationsConfig() {
	var err error // Issue #3 Fix: Local error variable instead of global

	notificationLogFatal, err = strconv.ParseBool(os.Getenv("NOTIFICATION_FATAL_LOG"))
	if err != nil {
		log.Fatalf("Error parsing NOTIFICATION_FATAL_LOG: %v", err)
	}

	notificationLogError, err = strconv.ParseBool(os.Getenv("NOTIFICATION_ERROR_LOG"))
	if err != nil {
		log.Fatalf("Error parsing NOTIFICATION_ERROR_LOG: %v", err)
	}

	notificationLogWarning, err = strconv.ParseBool(os.Getenv("NOTIFICATION_WARNING_LOG"))
	if err != nil {
		log.Fatalf("Error parsing NOTIFICATION_WARNING_LOG: %v", err)
	}

	notificationLogInfo, err = strconv.ParseBool(os.Getenv("NOTIFICATION_INFO_LOG"))
	if err != nil {
		log.Fatalf("Error parsing NOTIFICATION_INFO_LOG: %v", err)
	}

	notificationLogSuccess, err = strconv.ParseBool(os.Getenv("NOTIFICATION_SUCCESS_LOG"))
	if err != nil {
		log.Fatalf("Error parsing NOTIFICATION_SUCCESS_LOG: %v", err)
	}

	// Detect if legacy system is configured (any notification level explicitly enabled)
	// This ensures backward compatibility by taking precedence over LOG_LEVEL
	useLegacySystem = notificationLogFatal || notificationLogError ||
		notificationLogWarning || notificationLogInfo || notificationLogSuccess

	// Issue #1 Fix: Protect map write with mutex
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"FATAL":   notificationLogFatal,
		"ERROR":   notificationLogError,
		"WARNING": notificationLogWarning,
		"INFO":    notificationLogInfo,
		"SUCCESS": notificationLogSuccess,
	}
	notificationSettingsMutex.Unlock()
}

// getNotificationSettings returns whether notifications are enabled for a given log level
//
// This function implements a dual-system approach for backward compatibility:
// 1. Legacy system: If any NOTIFICATION_*_LOG variable is set, use the boolean map
// 2. New system: If LOG_LEVEL is set, use numeric threshold comparison
// 3. Legacy takes precedence when both are configured
//
// Thread-safe getter for notificationSettings with nil-check (Issue #2)
func getNotificationSettings(level string) bool {
	notificationSettingsMutex.RLock()
	defer notificationSettingsMutex.RUnlock()

	// Issue #2 Fix: Return false if map is not yet initialized
	if notificationSettings == nil {
		return false
	}

	// Legacy system takes precedence for backward compatibility
	if useLegacySystem {
		return notificationSettings[level]
	}

	// New system: Use numeric log level threshold
	// Log if message level >= configured level (syslog-style)
	if logLevel > 0 {
		messageLevel := getNumericLevel(level)
		return messageLevel >= logLevel
	}

	// No system configured, default to false
	return false
}

func loadSlackConfig() {
	// Issue #7 Fix: Handle error from NewSlackNotifier
	var err error
	notifier, err = adapters.NewSlackNotifier()
	if err != nil {
		// Log warning but don't fail - notifications will be disabled
		log.Printf("Warning: Slack notifications disabled: %v", err)
	}
}

// Issue #9 Fix: Initialize persistent log file with buffering (called once via sync.Once)
func initPersistentLogFile() {
	// Issue #4 Fix: Use filepath.Join() for portable path construction
	// Handle empty logFilePath gracefully
	var fullPath string
	if logFilePath == "" {
		fullPath = logFileName
	} else {
		fullPath = filepath.Join(logFilePath, logFileName)
	}

	// Issue #6 Fix: Use 0600 permissions (owner read/write only) instead of 0666 (world readable)
	// Logs may contain sensitive information, so they should not be world-readable
	var err error
	logFile, err = os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	// Create buffered writer for performance (Issue #9)
	logWriter = bufio.NewWriter(logFile)
}

// openLogFile initializes the persistent log file (called once via sync.Once)
// Issue #9: Changed from open/close per write to persistent file with buffering
func openLogFile() *os.File {
	logFileOnce.Do(initPersistentLogFile)
	return logFile
}

// getLogWriter returns the buffered writer for the log file
// Issue #9: New function to provide buffered writer for efficient writes
func getLogWriter() *bufio.Writer {
	logFileOnce.Do(initPersistentLogFile)
	return logWriter
}

// closeLogFile flushes and closes the persistent log file
// Issue #9: Changed to close the persistent file instead of per-operation file
// Made idempotent to handle multiple calls safely
func closeLogFile(file *os.File) {
	logFileMu.Lock()
	defer logFileMu.Unlock()

	// Idempotent: safe to call multiple times
	if logWriter != nil {
		err := logWriter.Flush()
		if err != nil {
			log.Printf("Error flushing log writer: %v", err)
		}
		logWriter = nil
	}

	if logFile != nil {
		err := logFile.Close()
		// Idempotent: don't fail on "already closed" errors
		if err != nil && !os.IsNotExist(err) {
			// Use log.Printf instead of log.Fatalf to avoid terminating tests
			// Errors closing are logged but not fatal
			log.Printf("Error closing log file (may already be closed): %v", err)
		}
		logFile = nil
	}

	// Reset sync.Once to allow re-initialization if needed
	logFileOnce = sync.Once{}
}

// Close flushes and closes the persistent log file cleanly.
//
// This function should be called when shutting down the application to ensure
// all buffered log messages are written to the log file. It is safe to call
// multiple times (idempotent).
//
// After calling Close(), the log file can be reopened by calling any logging function,
// which will automatically reinitialize the log file.
//
// Example:
//   defer go_logs.Close()
//   go_logs.InfoLog("Application shutting down")
//
// Note: If the log file was never opened (SAVE_LOG_FILE=0), this function does nothing.
func Close() {
	closeLogFile(nil)
}

// IsNotifierEnabled returns whether Slack notifications are enabled.
//
// This function provides a way to check if Slack notifications have been successfully
// initialized and are available. It returns false if:
//   - Slack notifications are disabled (NOTIFICATIONS_SLACK_ENABLED=0)
//   - Slack credentials are missing (SLACK_TOKEN or SLACK_CHANNEL_ID not set)
//   - The notifier failed to initialize
//
// Returns:
//   true if Slack notifications are enabled and available, false otherwise
//
// Example:
//   if go_logs.IsNotifierEnabled() {
//       go_logs.InfoLog("Slack notifications are active")
//   } else {
//       go_logs.WarningLog("Slack notifications are not configured")
//   }
func IsNotifierEnabled() bool {
	if notifier == nil {
		return false
	}
	return notifier.IsEnabled()
}
