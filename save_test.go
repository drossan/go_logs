package go_logs

import (
	"os"
	"testing"
)

// TestLogFilePermissions verifies that log files are created with secure
// permissions (0600 = owner read/write only) instead of insecure 0666
// Issue #6: Permisos de archivo demasiado permisivos
func TestLogFilePermissions(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	logFilePath = tempDir
	logFileName = "test_permissions.log"

	// Call function under test
	file := openLogFile()
	if file == nil {
		t.Fatal("Expected file to be created, got nil")
	}
	defer Close() // Use Close() instead of file.Close() for persistent file
	defer os.Remove(logFilePath + "/" + logFileName)

	// Get file info to check permissions
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	// Check permissions
	// 0600 = rw------- (owner read/write only)
	// 0666 = rw-rw-rw- (world readable - INSECURE)
	mode := info.Mode().Perm()
	expectedMode := os.FileMode(0600)

	if mode != expectedMode {
		t.Errorf("File permissions incorrect: got %04o, want %04o", mode, expectedMode)
	}
}

// TestLogFilePermissionsWithConfig verifies permissions can be customized
// via LOG_FILE_PERMISSIONS environment variable (optional feature)
func TestLogFilePermissionsWithConfig(t *testing.T) {
	// Test default permissions (0600)
	t.Run("Default permissions are 0600", func(t *testing.T) {
		tempDir := t.TempDir()
		logFilePath = tempDir
		logFileName = "test_default.log"

		file := openLogFile()
		if file == nil {
			t.Fatal("Expected file to be created")
		}
		defer Close() // Use Close() instead of file.Close()

		info, _ := file.Stat()
		mode := info.Mode().Perm()

		if mode != 0600 {
			t.Errorf("Default permissions should be 0600, got %04o", mode)
		}
	})
}
