package go_logs

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestMultiWriter_BasicWrite verifies basic multi-output functionality
func TestMultiWriter_BasicWrite(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	mw := NewMultiWriter(&buf1, &buf2)

	data := []byte("test message")
	n, err := mw.Write(data)

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	// Both buffers should contain the message
	if buf1.String() != "test message" {
		t.Errorf("buf1: expected 'test message', got %q", buf1.String())
	}
	if buf2.String() != "test message" {
		t.Errorf("buf2: expected 'test message', got %q", buf2.String())
	}
}

// TestMultiWriter_Empty verifies empty writer handling
func TestMultiWriter_Empty(t *testing.T) {
	mw := NewMultiWriter()

	data := []byte("test")
	n, err := mw.Write(data)

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}
}

// TestMultiWriter_ThreeWriters verifies multiple outputs
func TestMultiWriter_ThreeWriters(t *testing.T) {
	var buf1, buf2, buf3 bytes.Buffer
	mw := NewMultiWriter(&buf1, &buf2, &buf3)

	data := []byte("multi-output test")
	n, err := mw.Write(data)

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}

	// All three buffers should contain the message
	for i, buf := range []bytes.Buffer{buf1, buf2, buf3} {
		if buf.String() != "multi-output test" {
			t.Errorf("buf%d: expected 'multi-output test', got %q", i+1, buf.String())
		}
	}
}

// TestMultiWriter_WithErrorWriter verifies error handling
func TestMultiWriter_WithErrorWriter(t *testing.T) {
	var buf bytes.Buffer
	errWriter := &errorWriter{err: io.ErrClosedPipe}
	mw := NewWriterWithErrorHandler(func(err error) {
		// Error handler called - we expect this
		if err != io.ErrClosedPipe {
			t.Errorf("Expected ErrClosedPipe, got %v", err)
		}
	}, &buf, errWriter)

	data := []byte("test")
	_, err := mw.Write(data)

	// Write should succeed even if one writer fails
	if err != nil {
		t.Errorf("Write should not fail: %v", err)
	}
}

// errorWriter is a test helper that always returns an error
type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

// TestTimeBasedRotation verifies daily rotation
func TestTimeBasedRotation(t *testing.T) {
	tempDir := t.TempDir()
	logFile := tempDir + "/test.log"

	// Create rotating writer with daily rotation
	rw, err := NewRotatingFileWriterWithRotation(logFile, 100, 3, RotateDaily)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer rw.Close()

	// Write some data
	_, err = rw.Write([]byte("test message\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}
}

// TestRotationType verifies rotation type configuration
func TestRotationType(t *testing.T) {
	tests := []struct {
		name     string
		rotate   RotationType
		expected string
	}{
		{"size", RotateSize, "size"},
		{"daily", RotateDaily, "daily"},
		{"hourly", RotateHourly, "hourly"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rotate.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.rotate.String())
			}
		})
	}
}

// TestCompressionEnabled verifies gzip compression
func TestCompressionEnabled(t *testing.T) {
	tempDir := t.TempDir()
	logFile := tempDir + "/test.log"

	// Create rotating writer with compression enabled
	rw, err := NewRotatingFileWriterWithConfig(RotatingFileConfig{
		Filename:     logFile,
		MaxSizeMB:    1,
		MaxBackups:   3,
		RotationType: RotateSize,
		Compress:     true,
	})
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer rw.Close()

	// Write enough data to trigger rotation
	largeData := strings.Repeat("x", 1024*1024+1) // > 1MB
	_, err = rw.Write([]byte(largeData))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Force rotation
	rw.Rotate()

	// Check for compressed backup (.gz file)
	files, _ := os.ReadDir(tempDir)
	foundGz := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".gz") {
			foundGz = true
			break
		}
	}

	if !foundGz {
		t.Error("Expected compressed backup file (.gz)")
	}
}

// TestRotatingFileConfig validates configuration
func TestRotatingFileConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  RotatingFileConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: RotatingFileConfig{
				Filename:   "/tmp/test.log",
				MaxSizeMB:  100,
				MaxBackups: 5,
			},
			wantErr: false,
		},
		{
			name: "empty filename",
			config: RotatingFileConfig{
				Filename:   "",
				MaxSizeMB:  100,
				MaxBackups: 5,
			},
			wantErr: true,
		},
		{
			name: "invalid maxSize",
			config: RotatingFileConfig{
				Filename:   "/tmp/test.log",
				MaxSizeMB:  0,
				MaxBackups: 5,
			},
			wantErr: true,
		},
		{
			name: "negative backups",
			config: RotatingFileConfig{
				Filename:   "/tmp/test.log",
				MaxSizeMB:  100,
				MaxBackups: -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRotatingFileWriterWithConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRotatingFileWriterWithConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestMultiWriterWithFileAndConsole verifies file + console output
func TestMultiWriterWithFileAndConsole(t *testing.T) {
	tempDir := t.TempDir()
	logFile := tempDir + "/multi.log"

	// Create file writer
	fileWriter, err := NewRotatingFileWriter(logFile, 100, 3)
	if err != nil {
		t.Fatalf("Failed to create file writer: %v", err)
	}
	defer fileWriter.Close()

	// Create multi-writer with file + console (buffer for test)
	var consoleBuf bytes.Buffer
	mw := NewMultiWriter(fileWriter, &consoleBuf)

	// Write to both
	data := []byte("log to file and console\n")
	_, err = mw.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify console output
	if !strings.Contains(consoleBuf.String(), "log to file and console") {
		t.Error("Console output missing")
	}

	// Sync and verify file output
	fileWriter.Sync()
	fileContent, _ := os.ReadFile(logFile)
	if !strings.Contains(string(fileContent), "log to file and console") {
		t.Error("File output missing")
	}
}
