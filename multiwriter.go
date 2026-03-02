package go_logs

import (
	"io"
	"sync"
)

// MultiWriter writes to multiple writers simultaneously.
// This enables logging to both file and console at the same time.
//
// Example:
//
//	file, _ := go_logs.NewRotatingFileWriter("app.log", 100, 5)
//	multi := go_logs.NewMultiWriter(file, os.Stdout)
//	logger := go_logs.New(go_logs.WithOutput(multi))
type MultiWriter struct {
	writers []io.Writer
	mu      sync.RWMutex

	// errorHandler is called when a writer fails (optional)
	errorHandler func(error)
}

// NewMultiWriter creates a new MultiWriter that writes to all provided writers.
// If no writers are provided, writes are discarded (similar to io.Discard).
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// NewWriterWithErrorHandler creates a MultiWriter with custom error handling.
// When a write fails, the error handler is called but writing continues to other writers.
func NewWriterWithErrorHandler(errorHandler func(error), writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers:      writers,
		errorHandler: errorHandler,
	}
}

// Write writes the data to all writers.
// It returns the length of the data if at least one writer succeeded.
// Errors from individual writers are handled but don't fail the operation.
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	if len(mw.writers) == 0 {
		return len(p), nil // Behave like io.Discard
	}

	// Write to all writers, collecting errors
	var lastErr error
	successCount := 0

	for _, w := range mw.writers {
		wn, werr := w.Write(p)
		if werr != nil {
			lastErr = werr
			if mw.errorHandler != nil {
				mw.errorHandler(werr)
			}
		} else {
			successCount++
			// Track the minimum bytes written (should all be len(p))
			if wn > n {
				n = wn
			}
		}
	}

	// If at least one writer succeeded, return success
	if successCount > 0 {
		return n, nil
	}

	// All writers failed
	return 0, lastErr
}

// AddWriter adds a new writer to the multi-writer.
// This is thread-safe and can be called at runtime.
func (mw *MultiWriter) AddWriter(w io.Writer) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.writers = append(mw.writers, w)
}

// RemoveWriter removes a writer from the multi-writer.
// Writers are compared by pointer equality.
func (mw *MultiWriter) RemoveWriter(w io.Writer) {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	for i, writer := range mw.writers {
		if writer == w {
			mw.writers = append(mw.writers[:i], mw.writers[i+1:]...)
			break
		}
	}
}

// SetErrorHandler sets the error handler for write failures.
func (mw *MultiWriter) SetErrorHandler(handler func(error)) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	mw.errorHandler = handler
}

// Writers returns a copy of the current writers slice.
func (mw *MultiWriter) Writers() []io.Writer {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	result := make([]io.Writer, len(mw.writers))
	copy(result, mw.writers)
	return result
}

// Close closes all writers that implement io.Closer.
// Errors from individual closers are collected but don't fail the operation.
func (mw *MultiWriter) Close() error {
	mw.mu.Lock()
	defer mw.mu.Unlock()

	var lastErr error
	for _, w := range mw.writers {
		if closer, ok := w.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				lastErr = err
			}
		}
	}

	return lastErr
}

// Sync syncs all writers that implement the Syncer interface.
type Syncer interface {
	Sync() error
}

// Sync calls Sync on all writers that implement Syncer.
func (mw *MultiWriter) Sync() error {
	mw.mu.RLock()
	defer mw.mu.RUnlock()

	var lastErr error
	for _, w := range mw.writers {
		if syncer, ok := w.(Syncer); ok {
			if err := syncer.Sync(); err != nil {
				lastErr = err
			}
		}
	}

	return lastErr
}
