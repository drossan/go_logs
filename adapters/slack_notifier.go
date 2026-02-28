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
	enabled     bool // Issue #7: Track if notifier is properly configured
}

// NewSlackNotifier crea una nueva instancia de SlackNotifier.
// Issue #7 Fix: Returns error if credentials are missing or invalid
// Issue #8 Fix: Support both SLACK_CHANNEL_ID (correct) and SLACK_CHANEL_ID (typo) for backward compatibility
func NewSlackNotifier() (*SlackNotifier, error) {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		log.Printf("Warning: SLACK_TOKEN is not set, Slack notifications disabled")
		return &SlackNotifier{Client: nil, enabled: false}, ErrSlackTokenMissing
	}

	// Issue #8 Fix: Check for correct variable name first, then fall back to typo
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
		return &SlackNotifier{Client: nil, enabled: false}, ErrSlackChannelMissing
	}

	client := slack.New(token)
	return &SlackNotifier{
		Client:  client,
		enabled: true,
	}, nil
}

// SendNotification envía un mensaje a Slack.
// Issue #7 Fix: Returns early if notifier is disabled due to missing credentials
// Issue #8 Fix: Support both SLACK_CHANNEL_ID (correct) and SLACK_CHANEL_ID (typo)
func (s *SlackNotifier) SendNotification(message string) error {
	if !s.enabled {
		// Notifier is disabled due to missing credentials
		return nil
	}

	// Issue #8 Fix: Check for correct variable name first, then fall back to typo
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	if channelID == "" {
		channelID = os.Getenv("SLACK_CHANEL_ID")
	}

	_, _, err := s.Client.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Printf("Failed to send notification to Slack: %v", err)
		return err
	}
	log.Printf("Notification sent to Slack channel %s", channelID)
	return nil
}

// SendNotificationWithAttachments envía un mensaje a Slack con uno o más attachments.
// Issue #8 Fix: Support both SLACK_CHANNEL_ID (correct) and SLACK_CHANEL_ID (typo)
func (s *SlackNotifier) SendNotificationWithAttachments(attachments []slack.Attachment) error {
	if !s.enabled {
		// Notifier is disabled due to missing credentials
		return nil
	}

	// Issue #8 Fix: Check for correct variable name first, then fall back to typo
	channelID := os.Getenv("SLACK_CHANNEL_ID")
	if channelID == "" {
		channelID = os.Getenv("SLACK_CHANEL_ID")
	}

	_, _, err := s.Client.PostMessage(channelID, slack.MsgOptionAttachments(attachments...))
	if err != nil {
		log.Printf("Failed to send notification with attachments to Slack: %v", err)
		return err
	}
	log.Printf("Notification with attachments sent to Slack channel %s", channelID)
	return nil
}
