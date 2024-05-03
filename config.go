package go_logs

import (
	"github.com/drossan/go_logs/adapters"
	"log"
	"os"
	"strconv"
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

var notificationSettings = map[string]bool{
	"FATAL":   notificationLogFatal,
	"ERROR":   notificationLogError,
	"WARNING": notificationLogWarning,
	"INFO":    notificationLogInfo,
	"SUCCESS": notificationLogSuccess,
}

var notifier *adapters.SlackNotifier

var err error

func Init() {
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

	notificationSettings = map[string]bool{
		"FATAL":   notificationLogFatal,
		"ERROR":   notificationLogError,
		"WARNING": notificationLogWarning,
		"INFO":    notificationLogInfo,
		"SUCCESS": notificationLogSuccess,
	}
}

func loadSlackConfig() {
	notifier = adapters.NewSlackNotifier()
}

func openLogFile() *os.File {
	file, err := os.OpenFile(logFilePath+"/"+logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	return file
}

func closeLogFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Fatalf("Error closing the log file: %v", err)
	}
}
