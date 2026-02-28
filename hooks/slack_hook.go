package hooks

import (
	"fmt"

	"github.com/drossan/go_logs"
)

// SlackNotifier is an interface for sending notifications to Slack.
// This allows mocking in tests and using the real adapter in production.
type SlackNotifier interface {
	SendNotification(message string) error
}

// SlackHook sends log entries to Slack based on level threshold.
//
// This hook integrates with the existing SlackNotifier adapter and
// only sends notifications for log entries at or above the configured level.
// This prevents spamming Slack with info/debug messages.
//
// Example:
//
//	notifier, _ := adapters.NewSlackNotifier()
//	hook := hooks.NewSlackHook(notifier, go_logs.ErrorLevel)
//	logger, _ := go_logs.New(go_logs.WithHooks(hook))
type SlackHook struct {
	notifier SlackNotifier
	level    go_logs.Level // Minimum level to send to Slack
}

// NewSlackHook creates a new Slack hook with the given notifier and level threshold.
//
// Parameters:
//
//	notifier - The SlackNotifier to use for sending messages
//	level    - The minimum log level to send to Slack (e.g., ErrorLevel)
//
// Returns:
//
//	A configured Hook that sends logs to Slack
//
// Example:
//
//	notifier, _ := adapters.NewSlackNotifier()
//	hook := hooks.NewSlackHook(notifier, go_logs.ErrorLevel)
//
//	// Only Error and Fatal logs will be sent to Slack
//	logger.Error("Database connection failed", go_logs.Err(err))
func NewSlackHook(notifier SlackNotifier, level go_logs.Level) *SlackHook {
	return &SlackHook{
		notifier: notifier,
		level:    level,
	}
}

// Run implements Hook.Run by sending the log entry to Slack if it meets the level threshold.
//
// Only logs at or above the configured level are sent to Slack. If the notifier
// is disabled (e.g., missing credentials), this method returns nil without error.
//
// Parameters:
//
//	entry - The log entry to potentially send to Slack
//
// Returns:
//
//	error - An error if sending to Slack failed, nil otherwise
//
// Example:
//
//	func (h *SlackHook) Run(entry *go_logs.Entry) error {
//	    if entry.Level >= h.level {
//	        message := h.formatMessage(entry)
//	        return h.notifier.SendNotification(message)
//	    }
//	    return nil
//	}
func (h *SlackHook) Run(entry *go_logs.Entry) error {
	// Fast-path: Check if we should send this level to Slack
	if !entry.Level.ShouldLog(h.level) {
		return nil
	}

	// Format and send the message
	message := h.formatMessage(entry)
	return h.notifier.SendNotification(message)
}

// formatMessage formats a log entry as a Slack message.
//
// The format includes the level and message, with fields appended as key=value pairs.
//
// Parameters:
//
//	entry - The log entry to format
//
// Returns:
//
//	A formatted string suitable for sending to Slack
func (h *SlackHook) formatMessage(entry *go_logs.Entry) string {
	// Start with level and message
	msg := fmt.Sprintf("[%s] %s", entry.Level.String(), entry.Message)

	// Append fields if present
	if len(entry.Fields) > 0 {
		msg += " "
		for i, field := range entry.Fields {
			if i > 0 {
				msg += " "
			}
			msg += fmt.Sprintf("%s=%v", field.Key(), field.Value())
		}
	}

	return msg
}

// GetLevel returns the minimum log level that will be sent to Slack.
//
// Example:
//
//	if hook.GetLevel() == go_logs.ErrorLevel {
//	    // Only errors and fatals are sent to Slack
//	}
func (h *SlackHook) GetLevel() go_logs.Level {
	return h.level
}

// SetLevel changes the minimum log level threshold.
//
// This can be used to dynamically adjust what gets sent to Slack.
//
// Parameters:
//
//	level - The new minimum level to send to Slack
//
// Example:
//
//	hook.SetLevel(go_logs.WarnLevel) // Send warnings, errors, and fatals
func (h *SlackHook) SetLevel(level go_logs.Level) {
	h.level = level
}
