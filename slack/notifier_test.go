package slack

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/drossan/go_logs/v3/hooks"
	slackapi "github.com/slack-go/slack"
)

// Compile-time check: *Notifier can be used as the notifier of hooks.SlackHook.
var _ hooks.SlackNotifier = (*Notifier)(nil)

// setEnv sets the Slack environment variables for the test; an empty value
// leaves the variable empty (t.Setenv restores the previous value afterwards).
func setEnv(t *testing.T, token, channel, legacyChannel string) {
	t.Helper()
	t.Setenv("SLACK_TOKEN", token)
	t.Setenv("SLACK_CHANNEL_ID", channel)
	t.Setenv("SLACK_CHANEL_ID", legacyChannel)
}

func TestNewNotifierFromEnv(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		channel     string
		wantErr     error
		wantMention string
	}{
		{"no token no channel", "", "", ErrTokenMissing, "SLACK_TOKEN"},
		{"token without channel", "xoxb-1", "", ErrChannelMissing, "SLACK_CHANNEL_ID"},
		{"token and channel", "xoxb-1", "C123", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.token, tt.channel, "")

			n, err := NewNotifierFromEnv()

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if n == nil || !n.IsEnabled() {
					t.Fatalf("got notifier %+v, want a non-nil enabled notifier", n)
				}
				if n.channelID != tt.channel {
					t.Errorf("channelID = %q, want %q", n.channelID, tt.channel)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want errors.Is %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantMention) {
				t.Errorf("error %q does not mention %q", err, tt.wantMention)
			}
			if n == nil || n.IsEnabled() {
				t.Errorf("got notifier %+v, want a non-nil disabled notifier", n)
			}
		})
	}
}

func TestNewNotifierFromEnv_ErrorDoesNotLeakToken(t *testing.T) {
	setEnv(t, "xoxb-secret-value", "", "")

	_, err := NewNotifierFromEnv()

	if err == nil || strings.Contains(err.Error(), "xoxb-secret-value") {
		t.Errorf("error %v must exist and must not contain the token", err)
	}
}

func TestNewNotifierFromEnv_LegacyChannelVariable(t *testing.T) {
	setEnv(t, "xoxb-1", "", "C999")

	n, err := NewNotifierFromEnv()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil || !n.IsEnabled() || n.channelID != "C999" {
		t.Fatalf("got notifier %+v, want enabled notifier with channel C999", n)
	}
}

func TestNewNotifierFromEnv_CorrectChannelWinsOverLegacy(t *testing.T) {
	setEnv(t, "xoxb-1", "C123", "C999")

	n, err := NewNotifierFromEnv()

	if err != nil || n.channelID != "C123" {
		t.Fatalf("got (%+v, %v), want channel C123 and no error", n, err)
	}
}

func TestNewNotifier(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		channel     string
		wantErr     error
		wantMention string
	}{
		{"no token no channel", "", "", ErrTokenMissing, "token"},
		{"token without channel", "xoxb-1", "", ErrChannelMissing, "channel"},
		{"token and channel", "xoxb-1", "C1", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The environment must not be consulted.
			setEnv(t, "xoxb-env", "CENV", "CENV")

			n, err := NewNotifier(tt.token, tt.channel)

			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if n == nil || !n.IsEnabled() || n.channelID != tt.channel {
					t.Fatalf("got notifier %+v, want enabled notifier with channel %q", n, tt.channel)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want errors.Is %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantMention) {
				t.Errorf("error %q does not mention %q", err, tt.wantMention)
			}
			if n == nil || n.IsEnabled() {
				t.Errorf("got notifier %+v, want a non-nil disabled notifier", n)
			}
		})
	}
}

func TestDisabledNotifier_SendIsNoOp(t *testing.T) {
	n, _ := NewNotifier("", "")

	if err := n.SendNotification("msg"); err != nil {
		t.Errorf("SendNotification on disabled notifier = %v, want nil", err)
	}
	if err := n.SendNotificationWithAttachments([]slackapi.Attachment{{Title: "t"}}); err != nil {
		t.Errorf("SendNotificationWithAttachments on disabled notifier = %v, want nil", err)
	}
}

func TestNilNotifier_SendIsNoOp(t *testing.T) {
	var n *Notifier

	if err := n.SendNotification("msg"); err != nil {
		t.Errorf("SendNotification on nil notifier = %v, want nil", err)
	}
	if n.IsEnabled() {
		t.Error("nil notifier reports enabled")
	}
}

// fakeSlackAPI starts a local server that answers chat.postMessage and counts
// calls; ok controls whether Slack reports success.
func fakeSlackAPI(t *testing.T, ok bool) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if ok {
			_, _ = w.Write([]byte(`{"ok":true,"channel":"C123","ts":"1.0"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func TestSendNotification_PostsToSlack(t *testing.T) {
	srv, calls := fakeSlackAPI(t, true)
	n, err := NewNotifier("xoxb-1", "C123")
	if err != nil {
		t.Fatal(err)
	}
	n.Client = slackapi.New("xoxb-1", slackapi.OptionAPIURL(srv.URL+"/"))

	if err := n.SendNotification("hola"); err != nil {
		t.Fatalf("SendNotification = %v, want nil", err)
	}
	if err := n.SendNotificationWithAttachments([]slackapi.Attachment{{Title: "a"}, {Title: "b"}}); err != nil {
		t.Fatalf("SendNotificationWithAttachments = %v, want nil", err)
	}
	if got := calls.Load(); got != 2 {
		t.Errorf("Slack API called %d times, want 2", got)
	}
}

func TestSendNotification_ReturnsSlackError(t *testing.T) {
	srv, _ := fakeSlackAPI(t, false)
	n, err := NewNotifier("xoxb-1", "C123")
	if err != nil {
		t.Fatal(err)
	}
	n.Client = slackapi.New("xoxb-1", slackapi.OptionAPIURL(srv.URL+"/"))

	if err := n.SendNotification("hola"); err == nil {
		t.Error("SendNotification = nil, want the Slack API error")
	}
	if err := n.SendNotificationWithAttachments(nil); err == nil {
		t.Error("SendNotificationWithAttachments = nil, want the Slack API error")
	}
}

func TestChannelIsCachedAtConstruction(t *testing.T) {
	setEnv(t, "xoxb-1", "C123", "")
	n, err := NewNotifierFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("SLACK_CHANNEL_ID", "C999")

	if n.channelID != "C123" {
		t.Errorf("channelID = %q after changing env, want cached C123", n.channelID)
	}
}
