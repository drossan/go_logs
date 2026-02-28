package go_logs

import (
	"fmt"
	"log"
	"time"
)

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
		// Issue #9 Fix: Use persistent buffered writer with mutex for thread safety
		// Buffer flushes automatically when full (default 4KB)
		logFileMu.Lock()
		defer logFileMu.Unlock()

		writer := getLogWriter()
		if writer != nil {
			// Write formatted log message with timestamp
			timestamp := time.Now().Format("2006/01/02 15:04:05")
			logLine := fmt.Sprintf("%s %s\n", timestamp, message)

			_, err := writer.WriteString(logLine)
			if err != nil {
				log.Printf("Error writing to log file: %v", err)
			}

			// Note: We rely on bufio's automatic flushing when buffer is full
			// For immediate flush (e.g., before program exit), call Close()
		}
	}

	if notificationsEnabled {
		_ = notifier.SendNotification(message)
	}
}
