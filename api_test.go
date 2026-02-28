package go_logs

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"
)

// TestInfof verifies Infof formats and logs messages correctly
func TestInfof(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"INFO": false, // Disable file logging for this test
	}
	notificationSettingsMutex.Unlock()

	// Call function under test
	testFormat := "Test info message: %s, number: %d"
	Infof(testFormat, "hello", 42)

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from Infof, got empty")
	}
}

// TestErrorf verifies Errorf formats and logs messages correctly
func TestErrorf(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"ERROR": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test
	testFormat := "Error %d: %s"
	Errorf(testFormat, 500, "Internal Server Error")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from Errorf, got empty")
	}
}

// TestWarningf verifies Warningf formats and logs messages correctly
func TestWarningf(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"WARNING": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test
	testFormat := "Warning: %s"
	Warningf(testFormat, "deprecated API usage")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from Warningf, got empty")
	}
}

// TestSuccessf verifies Successf formats and logs messages correctly
func TestSuccessf(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"SUCCESS": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test
	testFormat := "Operation %s completed in %dms"
	Successf(testFormat, "database migration", 250)

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from Successf, got empty")
	}
}

// TestWarningLog verifies WarningLog function works
// Issue #14: Previously notificationLogWarning was loaded but never used
func TestWarningLog(t *testing.T) {
	// Reset state from any previous test
	Close()

	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup: Enable file logging
	tempDir := t.TempDir()
	logFilePath = tempDir
	logFileName = "test_warning.log"
	saveLogFile = true
	isInit = true

	// Initialize notification settings
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"WARNING": true, // Enable WARNING logging
	}
	useLegacySystem = true // Enable legacy system for file logging
	notificationSettingsMutex.Unlock()

	// Call function under test
	testMessage := "This is a warning"
	WarningLog(testMessage)

	// Verify console output
	consoleOutput := buf.String()
	if consoleOutput == "" {
		t.Error("Expected console output from WarningLog, got empty string")
	}

	// Flush buffer to ensure content is written to file
	Close()

	// Verify file was created
	fullPath := logFilePath + "/" + logFileName
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	fileContent := string(content)
	if fileContent == "" {
		t.Error("Expected file to contain log message, got empty string")
	}

	// Cleanup
	os.Remove(fullPath)
}

// TestInfoLogCtx verifies context support for InfoLog
// Issue #11: Context support for distributed tracing
func TestInfoLogCtx(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"INFO": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test with context
	ctx := context.Background()
	InfoLogCtx(ctx, "Test message with context")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from InfoLogCtx, got empty")
	}
}

// TestErrorLogCtx verifies context support for ErrorLog
func TestErrorLogCtx(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"ERROR": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test with context
	ctx := context.Background()
	ErrorLogCtx(ctx, "Error with context")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from ErrorLogCtx, got empty")
	}
}

// TestWarningLogCtx verifies context support for WarningLog
func TestWarningLogCtx(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"WARNING": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test with context
	ctx := context.Background()
	WarningLogCtx(ctx, "Warning with context")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from WarningLogCtx, got empty")
	}
}

// TestSuccessLogCtx verifies context support for SuccessLog
func TestSuccessLogCtx(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"SUCCESS": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test with context
	ctx := context.Background()
	SuccessLogCtx(ctx, "Success with context")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from SuccessLogCtx, got empty")
	}
}

// TestInfoLogCtxf verifies combined context + formatting support
// Issue #11 + #13: Combined context and formatted logging
func TestInfoLogCtxf(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"INFO": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test with context and formatting
	ctx := context.Background()
	InfoLogCtxf(ctx, "User %s logged in from %s", "john_doe", "192.168.1.1")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from InfoLogCtxf, got empty")
	}
}

// TestErrorLogCtxf verifies combined context + formatting for errors
func TestErrorLogCtxf(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"ERROR": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test
	ctx := context.Background()
	ErrorLogCtxf(ctx, "Request failed with status %d: %s", 500, "Internal Server Error")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from ErrorLogCtxf, got empty")
	}
}

// TestWarningLogCtxf verifies combined context + formatting for warnings
func TestWarningLogCtxf(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"WARNING": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test
	ctx := context.Background()
	WarningLogCtxf(ctx, "API version %s is deprecated, use %s", "v1", "v2")

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from WarningLogCtxf, got empty")
	}
}

// TestSuccessLogCtxf verifies combined context + formatting for success
func TestSuccessLogCtxf(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"SUCCESS": false,
	}
	notificationSettingsMutex.Unlock()

	// Call function under test
	ctx := context.Background()
	SuccessLogCtxf(ctx, "Processed %d records in %dms", 1000, 45)

	// Verify output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from SuccessLogCtxf, got empty")
	}
}

// TestBackwardCompatibility verifies old API still works
func TestBackwardCompatibility(t *testing.T) {
	// Setup: Redirect log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Setup minimal config
	isInit = true
	notificationSettingsMutex.Lock()
	notificationSettings = map[string]bool{
		"INFO":     false,
		"ERROR":    false,
		"SUCCESS":  false,
		"WARNING":  false,
	}
	notificationSettingsMutex.Unlock()

	// Test old API still works
	InfoLog("Test InfoLog")
	ErrorLog("Test ErrorLog")
	SuccessLog("Test SuccessLog")
	WarningLog("Test WarningLog")

	// Verify all produced output
	output := buf.String()
	if output == "" {
		t.Error("Expected output from old API functions, got empty")
	}
}
