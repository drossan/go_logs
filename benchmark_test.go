package go_logs

import (
	"io"
	"log"
	"os"
	"testing"
)

// BenchmarkFileWrite measures the performance of writing to log file
// Issue #9: Current implementation opens/closes file on each write (inefficient)
func BenchmarkFileWrite(t *testing.B) {
	// Setup: Redirect log output to suppress console spam during benchmark
	log.SetOutput(io.Discard)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tempDir := t.TempDir()
	os.Setenv("SAVE_LOG_FILE", "1")
	os.Setenv("LOG_FILE_NAME", "bench.log")
	os.Setenv("LOG_FILE_PATH", tempDir)
	os.Setenv("NOTIFICATION_FATAL_LOG", "0")
	os.Setenv("NOTIFICATION_ERROR_LOG", "0")
	os.Setenv("NOTIFICATION_WARNING_LOG", "0")
	os.Setenv("NOTIFICATION_INFO_LOG", "1")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "0")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")

	Init()

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		InfoLog("Benchmark test message")
	}
}

// BenchmarkConcurrentLogs measures concurrent logging performance
func BenchmarkConcurrentLogs(t *testing.B) {
	// Setup: Redirect log output to suppress console spam during benchmark
	log.SetOutput(io.Discard)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tempDir := t.TempDir()
	os.Setenv("SAVE_LOG_FILE", "1")
	os.Setenv("LOG_FILE_NAME", "bench_concurrent.log")
	os.Setenv("LOG_FILE_PATH", tempDir)
	os.Setenv("NOTIFICATION_FATAL_LOG", "0")
	os.Setenv("NOTIFICATION_ERROR_LOG", "0")
	os.Setenv("NOTIFICATION_WARNING_LOG", "0")
	os.Setenv("NOTIFICATION_INFO_LOG", "1")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "0")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")

	Init()

	t.ResetTimer()
	t.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			InfoLog("Concurrent log message")
		}
	})
}

// BenchmarkErrorLog measures ErrorLog performance
func BenchmarkErrorLog(t *testing.B) {
	// Setup: Redirect log output to suppress console spam during benchmark
	log.SetOutput(io.Discard)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tempDir := t.TempDir()
	os.Setenv("SAVE_LOG_FILE", "1")
	os.Setenv("LOG_FILE_NAME", "bench_error.log")
	os.Setenv("LOG_FILE_PATH", tempDir)
	os.Setenv("NOTIFICATION_FATAL_LOG", "0")
	os.Setenv("NOTIFICATION_ERROR_LOG", "0")
	os.Setenv("NOTIFICATION_WARNING_LOG", "0")
	os.Setenv("NOTIFICATION_INFO_LOG", "1")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "0")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")

	Init()

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		ErrorLog("Benchmark error message")
	}
}

// BenchmarkSuccessLog measures SuccessLog performance
func BenchmarkSuccessLog(t *testing.B) {
	// Setup: Redirect log output to suppress console spam during benchmark
	log.SetOutput(io.Discard)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tempDir := t.TempDir()
	os.Setenv("SAVE_LOG_FILE", "1")
	os.Setenv("LOG_FILE_NAME", "bench_success.log")
	os.Setenv("LOG_FILE_PATH", tempDir)
	os.Setenv("NOTIFICATION_FATAL_LOG", "0")
	os.Setenv("NOTIFICATION_ERROR_LOG", "0")
	os.Setenv("NOTIFICATION_WARNING_LOG", "0")
	os.Setenv("NOTIFICATION_INFO_LOG", "1")
	os.Setenv("NOTIFICATION_SUCCESS_LOG", "0")
	os.Setenv("NOTIFICATIONS_SLACK_ENABLED", "0")

	Init()

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		SuccessLog("Benchmark success message")
	}
}
