package go_logs

import "log"

func saveLog(message string, logType string) {
	if !isInit {
		Init()
	}

	// Issue #1 Fix: Use thread-safe getter instead of direct map access
	if getNotificationSettings(logType) {
		registerMessage(message)
	}
}

func registerMessage(message string) {
	if saveLogFile {
		file := openLogFile()
		logger := log.New(file, "", log.LstdFlags)

		// Issue #5 Fix: log.New never returns nil, remove useless check
		logger.Println(message)

		closeLogFile(file)
	}

	if notificationsEnabled {
		_ = notifier.SendNotification(message)
	}
}
