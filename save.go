package go_logs

import "log"

func saveLog(message string, logType string) {
	if !isInit {
		Init()
	}

	if notificationSettings[logType] && saveLogFile {
		registerMessage(message)
	}
}

func registerMessage(message string) {
	file := openLogFile()
	logger := log.New(file, "", log.LstdFlags)

	if logger != nil {
		logger.Println(message)
	}

	closeLogFile(file)

	if notificationsEnabled {
		_ = notifier.SendNotification(message)
	}
}
