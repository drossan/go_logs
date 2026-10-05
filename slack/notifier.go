// Package slack sends go_logs notifications to a Slack channel.
//
// It lives in its own module (github.com/drossan/go_logs/slack/v3) so that the
// go_logs core does not pull the Slack client and its dependencies into every
// consumer. A *Notifier satisfies both domain.Notifier (for the v2 API through
// go_logs.SetNotifier) and hooks.SlackNotifier (for the v3 hooks.SlackHook):
//
//	n, err := slack.NewNotifierFromEnv()
//	if err != nil {
//	    log.Printf("Slack disabled: %v", err)
//	}
//	go_logs.SetNotifier(n) // v2: ErrorLog, FatalLog… with NOTIFICATIONS_SLACK_ENABLED=1
//
//	hook := hooks.NewSlackHook(n, go_logs.ErrorLevel) // v3
package slack

import (
	"errors"
	"fmt"
	"log"
	"os"

	slackapi "github.com/slack-go/slack"
)

var (
	// ErrTokenMissing is returned when the Slack bot token is empty.
	ErrTokenMissing = errors.New("slack: token is empty")
	// ErrChannelMissing is returned when the Slack channel ID is empty.
	ErrChannelMissing = errors.New("slack: channel ID is empty")
)

// Notifier posts messages to a single Slack channel.
//
// A Notifier returned together with an error is disabled: its Send methods
// return nil without contacting Slack, so it is safe to register it anyway.
// Methods are also safe on a nil *Notifier.
type Notifier struct {
	// Client is the underlying Slack API client (nil when disabled).
	Client    *slackapi.Client
	channelID string
	enabled   bool
}

// NewNotifier builds a Notifier from an explicit bot token and channel ID,
// without reading the environment.
//
// If token or channelID is empty it returns a disabled, non-nil Notifier and an
// error that wraps ErrTokenMissing or ErrChannelMissing.
func NewNotifier(token, channelID string) (*Notifier, error) {
	if token == "" {
		return &Notifier{}, ErrTokenMissing
	}
	if channelID == "" {
		return &Notifier{}, ErrChannelMissing
	}
	return &Notifier{
		Client:    slackapi.New(token),
		channelID: channelID,
		enabled:   true,
	}, nil
}

// NewNotifierFromEnv builds a Notifier from SLACK_TOKEN and SLACK_CHANNEL_ID.
//
// The deprecated SLACK_CHANEL_ID is used (with a warning through the standard
// logger) only when SLACK_CHANNEL_ID is empty. The channel ID is read once and
// cached. If a variable is missing it returns a disabled, non-nil Notifier and
// an error that wraps ErrTokenMissing or ErrChannelMissing and names the
// variable; the token value never appears in errors or warnings.
func NewNotifierFromEnv() (*Notifier, error) {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return &Notifier{}, fmt.Errorf("%w: SLACK_TOKEN is empty or not set", ErrTokenMissing)
	}

	channelID := os.Getenv("SLACK_CHANNEL_ID")
	if channelID == "" {
		channelID = os.Getenv("SLACK_CHANEL_ID")
		if channelID != "" {
			log.Printf("Warning: Using deprecated env var SLACK_CHANEL_ID. Please use SLACK_CHANNEL_ID instead")
		}
	}
	if channelID == "" {
		return &Notifier{}, fmt.Errorf("%w: SLACK_CHANNEL_ID is empty or not set", ErrChannelMissing)
	}

	return NewNotifier(token, channelID)
}

// SendNotification posts message as plain text to the configured channel.
// It returns nil without doing anything if the notifier is disabled or nil,
// and the Slack API error otherwise.
func (n *Notifier) SendNotification(message string) error {
	if !n.IsEnabled() {
		return nil
	}
	if _, _, err := n.Client.PostMessage(n.channelID, slackapi.MsgOptionText(message, false)); err != nil {
		log.Printf("Failed to send notification to Slack: %v", err)
		return err
	}
	log.Printf("Notification sent to Slack via channel %s", n.channelID)
	return nil
}

// SendNotificationWithAttachments posts one or more attachments to the
// configured channel. It returns nil without doing anything if the notifier is
// disabled or nil, and the Slack API error otherwise.
func (n *Notifier) SendNotificationWithAttachments(attachments []slackapi.Attachment) error {
	if !n.IsEnabled() {
		return nil
	}
	if _, _, err := n.Client.PostMessage(n.channelID, slackapi.MsgOptionAttachments(attachments...)); err != nil {
		log.Printf("Failed to send notification with attachments to Slack: %v", err)
		return err
	}
	log.Printf("Notification with attachments sent to Slack channel %s", n.channelID)
	return nil
}

// IsEnabled reports whether the notifier has credentials and will contact
// Slack. It is false for a nil *Notifier.
func (n *Notifier) IsEnabled() bool {
	return n != nil && n.enabled
}
