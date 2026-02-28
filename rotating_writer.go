package go_logs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// RotatingFileWriter implements io.Writer with automatic log file rotation by size.
//
// Rotation occurs when the current log file exceeds maxSize bytes. Backups are
// created with incrementing suffixes (.1, .2, .3, etc.) up to maxBackups.
//
// This is a zero-dependency implementation using only Go standard library.
//
// Example:
//   writer, err := go_logs.NewRotatingFileWriter("app.log", 100, 3)
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer writer.Close()
//
//   logger := go_logs.New(go_logs.WithOutput(writer))
type RotatingFileWriter struct {
	// filename is the base name of the log file (without suffix)
	filename string

	// maxSize is the maximum size in bytes before rotation occurs
	maxSize int64

	// maxBackups is the maximum number of backup files to keep
	maxBackups int

	// currentSize tracks the current size of the active log file
	currentSize int64

	// file is the active log file handle
	file *os.File

	// writer is the buffered writer wrapping the file
	writer *bufio.Writer

	// mu protects all operations for thread-safety
	mu sync.Mutex
}

// NewRotatingFileWriter creates a new RotatingFileWriter with the specified parameters.
//
// Parameters:
//   - filename: Path to the log file (e.g., "app.log" or "/var/log/app.log")
//   - maxSizeMB: Maximum size in megabytes before rotation (e.g., 100 for 100MB)
//   - maxBackups: Maximum number of backup files to keep (e.g., 3 keeps app.log.1, .2, .3)
//
// The writer will resume appending to an existing file if it exists.
// If the file is new, it will be created with permissions 0600 (owner read/write only).
//
// Returns an error if the file cannot be created or opened.
func NewRotatingFileWriter(filename string, maxSizeMB int, maxBackups int) (*RotatingFileWriter, error) {
	if maxSizeMB <= 0 {
		return nil, fmt.Errorf("maxSizeMB must be positive, got %d", maxSizeMB)
	}
	if maxBackups < 0 {
		return nil, fmt.Errorf("maxBackups must be non-negative, got %d", maxBackups)
	}

	w := &RotatingFileWriter{
		filename:   filename,
		maxSize:    int64(maxSizeMB) * 1024 * 1024, // Convert MB to bytes
		maxBackups: maxBackups,
	}

	if err := w.openFile(); err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return w, nil
}

// Write implements io.Writer. It writes the byte slice to the log file,
// rotating if necessary before the write.
//
// Write is thread-safe and can be called concurrently from multiple goroutines.
// If the write would cause the file to exceed maxSize, rotation occurs first.
//
// If the writer has been closed, Write returns an error.
//
// Returns the number of bytes written and any error encountered.
func (w *RotatingFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if writer is closed
	if w.writer == nil {
		return 0, fmt.Errorf("write to closed RotatingFileWriter")
	}

	writeLen := int64(len(p))

	// Check if we need to rotate BEFORE writing
	// Rotation is triggered if: current file size + new data > max size
	if w.currentSize+writeLen > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, fmt.Errorf("rotation failed: %w", err)
		}
	}

	// Write to buffered writer
	n, err = w.writer.Write(p)
	if err != nil {
		return n, fmt.Errorf("write failed: %w", err)
	}

	// Update current size (actual bytes written, not necessarily len(p))
	w.currentSize += int64(n)

	return n, nil
}

// Rotate performs a manual rotation of the log file.
//
// This method can be called to force rotation even if maxSize hasn't been
// exceeded (e.g., for SIGHUP signal handling in Unix daemons).
//
// Rotation follows this sequence:
// 1. Close and flush current log file
// 2. Rename backups: .3 -> .4, .2 -> .3, .1 -> .2 (if they exist)
// 3. Rename current file: app.log -> app.log.1
// 4. Create new empty app.log
// 5. Delete backup .4 if maxBackups = 3 (keep only .1, .2, .3)
//
// Returns an error if any step fails.
func (w *RotatingFileWriter) Rotate() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.rotate()
}

// rotate is the internal rotation implementation (caller must hold mu lock)
func (w *RotatingFileWriter) rotate() error {
	// Flush and close current file
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flush failed: %w", err)
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	// Rotate existing backups: .3 -> .4, .2 -> .3, .1 -> .2
	// We iterate backwards to avoid overwriting files
	for i := w.maxBackups - 1; i >= 1; i-- {
		oldBackup := fmt.Sprintf("%s.%d", w.filename, i)
		newBackup := fmt.Sprintf("%s.%d", w.filename, i+1)

		// Only rename if the old backup exists
		if _, err := os.Stat(oldBackup); err == nil {
			if err := os.Rename(oldBackup, newBackup); err != nil {
				return fmt.Errorf("failed to rotate backup %d: %w", i, err)
			}
		}
	}

	// Rename current file to .1 (first backup)
	backupName := fmt.Sprintf("%s.1", w.filename)
	if err := os.Rename(w.filename, backupName); err != nil {
		// If rename fails (e.g., file doesn't exist), we can continue
		// This can happen if the file was just created
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to rotate current file: %w", err)
		}
	}

	// Open new file
	if err := w.openFile(); err != nil {
		return fmt.Errorf("failed to open new file: %w", err)
	}

	return nil
}

// openFile opens (or creates) the log file and initializes the buffered writer.
// This method should be called with mu lock held.
func (w *RotatingFileWriter) openFile() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(w.filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Open file in append mode, create if doesn't exist
	// Use 0600 permissions (owner read/write only) for security
	file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	w.file = file
	w.writer = bufio.NewWriter(file)

	// Get current file size to initialize currentSize
	info, err := file.Stat()
	if err != nil {
		// If stat fails, assume empty file
		w.currentSize = 0
		return nil
	}

	w.currentSize = info.Size()
	return nil
}

// Sync flushes any buffered data to the underlying file.
//
// This method should be called before exiting the application to ensure
// all buffered log messages are written to disk.
//
// Returns an error if flushing fails.
func (w *RotatingFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer == nil {
		return nil // Already closed
	}

	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flush failed: %w", err)
	}

	// Also sync the file descriptor to disk
	if w.file != nil {
		if err := w.file.Sync(); err != nil {
			return fmt.Errorf("file sync failed: %w", err)
		}
	}

	return nil
}

// Close flushes any buffered data and closes the log file.
//
// Close is idempotent and can be called multiple times safely.
// After Close is called, Write operations will fail.
//
// Returns an error if flushing or closing fails.
func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer == nil {
		return nil // Already closed
	}

	// Flush buffered data
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flush failed: %w", err)
	}

	// Close file
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}

	// Mark as closed
	w.writer = nil
	w.file = nil

	return nil
}

// GetMaxSize returns the maximum file size in bytes before rotation occurs.
//
// This is useful for testing and monitoring purposes.
func (w *RotatingFileWriter) GetMaxSize() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.maxSize
}

// GetMaxBackups returns the maximum number of backup files to keep.
//
// This is useful for testing and monitoring purposes.
func (w *RotatingFileWriter) GetMaxBackups() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.maxBackups
}
