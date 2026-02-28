package adapters

import (
	"os"
	"testing"

	"github.com/slack-go/slack"
)

// TestNewSlackNotifierValid verifies that NewSlackNotifier creates
// a valid notifier when both token and channel ID are provided
func TestNewSlackNotifierValid(t *testing.T) {
	// Setup: Set valid environment variables
	os.Setenv("SLACK_TOKEN", "xoxb-test-token")
	os.Setenv("SLACK_CHANEL_ID", "C1234567890")
	defer func() {
		os.Unsetenv("SLACK_TOKEN")
		os.Unsetenv("SLACK_CHANEL_ID")
	}()

	// Call constructor
	notifier, err := NewSlackNotifier()

	// Verify no error
	if err != nil {
		t.Errorf("Expected no error with valid credentials, got: %v", err)
	}

	// Verify notifier was created
	if notifier == nil {
		t.Error("Expected notifier to be created, got nil")
	}

	// Verify notifier is enabled
	if !notifier.enabled {
		t.Error("Expected notifier to be enabled with valid credentials")
	}

	// Verify client was initialized
	if notifier.Client == nil {
		t.Error("Expected client to be initialized, got nil")
	}
}

// TestNewSlackNotifierMissingToken verifies that NewSlackNotifier returns
// an error when SLACK_TOKEN is missing or empty
// Issue #7: Sin validación de credenciales de Slack
func TestNewSlackNotifierMissingToken(t *testing.T) {
	// Setup: Set channel but NOT token
	os.Setenv("SLACK_CHANEL_ID", "C1234567890")
	os.Unsetenv("SLACK_TOKEN")
	defer func() {
		os.Unsetenv("SLACK_CHANEL_ID")
	}()

	// Call constructor - should return error
	notifier, err := NewSlackNotifier()

	// Verify error returned
	if err == nil {
		t.Error("Expected error when SLACK_TOKEN is missing, got nil")
	}

	if err != ErrSlackTokenMissing {
		t.Errorf("Expected ErrSlackTokenMissing, got: %v", err)
	}

	// Verify notifier was created but disabled
	if notifier == nil {
		t.Error("Expected notifier to be created (but disabled)")
	}

	if notifier.enabled {
		t.Error("Expected notifier to be disabled when token is missing")
	}
}

// TestNewSlackNotifierMissingChannel verifies that NewSlackNotifier returns
// an error when SLACK_CHANEL_ID is missing or empty
func TestNewSlackNotifierMissingChannel(t *testing.T) {
	// Setup: Set token but NOT channel
	os.Setenv("SLACK_TOKEN", "xoxb-test-token")
	os.Unsetenv("SLACK_CHANEL_ID")
	defer func() {
		os.Unsetenv("SLACK_TOKEN")
	}()

	// Call constructor - should return error
	notifier, err := NewSlackNotifier()

	// Verify error returned
	if err == nil {
		t.Error("Expected error when SLACK_CHANNEL_ID is missing, got nil")
	}

	if err != ErrSlackChannelMissing {
		t.Errorf("Expected ErrSlackChannelMissing, got: %v", err)
	}

	// Verify notifier was created but disabled
	if notifier == nil {
		t.Error("Expected notifier to be created (but disabled)")
	}

	if notifier.enabled {
		t.Error("Expected notifier to be disabled when channel is missing")
	}
}

// TestNewSlackNotifierEmptyCredentials verifies behavior when both
// token and channel are empty strings
func TestNewSlackNotifierEmptyCredentials(t *testing.T) {
	// Setup: Set empty strings
	os.Setenv("SLACK_TOKEN", "")
	os.Setenv("SLACK_CHANEL_ID", "")
	defer func() {
		os.Unsetenv("SLACK_TOKEN")
		os.Unsetenv("SLACK_CHANEL_ID")
	}()

	// Call constructor
	notifier, err := NewSlackNotifier()

	// Verify error for missing token
	if err == nil {
		t.Error("Expected error when SLACK_TOKEN is empty, got nil")
	}

	// Verify notifier is disabled
	if notifier.enabled {
		t.Error("Expected notifier to be disabled when credentials are empty")
	}
}

// TestSlackNotifierSendNotificationWhenDisabled verifies that
// SendNotification returns early (no error) when notifier is disabled
func TestSlackNotifierSendNotificationWhenDisabled(t *testing.T) {
	// Create disabled notifier
	notifier := &SlackNotifier{
		Client:  nil,
		enabled: false,
	}

	// Call SendNotification - should return nil without attempting to send
	err := notifier.SendNotification("Test message")

	if err != nil {
		t.Errorf("Expected no error when notifier is disabled, got: %v", err)
	}
}

// TestSlackChannelNameTypo verifies backward compatibility with typo
// Issue #8: Support both SLACK_CHANNEL_ID (correct) and SLACK_CHANEL_ID (typo)
func TestSlackChannelNameTypo(t *testing.T) {
	// Test with typo version (should work with warning)
	t.Run("Typo version works with warning", func(t *testing.T) {
		os.Setenv("SLACK_TOKEN", "xoxb-test-token")
		os.Setenv("SLACK_CHANEL_ID", "C1234567890") // Typo version
		os.Unsetenv("SLACK_CHANNEL_ID") // Ensure correct version is NOT set
		defer func() {
			os.Unsetenv("SLACK_TOKEN")
			os.Unsetenv("SLACK_CHANEL_ID")
			os.Unsetenv("SLACK_CHANNEL_ID")
		}()

		notifier, err := NewSlackNotifier()

		if err != nil {
			t.Errorf("Expected no error with typo version, got: %v", err)
		}

		if !notifier.enabled {
			t.Error("Expected notifier to be enabled with typo version")
		}
	})

	// Test with correct version (should work without warning)
	t.Run("Correct version works without warning", func(t *testing.T) {
		os.Setenv("SLACK_TOKEN", "xoxb-test-token")
		os.Setenv("SLACK_CHANNEL_ID", "C1234567890") // Correct version
		os.Unsetenv("SLACK_CHANEL_ID") // Ensure typo version is NOT set
		defer func() {
			os.Unsetenv("SLACK_TOKEN")
			os.Unsetenv("SLACK_CHANNEL_ID")
			os.Unsetenv("SLACK_CHANEL_ID")
		}()

		notifier, err := NewSlackNotifier()

		if err != nil {
			t.Errorf("Expected no error with correct version, got: %v", err)
		}

		if !notifier.enabled {
			t.Error("Expected notifier to be enabled with correct version")
		}
	})
}

// TestSlackNotifierSendNotificationWithAttachments verifies SendNotificationWithAttachments
// Issue #8: Test backward compatibility with both env var names
func TestSlackNotifierSendNotificationWithAttachments(t *testing.T) {
	t.Run("With valid credentials", func(t *testing.T) {
		// Setup: Valid credentials
		os.Setenv("SLACK_TOKEN", "xoxb-test-token")
		os.Setenv("SLACK_CHANNEL_ID", "C1234567890")
		defer func() {
			os.Unsetenv("SLACK_TOKEN")
			os.Unsetenv("SLACK_CHANNEL_ID")
		}()

		// Create notifier
		notifier, err := NewSlackNotifier()
		if err != nil {
			t.Fatalf("Expected no error with valid credentials, got: %v", err)
		}

		if !notifier.enabled {
			t.Fatal("Expected notifier to be enabled")
		}

		// Create test attachments
		attachments := []slack.Attachment{
			{
				Title:      "Test Title",
				Text:       "Test Text",
				Color:      "#36a64f",
				AuthorName: "Test Author",
			},
		}

		// Call function under test
		// Note: This will fail to actually send (invalid token), but we can test the logic
		err = notifier.SendNotificationWithAttachments(attachments)

		// We expect an error (invalid token), but function should execute without panic
		if err == nil {
			// If no error, the API call succeeded (unexpected but acceptable)
			t.Log("Unexpected: API call succeeded with test token")
		}
	})

	t.Run("When disabled returns nil", func(t *testing.T) {
		// Create disabled notifier
		notifier := &SlackNotifier{
			Client:  nil,
			enabled: false,
		}

		// Create test attachments
		attachments := []slack.Attachment{
			{Title: "Test", Text: "Test"},
		}

		// Call function under test
		err := notifier.SendNotificationWithAttachments(attachments)

		// Should return nil without attempting to send
		if err != nil {
			t.Errorf("Expected no error when notifier is disabled, got: %v", err)
		}
	})

	t.Run("With empty attachments", func(t *testing.T) {
		// Setup: Valid credentials
		os.Setenv("SLACK_TOKEN", "xoxb-test-token")
		os.Setenv("SLACK_CHANNEL_ID", "C1234567890")
		defer func() {
			os.Unsetenv("SLACK_TOKEN")
			os.Unsetenv("SLACK_CHANNEL_ID")
		}()

		// Create notifier
		notifier, err := NewSlackNotifier()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Call with empty attachments
		emptyAttachments := []slack.Attachment{}
		err = notifier.SendNotificationWithAttachments(emptyAttachments)

		// Should not panic, may fail on API call
		// We're just testing it doesn't panic with empty slice
		_ = err
	})

	t.Run("With multiple attachments", func(t *testing.T) {
		// Setup: Valid credentials
		os.Setenv("SLACK_TOKEN", "xoxb-test-token")
		os.Setenv("SLACK_CHANNEL_ID", "C1234567890")
		defer func() {
			os.Unsetenv("SLACK_TOKEN")
			os.Unsetenv("SLACK_CHANNEL_ID")
		}()

		// Create notifier
		notifier, err := NewSlackNotifier()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Create multiple attachments
		attachments := []slack.Attachment{
			{Title: "Attachment 1", Text: "Text 1", Color: "good"},
			{Title: "Attachment 2", Text: "Text 2", Color: "warning"},
			{Title: "Attachment 3", Text: "Text 3", Color: "danger"},
		}

		// Call function under test
		err = notifier.SendNotificationWithAttachments(attachments)

		// Should not panic with multiple attachments
		_ = err
	})

	t.Run("Uses cached channel ID", func(t *testing.T) {
		// Setup: Valid credentials with correct env var
		os.Setenv("SLACK_TOKEN", "xoxb-test-token")
		os.Setenv("SLACK_CHANNEL_ID", "C1234567890")
		defer func() {
			os.Unsetenv("SLACK_TOKEN")
			os.Unsetenv("SLACK_CHANNEL_ID")
		}()

		// Create notifier
		notifier, err := NewSlackNotifier()
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Verify channel ID was cached
		if notifier.channelID != "C1234567890" {
			t.Errorf("Expected channelID to be cached, got: %s", notifier.channelID)
		}

		// Now change env var (should not affect notifier)
		os.Setenv("SLACK_CHANNEL_ID", "C9999999999")

		// Call function - should use cached value, not new env var
		attachments := []slack.Attachment{{Title: "Test"}}
		err = notifier.SendNotificationWithAttachments(attachments)

		// Verify notifier still has original cached channel ID
		if notifier.channelID != "C1234567890" {
			t.Error("Expected channelID to remain cached despite env var change")
		}

		_ = err // API call may fail, that's expected
	})
}
