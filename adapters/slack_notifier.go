// adapters/slack_notifier.go

package adapters

import (
	"errors"
	"github.com/slack-go/slack"
	"log"
	"os"
)

var (
	ErrSlackTokenMissing   = errors.New("SLACK_TOKEN environment variable is empty or not set")
	ErrSlackChannelMissing = errors.New("SLACK_CHANNEL_ID environment variable is empty or not set")
)

type SlackNotifier struct {
	Client      *slack.Client
	channelID   string // Issue #10: Cache channel ID to avoid repeated os.Getenv() calls
	enabled     bool   // Issue #7: Track if notifier is properly configured
}

// NewSlackNotifier crea una nueva instancia de SlackNotifier.
// Issue #7 Fix: Returns error if credentials are missing or invalid
// Issue #8 Fix: Support both SLACK_CHANNEL_ID (correct) and SLACK_CHANEL_ID (typo) for backward compatibility
// Issue #10 Fix: Cache channelID to avoid repeated os.Getenv() calls
func NewSlackNotifier() (*SlackNotifier, error) {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Printf("Warning: SLACK_TOKEN is not set, Slack notifications disabled")
		return &SlackNotifier{Client: nil, channelID: "", enabled: false}, ErrSlackTokenMissing
	}

	// Issue #8 Fix: Check for correct variable name first, then fall back to typo
	// Issue #10 Fix: Cache channelID to avoid repeated os.Getenv() calls
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	if channelID == "" {
		// Fall back to old typo version for backward compatibility
		channelID = os.Getenv("SLACK_CHANEL_ID")
		if channelID != "" {
			log.Printf("Warning: Using deprecated env var SLACK_CHANEL_ID. Please use SLACK_CHANNEL_ID instead")
		}
	}

	if channelID == "" {
		log.Printf("Warning: SLACK_CHANNEL_ID is not set, Slack notifications disabled")
		return &SlackNotifier{Client: nil, channelID: "", enabled: false}, ErrSlackChannelMissing
	}

	client := slack.New(token)
	return &SlackNotifier{
		Client:    client,
		channelID: channelID, // Issue #10: Cache channel ID
		enabled:   true,
	}, nil
}

// SendNotification envía un mensaje a Slack.
// Issue #7 Fix: Returns early if notifier is disabled due to missing credentials
// Issue #10 Fix: Use cached channelID instead of repeated os.Getenv() calls
func (s *SlackNotifier) SendNotification(message string) error {
	if !s.enabled {
		// Notifier is disabled due to missing credentials
		return nil
	}

	// Issue #10: Use cached channelID instead of os.Getenv()
	_, _, err := s.Client.PostMessage(s.channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Printf("Failed to send notification to Slack: %v", err)
		return err
	}
	log.Printf("Notification sent to Slack channel %s", s.channelID)
	return nil
}

// SendNotificationWithAttachments envía un mensaje a Slack con uno o más attachments.
// Issue #10 Fix: Use cached channelID instead of repeated os.Getenv() calls
func (s *SlackNotifier) SendNotificationWithAttachments(attachments []slack.Attachment) error {
	if !s.enabled {
		// Notifier is disabled due to missing credentials
		return nil
	}

	// Issue #10: Use cached channelID instead of os.Getenv()
	_, _, err := s.Client.PostMessage(s.channelID, slack.MsgOptionAttachments(attachments...))
	if err != nil {
		log.Printf("Failed to send notification with attachments to Slack: %v", err)
		return err
	}
	log.Printf("Notification with attachments sent to Slack channel %s", s.channelID)
	return nil
}

// IsEnabled returns whether the notifier is enabled (for testing)
// Issue #7: Public accessor for enabled field
func (s *SlackNotifier) IsEnabled() bool {
	return s.enabled
}
