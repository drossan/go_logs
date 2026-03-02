package go_logs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWithRotatingFile_WritesToFile tests that WithRotatingFile actually writes to the file
func TestWithRotatingFile_WritesToFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	// Create logger with rotating file
	logger, err := New(
		WithLevel(InfoLevel),
		WithRotatingFile(logFile, 10, 3),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Write some logs
	logger.Info("test message 1")
	logger.Info("test message 2")
	logger.Error("error message")

	// Read the file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Verify content
	contentStr := string(content)
	if !strings.Contains(contentStr, "test message 1") {
		t.Error("Expected 'test message 1' in log file")
	}
	if !strings.Contains(contentStr, "test message 2") {
		t.Error("Expected 'test message 2' in log file")
	}
	if !strings.Contains(contentStr, "error message") {
		t.Error("Expected 'error message' in log file")
	}
	if !strings.Contains(contentStr, "INFO") {
		t.Error("Expected 'INFO' level in log file")
	}
	if !strings.Contains(contentStr, "ERROR") {
		t.Error("Expected 'ERROR' level in log file")
	}
}

// TestWithRotatingFileEnhanced_WritesToFile tests the enhanced rotating file writer
func TestWithRotatingFileEnhanced_WritesToFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "enhanced.log")

	logger, err := New(
		WithLevel(DebugLevel),
		WithRotatingFileEnhanced(RotatingFileConfig{
			Filename:     logFile,
			MaxSizeMB:    10,
			MaxBackups:   3,
			RotationType: RotateSize,
		}),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "debug message") {
		t.Error("Expected 'debug message' in log file")
	}
	if !strings.Contains(contentStr, "info message") {
		t.Error("Expected 'info message' in log file")
	}
	if !strings.Contains(contentStr, "warn message") {
		t.Error("Expected 'warn message' in log file")
	}
}

// TestWithMultiOutput_WritesToAll tests that MultiWriter writes to all outputs
func TestWithMultiOutput_WritesToAll(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "multi.log")

	// Create file writer
	fileWriter, err := NewRotatingFileWriter(logFile, 10, 3)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}

	// Create capture buffer for console simulation
	buf := NewCaptureBuffer()

	// Create logger with multi-output
	logger, err := New(
		WithLevel(InfoLevel),
		WithMultiOutput(fileWriter, buf),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("multi output test")

	// Check buffer has content
	if !buf.Contains("multi output test") {
		t.Error("Expected message in buffer")
	}

	// Check file has content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !strings.Contains(string(content), "multi output test") {
		t.Error("Expected message in file")
	}
}

// TestRotatingFile_ImmediateFlush tests that logs are immediately flushed
func TestRotatingFile_ImmediateFlush(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "immediate.log")

	logger, err := New(
		WithLevel(InfoLevel),
		WithRotatingFile(logFile, 10, 3),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Write a single message
	logger.Info("immediate flush test")

	// Read immediately without closing or syncing
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// The content should be available immediately due to auto-sync
	if !strings.Contains(string(content), "immediate flush test") {
		t.Error("Expected immediate flush - message not found in file")
		t.Errorf("File content: %s", string(content))
	}
}

// TestRotatingFile_CreatesDirectory tests that the directory is created if it doesn't exist
func TestRotatingFile_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "deep", "path")
	logFile := filepath.Join(nestedDir, "test.log")

	logger, err := New(
		WithLevel(InfoLevel),
		WithRotatingFile(logFile, 10, 3),
	)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Info("directory creation test")

	// Check file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}

	// Check content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	if !strings.Contains(string(content), "directory creation test") {
		t.Error("Expected message in log file")
	}
}
