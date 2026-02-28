package hooks

import (
	"errors"
	"strings"
	"testing"

	"github.com/drossan/go_logs"
)

// mockSlackNotifier is a mock implementation of SlackNotifier for testing
type mockSlackNotifier struct {
	called     bool
	calledWith string
	sendError  error
}

func (m *mockSlackNotifier) SendNotification(message string) error {
	m.called = true
	m.calledWith = message
	return m.sendError
}

// TestNewSlackHook tests creating a new SlackHook
func TestNewSlackHook(t *testing.T) {
	notifier := &mockSlackNotifier{}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	if hook == nil {
		t.Fatal("NewSlackHook returned nil")
	}

	if hook.GetLevel() != go_logs.ErrorLevel {
		t.Errorf("NewSlackHook level = %v, want ErrorLevel", hook.GetLevel())
	}
}

// TestSlackHook_Run_WithMatchingLevel tests that hook runs when level matches threshold
func TestSlackHook_Run_WithMatchingLevel(t *testing.T) {
	notifier := &mockSlackNotifier{}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	entry := &go_logs.Entry{
		Level:   go_logs.ErrorLevel,
		Message: "Test error",
	}

	err := hook.Run(entry)

	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	if !notifier.called {
		t.Error("Run() did not call notifier")
	}

	if !strings.Contains(notifier.calledWith, "Test error") {
		t.Errorf("Run() message = %s, want to contain 'Test error'", notifier.calledWith)
	}

	if !strings.Contains(notifier.calledWith, "ERROR") {
		t.Errorf("Run() message = %s, want to contain 'ERROR'", notifier.calledWith)
	}
}

// TestSlackHook_Run_WithHigherLevel tests that hook runs when level is higher than threshold
func TestSlackHook_Run_WithHigherLevel(t *testing.T) {
	notifier := &mockSlackNotifier{}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	entry := &go_logs.Entry{
		Level:   go_logs.FatalLevel,
		Message: "Fatal error",
	}

	err := hook.Run(entry)

	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	if !notifier.called {
		t.Error("Run() did not call notifier for Fatal level")
	}
}

// TestSlackHook_Run_WithLowerLevel tests that hook doesn't run when level is below threshold
func TestSlackHook_Run_WithLowerLevel(t *testing.T) {
	notifier := &mockSlackNotifier{}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	entry := &go_logs.Entry{
		Level:   go_logs.InfoLevel,
		Message: "Info message",
	}

	err := hook.Run(entry)

	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	if notifier.called {
		t.Error("Run() called notifier for Info level below Error threshold")
	}
}

// TestSlackHook_Run_WithFields tests that hook includes fields in message
func TestSlackHook_Run_WithFields(t *testing.T) {
	notifier := &mockSlackNotifier{}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	entry := &go_logs.Entry{
		Level:   go_logs.ErrorLevel,
		Message: "Database error",
		Fields: []go_logs.Field{
			go_logs.String("host", "localhost"),
			go_logs.Int("port", 5432),
		},
	}

	err := hook.Run(entry)

	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	if !strings.Contains(notifier.calledWith, "host=localhost") {
		t.Errorf("Run() message = %s, want to contain 'host=localhost'", notifier.calledWith)
	}

	if !strings.Contains(notifier.calledWith, "port=5432") {
		t.Errorf("Run() message = %s, want to contain 'port=5432'", notifier.calledWith)
	}
}

// TestSlackHook_Run_NotifierError tests that hook returns notifier errors
func TestSlackHook_Run_NotifierError(t *testing.T) {
	testError := errors.New("slack API error")
	notifier := &mockSlackNotifier{sendError: testError}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	entry := &go_logs.Entry{
		Level:   go_logs.ErrorLevel,
		Message: "Test error",
	}

	err := hook.Run(entry)

	if err != testError {
		t.Errorf("Run() error = %v, want %v", err, testError)
	}
}

// TestSlackHook_SetLevel tests changing the level threshold
func TestSlackHook_SetLevel(t *testing.T) {
	notifier := &mockSlackNotifier{}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	// Initially should not send Info
	hook.SetLevel(go_logs.WarnLevel)

	if hook.GetLevel() != go_logs.WarnLevel {
		t.Errorf("SetLevel() level = %v, want WarnLevel", hook.GetLevel())
	}

	// Now Info should not be sent (still below Warn)
	entry := &go_logs.Entry{
		Level:   go_logs.InfoLevel,
		Message: "Info",
	}

	hook.Run(entry)

	if notifier.called {
		t.Error("Run() called notifier for Info level below Warn threshold")
	}

	// But Warn should be sent
	entry.Level = go_logs.WarnLevel
	hook.Run(entry)

	if !notifier.called {
		t.Error("Run() did not call notifier for Warn level")
	}
}

// TestSlackHook_formatMessage tests message formatting
func TestSlackHook_formatMessage(t *testing.T) {
	notifier := &mockSlackNotifier{}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	entry := &go_logs.Entry{
		Level:   go_logs.ErrorLevel,
		Message: "Test message",
	}

	message := hook.formatMessage(entry)

	// Check basic format
	if !strings.Contains(message, "[ERROR]") {
		t.Errorf("formatMessage() = %s, want to contain '[ERROR]'", message)
	}

	if !strings.Contains(message, "Test message") {
		t.Errorf("formatMessage() = %s, want to contain 'Test message'", message)
	}
}

// TestSlackHook_formatMessageWithFields tests formatting with fields
func TestSlackHook_formatMessageWithFields(t *testing.T) {
	notifier := &mockSlackNotifier{}
	hook := NewSlackHook(notifier, go_logs.ErrorLevel)

	entry := &go_logs.Entry{
		Level:   go_logs.ErrorLevel,
		Message: "Test message",
		Fields: []go_logs.Field{
			go_logs.String("key1", "value1"),
			go_logs.Int("key2", 42),
		},
	}

	message := hook.formatMessage(entry)

	if !strings.Contains(message, "key1=value1") {
		t.Errorf("formatMessage() = %s, want to contain 'key1=value1'", message)
	}

	if !strings.Contains(message, "key2=42") {
		t.Errorf("formatMessage() = %s, want to contain 'key2=42'", message)
	}
}
