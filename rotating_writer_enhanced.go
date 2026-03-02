package go_logs

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// RotationType defines when log files should be rotated
type RotationType int

const (
	// RotateSize rotates when file exceeds MaxSizeMB (default)
	RotateSize RotationType = iota
	// RotateDaily rotates at midnight (00:00)
	RotateDaily
	// RotateHourly rotates at the start of each hour
	RotateHourly
)

// String returns the string representation of RotationType
func (r RotationType) String() string {
	switch r {
	case RotateSize:
		return "size"
	case RotateDaily:
		return "daily"
	case RotateHourly:
		return "hourly"
	default:
		return "unknown"
	}
}

// RotatingFileConfig holds all configuration for RotatingFileWriter
type RotatingFileConfig struct {
	// Filename is the path to the log file
	Filename string

	// MaxSizeMB is the maximum size in megabytes before rotation (for RotateSize)
	MaxSizeMB int

	// MaxBackups is the maximum number of backup files to keep
	MaxBackups int

	// RotationType determines when rotation occurs (size, daily, hourly)
	RotationType RotationType

	// Compress enables gzip compression of rotated files
	Compress bool

	// MaxAge is the maximum number of days to keep old log files (0 = no limit)
	MaxAge int

	// LocalTime determines if local time or UTC is used for rotation timestamps
	LocalTime bool
}

// EnhancedRotatingFileWriter extends RotatingFileWriter with time-based rotation and compression
type EnhancedRotatingFileWriter struct {
	config RotatingFileConfig

	// currentSize tracks the current size of the active log file
	currentSize int64

	// file is the active log file handle
	file *os.File

	// writer is the buffered writer wrapping the file
	writer *bufio.Writer

	// rotationTime tracks when the next time-based rotation should occur
	rotationTime time.Time

	// mu protects all operations for thread-safety
	mu sync.Mutex
}

// NewRotatingFileWriterWithRotation creates a writer with specified rotation type
func NewRotatingFileWriterWithRotation(filename string, maxSizeMB int, maxBackups int, rotationType RotationType) (*EnhancedRotatingFileWriter, error) {
	return NewRotatingFileWriterWithConfig(RotatingFileConfig{
		Filename:     filename,
		MaxSizeMB:    maxSizeMB,
		MaxBackups:   maxBackups,
		RotationType: rotationType,
	})
}

// NewRotatingFileWriterWithConfig creates a writer with full configuration
func NewRotatingFileWriterWithConfig(config RotatingFileConfig) (*EnhancedRotatingFileWriter, error) {
	if config.Filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	if config.MaxSizeMB <= 0 {
		return nil, fmt.Errorf("maxSizeMB must be positive, got %d", config.MaxSizeMB)
	}
	if config.MaxBackups < 0 {
		return nil, fmt.Errorf("maxBackups must be non-negative, got %d", config.MaxBackups)
	}

	w := &EnhancedRotatingFileWriter{
		config: config,
	}

	// Calculate next rotation time for time-based rotation
	if config.RotationType != RotateSize {
		w.rotationTime = w.calculateNextRotationTime()
	}

	if err := w.openFile(); err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return w, nil
}

// calculateNextRotationTime returns the next time when rotation should occur
func (w *EnhancedRotatingFileWriter) calculateNextRotationTime() time.Time {
	now := time.Now()
	if !w.config.LocalTime {
		now = now.UTC()
	}

	switch w.config.RotationType {
	case RotateDaily:
		// Next midnight
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	case RotateHourly:
		// Next hour
		return time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())
	default:
		return time.Time{} // No time-based rotation
	}
}

// Write implements io.Writer
func (w *EnhancedRotatingFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer == nil {
		return 0, fmt.Errorf("write to closed EnhancedRotatingFileWriter")
	}

	writeLen := int64(len(p))

	// Check if rotation is needed
	needRotate := false

	// Size-based rotation check
	if w.config.RotationType == RotateSize && w.currentSize+writeLen > int64(w.config.MaxSizeMB)*1024*1024 {
		needRotate = true
	}

	// Time-based rotation check
	if w.config.RotationType != RotateSize && time.Now().After(w.rotationTime) {
		needRotate = true
	}

	if needRotate {
		if err := w.rotate(); err != nil {
			return 0, fmt.Errorf("rotation failed: %w", err)
		}
	}

	n, err = w.writer.Write(p)
	if err != nil {
		return n, fmt.Errorf("write failed: %w", err)
	}

	w.currentSize += int64(n)
	return n, nil
}

// Rotate forces a rotation
func (w *EnhancedRotatingFileWriter) Rotate() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.rotate()
}

// rotate is the internal rotation implementation
func (w *EnhancedRotatingFileWriter) rotate() error {
	// Flush and close current file
	if w.writer != nil {
		if err := w.writer.Flush(); err != nil {
			return fmt.Errorf("flush failed: %w", err)
		}
	}
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("close failed: %w", err)
		}
	}

	// Generate backup filename with timestamp for time-based rotation
	var backupName string
	if w.config.RotationType != RotateSize {
		now := time.Now()
		if !w.config.LocalTime {
			now = now.UTC()
		}
		timestamp := now.Format("2006-01-02_15-04-05")
		ext := filepath.Ext(w.config.Filename)
		base := strings.TrimSuffix(w.config.Filename, ext)
		backupName = fmt.Sprintf("%s-%s%s", base, timestamp, ext)
	}

	// Rotate existing backups
	w.rotateBackups()

	// Rename current file to backup
	if backupName == "" {
		backupName = fmt.Sprintf("%s.1", w.config.Filename)
	}

	if _, err := os.Stat(w.config.Filename); err == nil {
		// Compress old file if enabled
		if w.config.Compress {
			if err := w.compressFile(w.config.Filename, backupName+".gz"); err != nil {
				// If compression fails, just rename without compression
				if err := os.Rename(w.config.Filename, backupName); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to rotate current file: %w", err)
				}
			}
		} else {
			if err := os.Rename(w.config.Filename, backupName); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to rotate current file: %w", err)
			}
		}
	}

	// Clean old files based on MaxAge and MaxBackups
	w.cleanOldFiles()

	// Update rotation time for time-based rotation
	if w.config.RotationType != RotateSize {
		w.rotationTime = w.calculateNextRotationTime()
	}

	// Open new file
	return w.openFile()
}

// rotateBackups rotates existing backup files
func (w *EnhancedRotatingFileWriter) rotateBackups() {
	if w.config.RotationType != RotateSize {
		return // Time-based rotation doesn't use numbered backups
	}

	// Delete oldest backup if at max
	oldestBackup := fmt.Sprintf("%s.%d", w.config.Filename, w.config.MaxBackups)
	if w.config.Compress {
		oldestBackup += ".gz"
	}
	os.Remove(oldestBackup)

	// Rotate existing backups
	for i := w.config.MaxBackups - 1; i >= 1; i-- {
		oldBackup := fmt.Sprintf("%s.%d", w.config.Filename, i)
		newBackup := fmt.Sprintf("%s.%d", w.config.Filename, i+1)
		if w.config.Compress {
			oldBackup += ".gz"
			newBackup += ".gz"
		}

		if _, err := os.Stat(oldBackup); err == nil {
			os.Rename(oldBackup, newBackup)
		}
	}
}

// compressFile compresses a file using gzip
func (w *EnhancedRotatingFileWriter) compressFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(dstFile)
	defer gzWriter.Close()

	// Copy data
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := srcFile.Read(buf)
		if n > 0 {
			if _, writeErr := gzWriter.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			break
		}
	}

	// Remove original file after successful compression
	os.Remove(src)

	return nil
}

// cleanOldFiles removes old files based on MaxAge and MaxBackups
func (w *EnhancedRotatingFileWriter) cleanOldFiles() {
	if w.config.MaxAge <= 0 && w.config.MaxBackups <= 0 {
		return
	}

	dir := filepath.Dir(w.config.Filename)
	base := filepath.Base(w.config.Filename)
	ext := filepath.Ext(base)
	prefix := strings.TrimSuffix(base, ext)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var backups []os.DirEntry
	for _, entry := range entries {
		name := entry.Name()
		// Match backup files (with timestamp or numbered)
		if strings.HasPrefix(name, prefix) && name != base {
			backups = append(backups, entry)
		}
	}

	// Sort by modification time (oldest first)
	type fileInfo struct {
		entry    os.DirEntry
		modTime  time.Time
		fullPath string
	}

	var files []fileInfo
	for _, entry := range backups {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			entry:    entry,
			modTime:  info.ModTime(),
			fullPath: filepath.Join(dir, entry.Name()),
		})
	}

	// Sort by modification time
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove files based on MaxAge
	if w.config.MaxAge > 0 {
		cutoff := time.Now().AddDate(0, 0, -w.config.MaxAge)
		for _, f := range files {
			if f.modTime.Before(cutoff) {
				os.Remove(f.fullPath)
			}
		}
	}

	// Re-read directory after MaxAge cleanup
	entries, _ = os.ReadDir(dir)
	files = nil
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && name != base {
			info, _ := entry.Info()
			files = append(files, fileInfo{
				entry:    entry,
				modTime:  info.ModTime(),
				fullPath: filepath.Join(dir, name),
			})
		}
	}

	// Sort again
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].modTime.After(files[j].modTime) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Remove files based on MaxBackups
	if w.config.MaxBackups > 0 && len(files) > w.config.MaxBackups {
		for i := 0; i < len(files)-w.config.MaxBackups; i++ {
			os.Remove(files[i].fullPath)
		}
	}
}

// openFile opens the log file
func (w *EnhancedRotatingFileWriter) openFile() error {
	dir := filepath.Dir(w.config.Filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	file, err := os.OpenFile(w.config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	w.file = file
	w.writer = bufio.NewWriter(file)

	info, err := file.Stat()
	if err != nil {
		w.currentSize = 0
		return nil
	}

	w.currentSize = info.Size()
	return nil
}

// Sync flushes buffered data
func (w *EnhancedRotatingFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer == nil {
		return nil
	}

	if err := w.writer.Flush(); err != nil {
		return err
	}

	if w.file != nil {
		return w.file.Sync()
	}

	return nil
}

// Close closes the writer
func (w *EnhancedRotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.writer == nil {
		return nil
	}

	if err := w.writer.Flush(); err != nil {
		return err
	}

	if err := w.file.Close(); err != nil {
		return err
	}

	w.writer = nil
	w.file = nil

	return nil
}

// GetConfig returns the current configuration
func (w *EnhancedRotatingFileWriter) GetConfig() RotatingFileConfig {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.config
}

// GetCurrentSize returns the current file size
func (w *EnhancedRotatingFileWriter) GetCurrentSize() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.currentSize
}

// GetNextRotationTime returns the next scheduled rotation time (for time-based rotation)
func (w *EnhancedRotatingFileWriter) GetNextRotationTime() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.rotationTime
}
