package go_logs

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/drossan/go_logs/v3/domain"
)

var isInit bool

// warnOutput receives the configuration warnings emitted by Init() (invalid
// environment values, log file that cannot be opened). It is a variable so tests
// can capture it.
var warnOutput io.Writer = os.Stderr

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

// notifier is the notifier registered with SetNotifier (nil = none), guarded by
// notifierMu. missingNotifierWarnOnce makes the "no notifier" warning fire at
// most once per process; SetNotifier never re-arms it.
var (
	notifier                domain.Notifier
	notifierMu              sync.RWMutex
	missingNotifierWarnOnce sync.Once
)

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
	LevelTrace  = 10 // Trace: Extremely detailed, high-volume information
	LevelDebug  = 20 // Debug: Detailed diagnostic information for troubleshooting
	LevelInfo   = 30 // Info: General operational messages
	LevelWarn   = 40 // Warning: Potential issues or non-critical situations
	LevelError  = 50 // Error: Operational errors that need attention
	LevelFatal  = 60 // Fatal: Application crashes or critical errors
	LevelSilent = 0  // Silent: Disables all logging
)

// logLevel stores the configured logging threshold
// Messages with level >= logLevel will be logged
var logLevel int // Set by loadLogLevel(); LevelInfo when LOG_LEVEL is empty or invalid

// useLegacySystem tracks whether the old notification system is configured
// If true, legacy system takes precedence over LOG_LEVEL for backward compatibility
var useLegacySystem bool

// Init initializes the go_logs package with configuration from environment variables.
//
// This function must be called before using any logging functions. It reads configuration
// from environment variables and sets up file logging and notification settings.
//
// Environment Variables:
//   - LOG_LEVEL: Logging threshold (trace, debug, info, warn, error, fatal, silent; default: info)
//   - SAVE_LOG_FILE: Enable file logging (0 or 1, default: 0)
//   - LOG_FILE_NAME: Name of the log file (default: "log.txt")
//   - LOG_FILE_PATH: Directory path for log files (default: current directory)
//   - NOTIFICATIONS_SLACK_ENABLED: Send notifications through the notifier registered
//     with SetNotifier (0 or 1, default: 0). Init does not build a notifier: register
//     one, e.g. from github.com/drossan/go_logs/slack/v3
//   - NOTIFICATION_FATAL_LOG: Notify fatal logs (0 or 1)
//   - NOTIFICATION_ERROR_LOG: Notify error logs (0 or 1)
//   - NOTIFICATION_WARNING_LOG: Notify warning logs (0 or 1)
//   - NOTIFICATION_INFO_LOG: Notify info logs (0 or 1)
//   - NOTIFICATION_SUCCESS_LOG: Notify success logs (0 or 1)
//
// Example:
//
//	// Set environment variables before calling Init()
//	os.Setenv("SAVE_LOG_FILE", "1")
//	os.Setenv("LOG_FILE_NAME", "app.log")
//	go_logs.Init()
//	go_logs.InfoLog("Application started")
//
// Init never terminates the process. An empty or unset variable silently takes
// its default; a value that cannot be parsed takes its default and writes a
// warning to stderr naming the variable and the value received. If the log file
// cannot be opened, a warning is written to stderr and file logging is disabled.
// With an empty environment the effective level is Info, so InfoLog, SuccessLog,
// WarningLog, ErrorLog and FatalLog pass the level filter.
//
// Note: Init() can be called multiple times safely, but subsequent calls may not
// reinitialize components that are already set up (like the persistent log file).
func Init() {
	isInit = true

	saveLogFile = envBool("SAVE_LOG_FILE", false)

	if saveLogFile {
		logFileName = os.Getenv("LOG_FILE_NAME")

		if logFileName == "" {
			logFileName = "log.txt"
		}

		logFilePath = os.Getenv("LOG_FILE_PATH")
		openLogFile()
	}

	loadNotificationsConfig()

	notificationsEnabled = envBool("NOTIFICATIONS_SLACK_ENABLED", false)

	loadLogLevel()
}

// envBool lee una variable de entorno booleana. Vacía o ausente devuelve def en
// silencio; un valor no reconocido por strconv.ParseBool devuelve def y escribe un
// aviso en stderr con el nombre de la variable y el valor recibido.
func envBool(key string, def bool) bool {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		fmt.Fprintf(warnOutput, "go_logs: variable de entorno %s=%q no válida, se usa el valor por defecto (%t)\n", key, raw, def)
		return def
	}
	return v
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
	case "INFO", "SUCCESS":
		return LevelInfo
	case "DEBUG":
		return LevelDebug
	case "TRACE":
		return LevelTrace
	default:
		return LevelSilent
	}
}

// loadLogLevel loads the LOG_LEVEL environment variable and sets the logging threshold.
// Supports: trace, debug, info, warn, error, fatal, silent (case-insensitive).
// An empty value selects LevelInfo silently; an unknown value selects LevelInfo
// and writes a warning to warnOutput.
func loadLogLevel() {
	raw := os.Getenv("LOG_LEVEL")
	if raw == "" {
		logLevel = LevelInfo
		return
	}

	switch strings.ToLower(raw) {
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
		fmt.Fprintf(warnOutput, "go_logs: variable de entorno LOG_LEVEL=%q no válida, se usa el valor por defecto (info)\n", raw)
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
	notificationLogFatal = envBool("NOTIFICATION_FATAL_LOG", false)
	notificationLogError = envBool("NOTIFICATION_ERROR_LOG", false)
	notificationLogWarning = envBool("NOTIFICATION_WARNING_LOG", false)
	notificationLogInfo = envBool("NOTIFICATION_INFO_LOG", false)
	notificationLogSuccess = envBool("NOTIFICATION_SUCCESS_LOG", false)

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

// SetNotifier registra el notificador usado por la API v2 (ErrorLog, FatalLog…)
// cuando NOTIFICATIONS_SLACK_ENABLED está activo. nil lo desactiva. Los
// notificadores concretos viven fuera del core (p. ej. go_logs/slack/v3).
//
// It is safe to call concurrently with logging. If notifications are enabled and
// no notifier is registered when a message must be sent, a warning is written to
// stderr once per process and logging continues; errors returned by the notifier
// are ignored so they never interrupt logging.
//
// Example:
//
//	n, err := slack.NewNotifierFromEnv() // github.com/drossan/go_logs/slack/v3
//	if err == nil {
//	    go_logs.SetNotifier(n)
//	}
func SetNotifier(n domain.Notifier) {
	notifierMu.Lock()
	notifier = n
	notifierMu.Unlock()
}

// currentNotifier returns the notifier registered with SetNotifier, or nil.
func currentNotifier() domain.Notifier {
	notifierMu.RLock()
	defer notifierMu.RUnlock()
	return notifier
}

// sendNotification forwards message to the registered notifier. Without one it
// warns once per process through warnOutput and returns.
func sendNotification(message string) {
	n := currentNotifier()
	if n == nil {
		missingNotifierWarnOnce.Do(func() {
			fmt.Fprintln(warnOutput, "go_logs: NOTIFICATIONS_SLACK_ENABLED activo pero no hay notificador; llama a go_logs.SetNotifier (ver go_logs/slack/v3)")
		})
		return
	}
	_ = n.SendNotification(message)
}

// initPersistentLogFile opens the persistent log file with buffering (called once
// via sync.Once, Issue #9). If the file cannot be opened it writes a warning to
// warnOutput and disables file logging (saveLogFile = false, logWriter = nil)
// instead of terminating the process.
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
		fmt.Fprintf(warnOutput, "go_logs: no se puede abrir el fichero de log %s: %v; guardado en fichero desactivado\n", fullPath, err)
		saveLogFile = false
		logFile = nil
		logWriter = nil
		return
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
			// Use log.Printf: closing errors must never terminate the process
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
//
//	defer go_logs.Close()
//	go_logs.InfoLog("Application shutting down")
//
// Note: If the log file was never opened (SAVE_LOG_FILE=0), this function does nothing.
func Close() {
	closeLogFile(nil)
}

// IsNotifierEnabled reports whether v2 notifications will be sent: it returns
// true only if a notifier is registered with SetNotifier and
// NOTIFICATIONS_SLACK_ENABLED was active when Init ran.
//
// Example:
//
//	if go_logs.IsNotifierEnabled() {
//	    go_logs.InfoLog("Slack notifications are active")
//	} else {
//	    go_logs.WarningLog("Slack notifications are not configured")
//	}
func IsNotifierEnabled() bool {
	return currentNotifier() != nil && notificationsEnabled
}
