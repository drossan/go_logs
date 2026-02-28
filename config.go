package go_logs

import (
	"bufio"
	"github.com/drossan/go_logs/adapters"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
// Thread-safe getter for notificationSettings with nil-check (Issue #2)
func getNotificationSettings(level string) bool {
	notificationSettingsMutex.RLock()
	defer notificationSettingsMutex.RUnlock()

	// Issue #2 Fix: Return false if map is not yet initialized
	if notificationSettings == nil {
		return false
	}

	return notificationSettings[level]
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

// Close is the public API to close the log file cleanly
// Issue #9: New public function for explicit cleanup
func Close() {
	closeLogFile(nil)
}

// IsNotifierEnabled returns whether Slack notifications are enabled
// Issue #7: Public accessor for testing (enabled field is private)
func IsNotifierEnabled() bool {
	if notifier == nil {
		return false
	}
	return notifier.IsEnabled()
}
