// adapters/slack_notifier.go

package adapters

import (
	"github.com/slack-go/slack"
	"log"
	"os"
)

type SlackNotifier struct {
	Client *slack.Client
}

// NewSlackNotifier crea una nueva instancia de SlackNotifier.
func NewSlackNotifier() *SlackNotifier {
	token := os.Getenv("SLACK_TOKEN")
	client := slack.New(token)
	return &SlackNotifier{
		Client: client,
	}
}

// SendNotification envía un mensaje a Slack.
func (s *SlackNotifier) SendNotification(message string) error {
	channelID := os.Getenv("SLACK_CHANEL_ID")
	_, _, err := s.Client.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Printf("Failed to send notification to Slack: %v", err)
		return err
	}
	log.Printf("Notification sent to Slack channel %s", channelID)
	return nil
}

// SendNotificationWithAttachments envía un mensaje a Slack con uno o más attachments.
func (s *SlackNotifier) SendNotificationWithAttachments(attachments []slack.Attachment) error {
	channelID := os.Getenv("SLACK_CHANEL_ID")
	_, _, err := s.Client.PostMessage(channelID, slack.MsgOptionAttachments(attachments...))
	if err != nil {
		log.Printf("Failed to send notification with attachments to Slack: %v", err)
		return err
	}
	log.Printf("Notification with attachments sent to Slack channel %s", channelID)
	return nil
}
