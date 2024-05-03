package go_logs

var notificationSettings = map[string]bool{
	"FATAL":   notificationLogFatal,
	"ERROR":   notificationLogError,
	"WARNING": notificationLogWarning,
	"INFO":    notificationLogInfo,
	"SUCCESS": notificationLogSuccess,
}

func saveLog(message string, logType string) {
	if !isInit {
		Init()
	}

	if notificationSettings[logType] && saveLogFile {
		registerMessage(message)
	}
}

func registerMessage(message string) {
	logger.Println(message)
	if notificationsEnabled {
		_ = notifier.SendNotification(message)
	}
}
