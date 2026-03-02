package signal

import (
	"os"
	"syscall"
	"testing"
	"time"

	go_logs "github.com/drossan/go_logs"
)

// Ensure os.Signal is used
var _ = os.Signal(nil)

// TestSIGHUPHandler_Register tests registering a SIGHUP handler
func TestSIGHUPHandler_Register(t *testing.T) {
	// Create a mock rotator
	rotated := make(chan struct{}, 1)
	rotator := &mockRotator{
		rotateFunc: func() error {
			select {
			case rotated <- struct{}{}:
			default:
			}
			return nil
		},
	}

	handler := NewSIGHUPHandler(rotator)
	defer handler.Stop()

	// Register the handler
	handler.Register()

	// Send SIGHUP
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	// Wait for rotation
	select {
	case <-rotated:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Expected rotation to be triggered")
	}
}

// TestSIGHUPHandler_Stop tests stopping the handler
func TestSIGHUPHandler_Stop(t *testing.T) {
	rotator := &mockRotator{}
	handler := NewSIGHUPHandler(rotator)

	handler.Register()

	// Stop should not block
	done := make(chan struct{})
	go func() {
		handler.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Stop should not block")
	}
}

// TestSIGHUPHandler_CustomSignals tests custom signals
func TestSIGHUPHandler_CustomSignals(t *testing.T) {
	rotated := make(chan struct{}, 1)
	rotator := &mockRotator{
		rotateFunc: func() error {
			select {
			case rotated <- struct{}{}:
			default:
			}
			return nil
		},
	}

	handler := NewSIGHUPHandler(rotator, syscall.SIGUSR1)
	defer handler.Stop()

	handler.Register()

	// Send SIGUSR1
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)

	select {
	case <-rotated:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Expected rotation on SIGUSR1")
	}
}

// TestSIGHUPHandler_MultipleRotations tests multiple signal triggers
func TestSIGHUPHandler_MultipleRotations(t *testing.T) {
	count := 0
	rotator := &mockRotator{
		rotateFunc: func() error {
			count++
			return nil
		},
	}

	handler := NewSIGHUPHandler(rotator)
	defer handler.Stop()

	handler.Register()

	// Send multiple signals
	for i := 0; i < 3; i++ {
		syscall.Kill(syscall.Getpid(), syscall.SIGHUP)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if count < 3 {
		t.Errorf("Expected at least 3 rotations, got %d", count)
	}
}

// TestSIGHUPHandler_WithError tests error handling in rotation
func TestSIGHUPHandler_WithError(t *testing.T) {
	rotator := &mockRotator{
		rotateFunc: func() error {
			return os.ErrClosed
		},
	}

	handler := NewSIGHUPHandler(rotator)
	defer handler.Stop()

	// Should not panic on error
	handler.Register()
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)
}

// TestSIGHUPHandler_WithLogger tests with a real logger
func TestSIGHUPHandler_WithLogger(t *testing.T) {
	// Create a logger with rotating file writer
	// Note: This test requires file system access
	tempFile := "/tmp/go_logs_signal_test.log"
	defer os.Remove(tempFile)
	defer os.Remove(tempFile + ".1")

	writer, err := go_logs.NewRotatingFileWriter(tempFile, 1, 5)
	if err != nil {
		t.Fatalf("Failed to create rotating file writer: %v", err)
	}

	logger, _ := go_logs.New(
		go_logs.WithLevel(go_logs.InfoLevel),
		go_logs.WithOutput(writer),
	)

	// Wrap the writer as a Rotator
	rotator := WrapRotator(writer)

	handler := NewSIGHUPHandler(rotator)
	defer handler.Stop()

	handler.Register()

	// Log something
	logger.Info("test message")

	// Send SIGHUP
	syscall.Kill(syscall.Getpid(), syscall.SIGHUP)

	// Wait for rotation
	time.Sleep(100 * time.Millisecond)

	// Verify file still exists
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Log file should still exist after rotation")
	}
}

// TestSIGHUPHandler_NotifyChannel tests using a custom notify channel
func TestSIGHUPHandler_NotifyChannel(t *testing.T) {
	rotated := make(chan struct{}, 1)
	rotator := &mockRotator{
		rotateFunc: func() error {
			select {
			case rotated <- struct{}{}:
			default:
			}
			return nil
		},
	}

	// Create custom notify channel
	sigChan := make(chan os.Signal, 1)
	handler := NewSIGHUPHandlerWithChannel(rotator, sigChan)
	defer handler.Stop()

	handler.Register()

	// Send signal via channel
	sigChan <- syscall.SIGHUP

	select {
	case <-rotated:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Expected rotation on channel signal")
	}
}

// TestWrapRotator tests wrapping a RotatingFileWriter
func TestWrapRotator(t *testing.T) {
	tempFile := "/tmp/go_logs_wrap_rotator_test.log"
	defer os.Remove(tempFile)

	writer, err := go_logs.NewRotatingFileWriter(tempFile, 1, 5)
	if err != nil {
		t.Fatalf("Failed to create rotating file writer: %v", err)
	}

	rotator := WrapRotator(writer)

	// Should not error
	if err := rotator.Rotate(); err != nil {
		t.Errorf("Rotate should not error: %v", err)
	}
}

// mockRotator is a mock implementation of Rotator for testing
type mockRotator struct {
	rotateFunc func() error
}

func (m *mockRotator) Rotate() error {
	if m.rotateFunc != nil {
		return m.rotateFunc()
	}
	return nil
}
