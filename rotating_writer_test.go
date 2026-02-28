package go_logs

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRotatingFileWriter_BasicWrite tests basic write functionality
func TestRotatingFileWriter_BasicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 100, 3) // 100MB, 3 backups
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	testMessage := "Test log message\n"
	n, err := writer.Write([]byte(testMessage))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(testMessage) {
		t.Errorf("Write returned %d, expected %d", n, len(testMessage))
	}

	// Flush to ensure data is written
	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify file exists and contains message
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != testMessage {
		t.Errorf("File content = %q, want %q", string(data), testMessage)
	}
}

// TestRotatingFileWriter_RotateOnSize tests rotation when max size is exceeded
func TestRotatingFileWriter_RotateOnSize(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	// Create writer with 1MB max size and 3 backups
	writer, err := NewRotatingFileWriter(filename, 1, 3) // 1MB, 3 backups
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Write data in chunks to trigger multiple rotations
	// Each write is 1.5MB, which will trigger rotation
	chunkSize := int64(1.5 * 1024 * 1024) // 1.5MB
	numWrites := 4

	for i := 0; i < numWrites; i++ {
		chunk := make([]byte, chunkSize)
		for j := range chunk {
			chunk[j] = byte('A' + i)
		}

		n, err := writer.Write(chunk)
		if err != nil {
			t.Fatalf("Write %d failed: %v", i, err)
		}
		if n != len(chunk) {
			t.Errorf("Write %d returned %d, expected %d", i, n, len(chunk))
		}
	}

	// Sync to flush
	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Check that at least .1 backup was created
	if _, err := os.Stat(filename + ".1"); os.IsNotExist(err) {
		t.Error("Backup file .1 was not created")
	}

	// We expect up to 3 backups (.1, .2, .3) but the exact number
	// depends on when rotation is triggered
	// Just verify that .4 doesn't exist (max 3 backups)
	if _, err := os.Stat(filename + ".4"); !os.IsNotExist(err) {
		t.Error("Backup file .4 should not exist (max 3 backups)")
	}
}

// TestRotatingFileWriter_MaxBackups tests that maxBackups limit is respected
func TestRotatingFileWriter_MaxBackups(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	maxBackups := 2
	writer, err := NewRotatingFileWriter(filename, 1, maxBackups) // 1MB, 2 backups
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Trigger 4 rotations to exceed maxBackups
	for i := 0; i < 4; i++ {
		largeData := make([]byte, 2*1024*1024) // 2MB
		for j := range largeData {
			largeData[j] = byte('A' + i)
		}
		if _, err := writer.Write(largeData); err != nil {
			t.Fatalf("Write %d failed: %v", i, err)
		}
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify only maxBackups exist
	for i := 1; i <= maxBackups; i++ {
		backup := filename + "." + string(rune('0'+i))
		if _, err := os.Stat(backup); os.IsNotExist(err) {
			t.Errorf("Backup file %s should exist", backup)
		}
	}

	// Verify .3 doesn't exist
	if _, err := os.Stat(filename + ".3"); !os.IsNotExist(err) {
		t.Error("Backup file .3 should not exist (max 2 backups)")
	}
}

// TestRotatingFileWriter_ManualRotate tests manual rotation via Rotate()
func TestRotatingFileWriter_ManualRotate(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 100, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Write initial data
	msg1 := "First message\n"
	if _, err := writer.Write([]byte(msg1)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Manual rotation
	if err := writer.Rotate(); err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}

	// Write new data
	msg2 := "Second message\n"
	if _, err := writer.Write([]byte(msg2)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify .1 contains first message
	data1, err := os.ReadFile(filename + ".1")
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	if string(data1) != msg1 {
		t.Errorf("Backup content = %q, want %q", string(data1), msg1)
	}

	// Verify current file contains second message
	data2, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data2) != msg2 {
		t.Errorf("Current file content = %q, want %q", string(data2), msg2)
	}
}

// TestRotatingFileWriter_ConcurrentWrites tests thread-safety
func TestRotatingFileWriter_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 1, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Write concurrently from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			message := []byte(strings.Repeat("X", 100*1024) + "\n") // 100KB chunks
			for j := 0; j < 100; j++ {
				if _, err := writer.Write(message); err != nil {
					t.Errorf("Goroutine %d write failed: %v", id, err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify file exists and has content
	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("File is empty after concurrent writes")
	}
}

// TestRotatingFileWriter_CloseIdempotent tests that Close is idempotent
func TestRotatingFileWriter_CloseIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 100, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// Close multiple times
	if err := writer.Close(); err != nil {
		t.Fatalf("First Close failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Second Close failed: %v", err)
	}
}

// TestRotatingFileWriter_Sync tests Sync functionality
func TestRotatingFileWriter_Sync(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 100, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	message := "Test message\n"
	if _, err := writer.Write([]byte(message)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Sync to flush buffer
	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify data was flushed
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != message {
		t.Errorf("File content = %q, want %q", string(data), message)
	}
}

// TestRotatingFileWriter_ResumeExistingFile tests resuming an existing log file
func TestRotatingFileWriter_ResumeExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	// Create initial file with content
	existingContent := "Existing content\n"
	if err := os.WriteFile(filename, []byte(existingContent), 0600); err != nil {
		t.Fatalf("Failed to create initial file: %v", err)
	}

	// Create writer - should resume existing file
	writer, err := NewRotatingFileWriter(filename, 100, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Write new content
	newContent := "New content\n"
	if _, err := writer.Write([]byte(newContent)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify both contents exist
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	expected := existingContent + newContent
	if string(data) != expected {
		t.Errorf("File content = %q, want %q", string(data), expected)
	}
}

// TestRotatingFileWriter_GetMaxSize tests getMaxSize
func TestRotatingFileWriter_GetMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	maxSizeMB := 50
	writer, err := NewRotatingFileWriter(filename, maxSizeMB, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	expectedMaxSize := int64(maxSizeMB) * 1024 * 1024
	if writer.GetMaxSize() != expectedMaxSize {
		t.Errorf("GetMaxSize() = %d, want %d", writer.GetMaxSize(), expectedMaxSize)
	}
}

// TestRotatingFileWriter_GetMaxBackups tests getMaxBackups
func TestRotatingFileWriter_GetMaxBackups(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	maxBackups := 5
	writer, err := NewRotatingFileWriter(filename, 100, maxBackups)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	if writer.GetMaxBackups() != maxBackups {
		t.Errorf("GetMaxBackups() = %d, want %d", writer.GetMaxBackups(), maxBackups)
	}
}

// TestRotatingFileWriter_WriteAfterRotate tests writing after rotation
func TestRotatingFileWriter_WriteAfterRotate(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 1, 3) // 1MB
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Write 2MB to trigger rotation
	largeData := make([]byte, 2*1024*1024)
	for i := range largeData {
		largeData[i] = 'A'
	}
	if _, err := writer.Write(largeData); err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Write more data after rotation
	smallData := []byte("Small message after rotation\n")
	if _, err := writer.Write(smallData); err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify current file has the small data
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if !bytes.HasSuffix(data, smallData) {
		t.Errorf("File should end with %q, got %q", smallData, data[len(data)-len(smallData):])
	}
}

// TestRotatingFileWriter_RotateWithExistingBackups tests rotation when backups already exist
func TestRotatingFileWriter_RotateWithExistingBackups(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	// Create existing backup files
	os.WriteFile(filename+".1", []byte("Backup 1\n"), 0600)
	os.WriteFile(filename+".2", []byte("Backup 2\n"), 0600)

	writer, err := NewRotatingFileWriter(filename, 1, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Write initial content
	msg1 := "Initial content\n"
	if _, err := writer.Write([]byte(msg1)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Trigger rotation
	if err := writer.Rotate(); err != nil {
		t.Fatalf("Rotate failed: %v", err)
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify old .1 became .2, old .2 became .3, current became .1
	// .1 should have msg1
	data1, _ := os.ReadFile(filename + ".1")
	if string(data1) != msg1 {
		t.Errorf(".1 content = %q, want %q", string(data1), msg1)
	}

	// .2 and .3 should still exist
	if _, err := os.Stat(filename + ".2"); os.IsNotExist(err) {
		t.Error(".2 backup should still exist")
	}
	if _, err := os.Stat(filename + ".3"); os.IsNotExist(err) {
		t.Error(".3 backup should still exist")
	}
}

// TestRotatingFileWriter_ImplementsIoWriter tests that RotatingFileWriter implements io.Writer
func TestRotatingFileWriter_ImplementsIoWriter(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 100, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Test that it implements io.Writer interface
	var _ io.Writer = writer

	// Can use with functions expecting io.Writer
	message := "Test via io.Writer\n"
	if _, err := writer.Write([]byte(message)); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
}

// TestRotatingFileWriter_WriteAfterClose tests that writes fail after Close
func TestRotatingFileWriter_WriteAfterClose(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 100, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Write after close should return error
	message := "After close\n"
	n, err := writer.Write([]byte(message))

	// We expect error and no bytes written
	if err == nil {
		t.Error("Write after close should return error")
	}
	if n != 0 {
		t.Errorf("Write after close should return 0 bytes, got %d", n)
	}
}

// TestRotatingFileWriter_Performance tests performance characteristics
func TestRotatingFileWriter_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.log")

	writer, err := NewRotatingFileWriter(filename, 10, 3) // 10MB
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Benchmark write speed
	message := []byte("Test log message for performance testing\n")
	iterations := 10000

	start := time.Now()
	for i := 0; i < iterations; i++ {
		if _, err := writer.Write(message); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}
	elapsed := time.Since(start)

	t.Logf("Wrote %d messages in %v (%.2f msg/sec)",
		iterations, elapsed, float64(iterations)/elapsed.Seconds())

	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
}
