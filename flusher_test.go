package go_logs

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// The package writers implement Flusher.
var (
	_ Flusher = (*RotatingFileWriter)(nil)
	_ Flusher = (*EnhancedRotatingFileWriter)(nil)
	_ Flusher = (*MultiWriter)(nil)
	_ Flusher = (*SamplingWriter)(nil)
)

// flusher is the runtime shape of Flusher, so behaviour tests do not depend on
// the exported name.
type flusher interface{ Flush() error }

func newOutputLogger(t *testing.T, out io.Writer) Logger {
	t.Helper()
	logger, err := New(WithOutput(out), WithLevel(InfoLevel))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return logger
}

func newTestRotatingWriter(t *testing.T) (*RotatingFileWriter, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app.log")
	w, err := NewRotatingFileWriter(path, 10, 3)
	if err != nil {
		t.Fatalf("NewRotatingFileWriter() error = %v", err)
	}
	t.Cleanup(func() { w.Close() })
	return w, path
}

func fileContains(t *testing.T, path, want string) bool {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	return strings.Contains(string(content), want)
}

// Feature: Escribir una entrada no fuerza fsync

func TestWriteEntry_HundredEntriesNoSync(t *testing.T) {
	spy := &spyWriter{}
	logger := newOutputLogger(t, spy)

	for i := 0; i < 100; i++ {
		logger.Info("entry", Int("i", i))
	}

	writes, flushes, syncs := spy.counts()
	if syncs != 0 {
		t.Errorf("Sync calls = %d, want 0", syncs)
	}
	if writes != 100 {
		t.Errorf("Write calls = %d, want 100", writes)
	}
	if flushes != 100 {
		t.Errorf("Flush calls = %d, want 100", flushes)
	}
}

func TestWriteEntry_RotatingFileReadableImmediately(t *testing.T) {
	w, path := newTestRotatingWriter(t)
	logger := newOutputLogger(t, w)

	logger.Info("línea 1")

	if !fileContains(t, path, "línea 1") {
		t.Error("file does not contain \"línea 1\" before Sync/Close")
	}
}

func TestWriteEntry_MultiWriterFileAndBuffer(t *testing.T) {
	w, path := newTestRotatingWriter(t)
	var buf bytes.Buffer
	logger := newOutputLogger(t, NewMultiWriter(w, &buf))

	logger.Info("multi")

	if !fileContains(t, path, "multi") {
		t.Error("file does not contain \"multi\"")
	}
	if !strings.Contains(buf.String(), "multi") {
		t.Error("buffer does not contain \"multi\"")
	}
}

func TestWriteEntry_SamplingWriterPropagatesFlush(t *testing.T) {
	w, path := newTestRotatingWriter(t)
	// A threshold far above the number of writes lets every entry through.
	logger := newOutputLogger(t, NewSamplingWriter(w, 1000, time.Minute))

	logger.Info("sampled")

	if !fileContains(t, path, "sampled") {
		t.Error("file does not contain \"sampled\"")
	}
}

func TestWriteEntry_FlushErrorIsNotPropagated(t *testing.T) {
	spy := &spyWriter{flushErr: errors.New("flush failed")}
	logger := newOutputLogger(t, spy)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 3; i++ {
			logger.Info("msg", Int("i", i))
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("logging blocked after a Flush error")
	}

	writes, flushes, _ := spy.counts()
	if writes != 3 {
		t.Errorf("Write calls = %d, want 3", writes)
	}
	if flushes != 3 {
		t.Errorf("Flush calls = %d, want 3", flushes)
	}
}

func TestMultiWriter_EmptyFlushAndSync(t *testing.T) {
	var mw io.Writer = NewMultiWriter()

	f, ok := mw.(flusher)
	if !ok {
		t.Fatal("MultiWriter does not implement Flush() error")
	}
	if err := f.Flush(); err != nil {
		t.Errorf("Flush() = %v, want nil", err)
	}
	if err := mw.(Syncer).Sync(); err != nil {
		t.Errorf("Sync() = %v, want nil", err)
	}
}

func TestSamplingWriter_DelegatesFlushAndSync(t *testing.T) {
	spy := &spyWriter{}
	var sw io.Writer = NewSamplingWriter(spy, 1000, time.Minute)

	f, ok := sw.(flusher)
	if !ok {
		t.Fatal("SamplingWriter does not implement Flush() error")
	}
	s, ok := sw.(Syncer)
	if !ok {
		t.Fatal("SamplingWriter does not implement Sync() error")
	}
	f.Flush()
	s.Sync()

	if _, flushes, syncs := spy.counts(); flushes != 1 || syncs != 1 {
		t.Errorf("wrapped Flush/Sync calls = %d/%d, want 1/1", flushes, syncs)
	}

	// Wrapping a writer without Flush/Sync is a no-op, not an error.
	var plain io.Writer = NewSamplingWriter(&bytes.Buffer{}, 1000, time.Minute)
	if err := plain.(flusher).Flush(); err != nil {
		t.Errorf("Flush() on plain writer = %v, want nil", err)
	}
	if err := plain.(Syncer).Sync(); err != nil {
		t.Errorf("Sync() on plain writer = %v, want nil", err)
	}
}

func TestRotatingWriters_FlushWritesToFileWithoutClose(t *testing.T) {
	rw, path := newTestRotatingWriter(t)
	w, ok := io.Writer(rw).(interface {
		flusher
		io.WriteCloser
	})
	if !ok {
		t.Fatal("RotatingFileWriter does not implement Flush() error")
	}
	w.Write([]byte("rotating\n"))
	if err := w.Flush(); err != nil {
		t.Fatalf("RotatingFileWriter.Flush() = %v", err)
	}
	if !fileContains(t, path, "rotating") {
		t.Error("RotatingFileWriter: file does not contain data after Flush")
	}

	epath := filepath.Join(t.TempDir(), "enhanced.log")
	erw, err := NewRotatingFileWriterWithConfig(RotatingFileConfig{Filename: epath, MaxSizeMB: 10, MaxBackups: 3})
	if err != nil {
		t.Fatalf("NewRotatingFileWriterWithConfig() error = %v", err)
	}
	defer erw.Close()
	ew, ok := io.Writer(erw).(interface {
		flusher
		io.WriteCloser
	})
	if !ok {
		t.Fatal("EnhancedRotatingFileWriter does not implement Flush() error")
	}
	ew.Write([]byte("enhanced\n"))
	if err := ew.Flush(); err != nil {
		t.Fatalf("EnhancedRotatingFileWriter.Flush() = %v", err)
	}
	if !fileContains(t, epath, "enhanced") {
		t.Error("EnhancedRotatingFileWriter: file does not contain data after Flush")
	}

	// Flush after Close is a no-op.
	w.Close()
	ew.Close()
	if err := w.Flush(); err != nil {
		t.Errorf("RotatingFileWriter.Flush() after Close = %v, want nil", err)
	}
	if err := ew.Flush(); err != nil {
		t.Errorf("EnhancedRotatingFileWriter.Flush() after Close = %v, want nil", err)
	}
}

// Feature: Logger.Sync flushea y sincroniza el output

func TestLoggerSync_FlushesBufioWriter(t *testing.T) {
	var buf bytes.Buffer
	logger := newOutputLogger(t, bufio.NewWriter(&buf))

	logger.Info("pendiente")
	if err := logger.Sync(); err != nil {
		t.Fatalf("Sync() = %v", err)
	}

	if !strings.Contains(buf.String(), "pendiente") {
		t.Errorf("underlying buffer = %q, want it to contain \"pendiente\"", buf.String())
	}
}

func TestLoggerSync_CallsOutputSyncOnce(t *testing.T) {
	spy := &spyWriter{}
	logger := newOutputLogger(t, spy)

	for i := 0; i < 10; i++ {
		logger.Info("msg")
	}
	logger.Sync()

	if _, _, syncs := spy.counts(); syncs != 1 {
		t.Errorf("Sync calls = %d, want 1", syncs)
	}
}

func TestLoggerSync_PlainWriter(t *testing.T) {
	var buf bytes.Buffer
	logger := newOutputLogger(t, &buf)

	logger.Info("plain")
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() = %v, want nil", err)
	}
	if !strings.Contains(buf.String(), "plain") {
		t.Error("buffer does not contain \"plain\"")
	}
}

func TestLoggerSync_IgnorableErrors(t *testing.T) {
	diskFull := errors.New("disk full")
	pathErr := func(errno syscall.Errno) error {
		return &os.PathError{Op: "sync", Path: "/dev/stdout", Err: errno}
	}

	tests := []struct {
		name    string
		syncErr error
		want    error
	}{
		{"EINVAL", pathErr(syscall.EINVAL), nil},
		{"ENOTTY", pathErr(syscall.ENOTTY), nil},
		{"EBADF", pathErr(syscall.EBADF), nil},
		{"ErrClosedPipe", io.ErrClosedPipe, io.ErrClosedPipe},
		{"disk full", diskFull, diskFull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newOutputLogger(t, &spyWriter{syncErr: tt.syncErr})
			err := logger.Sync()
			if tt.want == nil {
				if err != nil {
					t.Errorf("Sync() = %v, want nil", err)
				}
				return
			}
			if !errors.Is(err, tt.want) {
				t.Errorf("Sync() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestLoggerSync_Stdout(t *testing.T) {
	logger := newOutputLogger(t, os.Stdout)
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() on os.Stdout = %v, want nil", err)
	}
}

func TestLoggerSync_MultiWriterIgnoresTerminalError(t *testing.T) {
	w, _ := newTestRotatingWriter(t)
	spy := &spyWriter{syncErr: &os.PathError{Op: "sync", Path: "/dev/stdout", Err: syscall.EBADF}}
	logger := newOutputLogger(t, NewMultiWriter(w, spy))

	logger.Info("msg")
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() = %v, want nil", err)
	}
}

func TestLoggerSync_MultiWriterPropagatesRealErrors(t *testing.T) {
	ebadf := &spyWriter{syncErr: &os.PathError{Op: "sync", Path: "/dev/stdout", Err: syscall.EBADF}}
	orders := map[string][]io.Writer{
		"valid then failing":    {&bytes.Buffer{}, &spyWriter{syncErr: errors.New("disk full")}},
		"failing then terminal": {&spyWriter{syncErr: errors.New("disk full")}, ebadf},
		"terminal then failing": {ebadf, &spyWriter{syncErr: errors.New("disk full")}},
	}
	for name, writers := range orders {
		t.Run(name, func(t *testing.T) {
			logger := newOutputLogger(t, NewMultiWriter(writers...))
			err := logger.Sync()
			if err == nil || !strings.Contains(err.Error(), "disk full") {
				t.Errorf("Sync() = %v, want an error mentioning \"disk full\"", err)
			}
		})
	}
}

func TestLoggerSync_ConcurrentWithLogging(t *testing.T) {
	w, _ := newTestRotatingWriter(t)
	logger := newOutputLogger(t, w)

	stop := make(chan struct{})
	var syncer sync.WaitGroup
	syncer.Add(1)
	go func() {
		defer syncer.Done()
		for {
			select {
			case <-stop:
				return
			default:
				logger.Sync()
			}
		}
	}()

	var loggers sync.WaitGroup
	for g := 0; g < 10; g++ {
		loggers.Add(1)
		go func(g int) {
			defer loggers.Done()
			for i := 0; i < 50; i++ {
				logger.Info("concurrent", Int("g", g), Int("i", i))
			}
		}(g)
	}
	loggers.Wait()
	close(stop)
	syncer.Wait()
}

func TestLoggerSync_RealFile(t *testing.T) {
	w, path := newTestRotatingWriter(t)
	logger := newOutputLogger(t, w)

	logger.Info("durable")
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() = %v, want nil", err)
	}
	if !fileContains(t, path, "durable") {
		t.Error("file does not contain \"durable\"")
	}
}

func TestIsIgnorableSyncErr(t *testing.T) {
	ebadf := &os.PathError{Op: "sync", Path: "/dev/stdout", Err: syscall.EBADF}
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"bare errno", syscall.ENOTTY, true},
		{"wrapped with %w like RotatingFileWriter.Sync", fmt.Errorf("file sync failed: %w", ebadf), true},
		{"other error", errors.New("disk full"), false},
		{"joined, all ignorable", errors.Join(ebadf, syscall.EINVAL), true},
		{"joined, one real", errors.Join(ebadf, errors.New("disk full")), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIgnorableSyncErr(tt.err); got != tt.want {
				t.Errorf("isIgnorableSyncErr(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
