package go_logs

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// fakeNotifier records every message it receives. When fail is true,
// SendNotification returns an error after recording the message.
type fakeNotifier struct {
	mu       sync.Mutex
	messages []string
	fail     bool
}

func (f *fakeNotifier) SendNotification(message string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = append(f.messages, message)
	if f.fail {
		return errors.New("fake notifier failure")
	}
	return nil
}

func (f *fakeNotifier) received() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.messages...)
}

const missingNotifierWarning = "NOTIFICATIONS_SLACK_ENABLED activo pero no hay notificador; llama a go_logs.SetNotifier"

// setupNotifierTest resets the v2 configuration, applies the given
// NOTIFICATIONS_SLACK_ENABLED value with error notifications on, runs Init(),
// re-arms the once-per-process warning (a fresh process for this test) and
// unregisters any notifier. It returns the stderr warnings and the standard
// log output buffers.
func setupNotifierTest(t *testing.T, slackEnabled string) (warnings, stdout *bytes.Buffer) {
	t.Helper()
	warnings = resetConfigForTest(t)
	stdout = captureStdLog(t)
	t.Setenv("NOTIFICATIONS_SLACK_ENABLED", slackEnabled)
	t.Setenv("NOTIFICATION_ERROR_LOG", "1")
	missingNotifierWarnOnce = sync.Once{}
	SetNotifier(nil)
	t.Cleanup(func() { SetNotifier(nil) })
	Init()
	return warnings, stdout
}

func TestSetNotifier_SendsWhenEnabled(t *testing.T) {
	setupNotifierTest(t, "1")
	fake := &fakeNotifier{}
	SetNotifier(fake)

	ErrorLog("fallo")

	got := fake.received()
	if len(got) != 1 || !strings.Contains(got[0], "fallo") {
		t.Fatalf("notifier received %q, want one message containing %q", got, "fallo")
	}
	if !IsNotifierEnabled() {
		t.Error("IsNotifierEnabled() = false with a registered notifier and notifications on")
	}
}

func TestNoNotifier_WarnsOnceAndKeepsLogging(t *testing.T) {
	warnings, stdout := setupNotifierTest(t, "1")

	ErrorLog("primer error")
	ErrorLog("segundo error")

	if n := strings.Count(warnings.String(), missingNotifierWarning); n != 1 {
		t.Errorf("warning appeared %d times, want exactly 1; stderr:\n%s", n, warnings.String())
	}
	for _, msg := range []string{"primer error", "segundo error"} {
		if !strings.Contains(stdout.String(), msg) {
			t.Errorf("stdout does not contain %q:\n%s", msg, stdout.String())
		}
	}
	if IsNotifierEnabled() {
		t.Error("IsNotifierEnabled() = true without a registered notifier")
	}
}

func TestNoNotifier_WarningNotRearmedAfterToggle(t *testing.T) {
	warnings, _ := setupNotifierTest(t, "1")

	ErrorLog("dispara el aviso")
	SetNotifier(&fakeNotifier{})
	SetNotifier(nil)
	ErrorLog("otro error")

	if n := strings.Count(warnings.String(), missingNotifierWarning); n != 1 {
		t.Errorf("warning appeared %d times, want exactly 1; stderr:\n%s", n, warnings.String())
	}
}

func TestFailingNotifier_DoesNotInterruptLogging(t *testing.T) {
	_, stdout := setupNotifierTest(t, "1")
	fake := &fakeNotifier{fail: true}
	SetNotifier(fake)

	ErrorLog("error con notificador roto")

	if !strings.Contains(stdout.String(), "error con notificador roto") {
		t.Errorf("stdout does not contain the error:\n%s", stdout.String())
	}
	if len(fake.received()) != 1 {
		t.Errorf("notifier received %d messages, want 1", len(fake.received()))
	}
}

func TestSetNotifierNil_DisablesNotifications(t *testing.T) {
	setupNotifierTest(t, "1")
	fake := &fakeNotifier{}
	SetNotifier(fake)
	SetNotifier(nil)

	ErrorLog("no debe notificarse")

	if got := fake.received(); len(got) != 0 {
		t.Errorf("notifier received %q after SetNotifier(nil), want nothing", got)
	}
}

func TestNotificationsDisabled_NotifierUnused(t *testing.T) {
	warnings, _ := setupNotifierTest(t, "0")
	fake := &fakeNotifier{}
	SetNotifier(fake)

	ErrorLog("no debe notificarse")

	if got := fake.received(); len(got) != 0 {
		t.Errorf("notifier received %q with NOTIFICATIONS_SLACK_ENABLED=0, want nothing", got)
	}
	if IsNotifierEnabled() {
		t.Error("IsNotifierEnabled() = true with NOTIFICATIONS_SLACK_ENABLED=0")
	}
	if strings.Contains(warnings.String(), missingNotifierWarning) {
		t.Errorf("missing-notifier warning emitted with notifications off:\n%s", warnings.String())
	}
}

func TestSetNotifier_ConcurrentWithErrorLog(t *testing.T) {
	setupNotifierTest(t, "1")

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if (i+j)%2 == 0 {
					SetNotifier(&fakeNotifier{})
				} else {
					SetNotifier(nil)
				}
				ErrorLog("concurrente")
				_ = IsNotifierEnabled()
			}
		}(i)
	}
	wg.Wait()
}

func TestCore_AdaptersPackageRemoved(t *testing.T) {
	if _, err := os.Stat("adapters"); !os.IsNotExist(err) {
		t.Errorf("adapters/ still exists (stat err = %v)", err)
	}
}

func TestCore_NoSlackDependency(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go binary not in PATH")
	}
	out, err := exec.Command(goBin, "list", "-deps", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps failed: %v\n%s", err, out)
	}
	for _, dep := range []string{"slack-go", "gorilla/websocket"} {
		if strings.Contains(string(out), dep) {
			t.Errorf("root package depends on %s", dep)
		}
	}
}
