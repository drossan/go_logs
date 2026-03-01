// Package signal provides signal handling for log rotation.
//
// This package enables automatic log rotation when the process receives a signal
// (typically SIGHUP from logrotate). This is essential for production systems
// where log rotation is managed externally.
//
// Features:
//   - SIGHUP handler for log rotation
//   - Configurable signals
//   - Thread-safe operation
//   - Graceful shutdown
//
// Example:
//
//	writer, _ := go_logs.NewRotatingFileWriter("/var/log/app.log", 100, 5)
//	logger, _ := go_logs.New(go_logs.WithOutput(writer))
//
//	handler := signal.NewSIGHUPHandler(writer)
//	handler.Register()
//	defer handler.Stop()
//
//	// Logs will rotate when SIGHUP is received
//	logger.Info("Server started")
package signal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	go_logs "github.com/drossan/go_logs"
)

// Rotator is an interface for types that can rotate their output.
type Rotator interface {
	Rotate() error
}

// SIGHUPHandler handles SIGHUP signals to trigger log rotation.
type SIGHUPHandler struct {
	rotator   Rotator
	signals   []os.Signal
	sigChan   chan os.Signal
	done      chan struct{}
	logger    go_logs.Logger
}

// SIGHUPHandlerOption is a functional option for configuring the handler.
type SIGHUPHandlerOption func(*SIGHUPHandler)

// WithLogger sets a custom logger for the handler.
func WithLogger(logger go_logs.Logger) SIGHUPHandlerOption {
	return func(h *SIGHUPHandler) {
		h.logger = logger
	}
}

// NewSIGHUPHandler creates a new SIGHUP handler for the given rotator.
// Additional signals can be specified for rotation triggers.
func NewSIGHUPHandler(rotator Rotator, additionalSignals ...os.Signal) *SIGHUPHandler {
	signals := []os.Signal{syscall.SIGHUP}
	signals = append(signals, additionalSignals...)

	return &SIGHUPHandler{
		rotator: rotator,
		signals: signals,
		sigChan: make(chan os.Signal, 1),
		done:    make(chan struct{}),
	}
}

// NewSIGHUPHandlerWithChannel creates a handler with a custom signal channel.
// This is useful for testing or when you want to control signal delivery.
func NewSIGHUPHandlerWithChannel(rotator Rotator, sigChan chan os.Signal) *SIGHUPHandler {
	return &SIGHUPHandler{
		rotator: rotator,
		signals: []os.Signal{syscall.SIGHUP},
		sigChan: sigChan,
		done:    make(chan struct{}),
	}
}

// Register starts listening for signals.
// This method is non-blocking and starts a goroutine to handle signals.
func (h *SIGHUPHandler) Register() {
	signal.Notify(h.sigChan, h.signals...)

	go h.handleSignals()
}

// handleSignals processes incoming signals.
func (h *SIGHUPHandler) handleSignals() {
	for {
		select {
		case sig := <-h.sigChan:
			h.logInfo("Received signal %v, rotating logs", sig)
			if err := h.rotator.Rotate(); err != nil {
				h.logError("Failed to rotate logs: %v", err)
			} else {
				h.logInfo("Log rotation completed")
			}
		case <-h.done:
			return
		}
	}
}

// Stop stops the signal handler.
// It unregisters the signal notification and stops the handler goroutine.
func (h *SIGHUPHandler) Stop() {
	signal.Stop(h.sigChan)
	close(h.done)
}

// logInfo logs an info message if a logger is configured.
func (h *SIGHUPHandler) logInfo(msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger.Info(fmt.Sprintf(msg, args...))
	}
}

// logError logs an error message if a logger is configured.
func (h *SIGHUPHandler) logError(msg string, args ...interface{}) {
	if h.logger != nil {
		h.logger.Error(fmt.Sprintf(msg, args...))
	}
}

// RotateWrapper wraps a RotatingFileWriter to implement Rotator.
type RotateWrapper struct {
	writer *go_logs.RotatingFileWriter
}

// WrapRotator creates a Rotator from a RotatingFileWriter.
func WrapRotator(writer *go_logs.RotatingFileWriter) Rotator {
	return &RotateWrapper{writer: writer}
}

// Rotate implements Rotator.
func (w *RotateWrapper) Rotate() error {
	return w.writer.Rotate()
}
