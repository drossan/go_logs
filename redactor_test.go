package go_logs

import (
	"testing"
)

// TestNewRedactor tests redactor creation
func TestNewRedactor(t *testing.T) {
	keys := []string{"password", "token", "secret"}
	redactor := NewRedactor(keys...)

	if redactor == nil {
		t.Fatal("NewRedactor returned nil")
	}

	if len(redactor.sensitiveKeys) != len(keys) {
		t.Errorf("Expected %d sensitive keys, got %d", len(keys), len(redactor.sensitiveKeys))
	}

	// Verify all keys are present
	for _, key := range keys {
		if !redactor.sensitiveKeys[key] {
			t.Errorf("Key %q not found in sensitive keys", key)
		}
	}

	// Verify default mask value
	if redactor.maskValue != "***" {
		t.Errorf("Expected mask value '***', got %q", redactor.maskValue)
	}
}

// TestNewRedactor_EmptyKeys tests redactor with no keys
func TestNewRedactor_EmptyKeys(t *testing.T) {
	redactor := NewRedactor()

	if redactor == nil {
		t.Fatal("NewRedactor returned nil")
	}

	if len(redactor.sensitiveKeys) != 0 {
		t.Errorf("Expected 0 sensitive keys, got %d", len(redactor.sensitiveKeys))
	}
}

// TestRedactor_Redact tests basic redaction functionality
func TestRedactor_Redact(t *testing.T) {
	redactor := NewRedactor("password", "token")

	entry := &Entry{
		Message: "User login",
		Fields: []Field{
			String("username", "john"),
			String("password", "secret123"),
			String("token", "xyz789"),
			String("email", "john@example.com"),
		},
	}

	redactor.Redact(entry)

	// Verify password is masked
	if entry.Fields[1].Key() != "password" {
		t.Errorf("Expected key 'password', got %q", entry.Fields[1].Key())
	}
	if entry.Fields[1].Value() != "***" {
		t.Errorf("Expected password to be masked with '***', got %q", entry.Fields[1].Value())
	}

	// Verify token is masked
	if entry.Fields[2].Key() != "token" {
		t.Errorf("Expected key 'token', got %q", entry.Fields[2].Key())
	}
	if entry.Fields[2].Value() != "***" {
		t.Errorf("Expected token to be masked with '***', got %q", entry.Fields[2].Value())
	}

	// Verify username is NOT masked
	if entry.Fields[0].Key() != "username" {
		t.Errorf("Expected key 'username', got %q", entry.Fields[0].Key())
	}
	if entry.Fields[0].Value() != "john" {
		t.Errorf("Expected username to not be masked, got %q", entry.Fields[0].Value())
	}

	// Verify email is NOT masked
	if entry.Fields[3].Key() != "email" {
		t.Errorf("Expected key 'email', got %q", entry.Fields[3].Key())
	}
	if entry.Fields[3].Value() != "john@example.com" {
		t.Errorf("Expected email to not be masked, got %q", entry.Fields[3].Value())
	}
}

// TestRedactor_RedactNoMatch tests redaction when no keys match
func TestRedactor_RedactNoMatch(t *testing.T) {
	redactor := NewRedactor("password", "secret")

	entry := &Entry{
		Message: "Request",
		Fields: []Field{
			String("username", "john"),
			String("email", "john@example.com"),
		},
	}

	originalUsername := entry.Fields[0].Value()
	originalEmail := entry.Fields[1].Value()

	redactor.Redact(entry)

	// Nothing should be masked
	if entry.Fields[0].Value() != originalUsername {
		t.Errorf("Expected username to remain unchanged, got %q", entry.Fields[0].Value())
	}
	if entry.Fields[1].Value() != originalEmail {
		t.Errorf("Expected email to remain unchanged, got %q", entry.Fields[1].Value())
	}
}

// TestRedactor_RedactEmptyFields tests redaction with no fields
func TestRedactor_RedactEmptyFields(t *testing.T) {
	redactor := NewRedactor("password")

	entry := &Entry{
		Message: "Test message",
		Fields:  []Field{},
	}

	// Should not panic
	redactor.Redact(entry)
}

// TestRedactor_RedactCaseSensitive tests that redaction is case-sensitive
func TestRedactor_RedactCaseSensitive(t *testing.T) {
	redactor := NewRedactor("Password") // Capital P

	entry := &Entry{
		Message: "Login",
		Fields: []Field{
			String("Password", "secret123"), // Should be masked
			String("password", "secret456"), // Should NOT be masked (case mismatch)
		},
	}

	redactor.Redact(entry)

	if entry.Fields[0].Value() != "***" {
		t.Errorf("Expected Password to be masked, got %q", entry.Fields[0].Value())
	}

	if entry.Fields[1].Value() != "secret456" {
		t.Errorf("Expected password (lowercase) to not be masked, got %q", entry.Fields[1].Value())
	}
}

// TestRedactor_RedactMultipleTypes tests redaction with different field types
func TestRedactor_RedactMultipleTypes(t *testing.T) {
	redactor := NewRedactor("password", "count")

	entry := &Entry{
		Message: "Test",
		Fields: []Field{
			String("password", "secret123"),
			Int("count", 42),
			Float64("price", 19.99),
			Bool("enabled", true),
		},
	}

	redactor.Redact(entry)

	// Verify password (String) is masked
	if entry.Fields[0].Value() != "***" {
		t.Errorf("Expected password to be masked, got %q", entry.Fields[0].Value())
	}

	// Verify count (Int) is masked
	if entry.Fields[1].Value() != "***" {
		t.Errorf("Expected count to be masked, got %q", entry.Fields[1].Value())
	}

	// Verify other fields are not masked
	if entry.Fields[2].Value() != 19.99 {
		t.Errorf("Expected price to not be masked, got %v", entry.Fields[2].Value())
	}

	if entry.Fields[3].Value() != true {
		t.Errorf("Expected enabled to not be masked, got %v", entry.Fields[3].Value())
	}
}

// TestCommonSensitiveKeys tests the predefined list of sensitive keys
func TestCommonSensitiveKeys(t *testing.T) {
	keys := CommonSensitiveKeys()

	expectedKeys := []string{
		"password", "passwd", "pwd",
		"token", "api_key", "apikey", "api-key",
		"secret", "authorization", "auth",
		"cookie", "session",
		"credit_card", "ssn", "social_security",
	}

	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	// Verify all expected keys are present
	keyMap := make(map[string]bool)
	for _, key := range keys {
		keyMap[key] = true
	}

	for _, expectedKey := range expectedKeys {
		if !keyMap[expectedKey] {
			t.Errorf("Expected key %q not found in CommonSensitiveKeys()", expectedKey)
		}
	}
}

// TestWithCommonRedaction tests the helper function
func TestWithCommonRedaction(t *testing.T) {
	option := WithCommonRedaction()

	if option == nil {
		t.Fatal("WithCommonRedaction returned nil")
	}

	redactorOption, ok := option.(*RedactorOption)
	if !ok {
		t.Fatal("WithCommonRedaction did not return a RedactorOption")
	}

	// Verify it contains common keys
	expectedKeys := CommonSensitiveKeys()
	if len(redactorOption.Keys) != len(expectedKeys) {
		t.Errorf("Expected %d keys, got %d", len(expectedKeys), len(redactorOption.Keys))
	}
}

// TestRedactor_CommonKeys tests redaction with common sensitive keys
func TestRedactor_CommonKeys(t *testing.T) {
	redactor := NewRedactor(CommonSensitiveKeys()...)

	entry := &Entry{
		Message: "API Request",
		Fields: []Field{
			String("username", "john"),
			String("password", "secret123"),
			String("token", "abc123"),
			String("api_key", "xyz789"),
			String("secret", "hidden"),
			String("authorization", "Bearer token"),
			String("cookie", "session=abc"),
			String("ssn", "123-45-6789"),
			String("email", "john@example.com"),
		},
	}

	redactor.Redact(entry)

	// All sensitive fields should be masked
	sensitiveFields := []string{"password", "token", "api_key", "secret", "authorization", "cookie", "ssn"}
	for _, fieldName := range sensitiveFields {
		found := false
		for _, field := range entry.Fields {
			if field.Key() == fieldName {
				if field.Value() != "***" {
					t.Errorf("Expected %s to be masked, got %q", fieldName, field.Value())
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Field %q not found in entry", fieldName)
		}
	}

	// Non-sensitive fields should NOT be masked
	if entry.Fields[0].Value() != "john" {
		t.Errorf("Expected username to not be masked, got %q", entry.Fields[0].Value())
	}
	if entry.Fields[8].Value() != "john@example.com" {
		t.Errorf("Expected email to not be masked, got %q", entry.Fields[8].Value())
	}
}

// TestRedactor_DuplicateKeys tests redactor with duplicate keys in constructor
func TestRedactor_DuplicateKeys(t *testing.T) {
	// Create redactor with duplicate keys
	redactor := NewRedactor("password", "password", "token", "token")

	if len(redactor.sensitiveKeys) != 2 {
		t.Errorf("Expected 2 unique sensitive keys, got %d", len(redactor.sensitiveKeys))
	}

	// Verify both keys work
	entry := &Entry{
		Message: "Test",
		Fields: []Field{
			String("password", "secret123"),
			String("token", "abc123"),
		},
	}

	redactor.Redact(entry)

	if entry.Fields[0].Value() != "***" {
		t.Errorf("Expected password to be masked")
	}

	if entry.Fields[1].Value() != "***" {
		t.Errorf("Expected token to be masked")
	}
}

// TestRedactor_WithLogger tests redactor integration with logger
func TestRedactor_WithLogger(t *testing.T) {
	logger, buf := newTestLogger(WithRedactor("password"))

	logger.Info("User login",
		String("username", "john"),
		String("password", "secret123"),
	)

	output := buf.String()

	// Verify password is masked in output
	if contains(output, "secret123") {
		t.Error("Password should be masked in log output")
	}

	if !contains(output, "***") {
		t.Error("Masked value *** should appear in log output")
	}

	if !contains(output, "username=john") {
		t.Error("Username should appear in log output")
	}
}

// TestRedactor_WithCommonRedactionLogger tests redactor with common keys in logger
func TestRedactor_WithCommonRedactionLogger(t *testing.T) {
	logger, buf := newTestLogger(WithCommonRedaction())

	logger.Info("API Request",
		String("api_key", "secret-key-123"),
		String("token", "auth-token-456"),
		String("endpoint", "/api/users"),
	)

	output := buf.String()

	// Verify sensitive fields are masked
	if contains(output, "secret-key-123") {
		t.Error("api_key should be masked in log output")
	}

	if contains(output, "auth-token-456") {
		t.Error("token should be masked in log output")
	}

	// Verify non-sensitive fields are NOT masked
	if !contains(output, "endpoint=/api/users") {
		t.Error("endpoint should appear in log output")
	}
}

// TestRedactor_MultipleLoggers tests that redactor works with multiple loggers
func TestRedactor_MultipleLoggers(t *testing.T) {
	logger1, buf1 := newTestLogger(WithRedactor("password"))
	logger2, buf2 := newTestLogger(WithRedactor("token"))

	logger1.Info("Login",
		String("password", "secret123"),
		String("token", "abc123"),
	)

	logger2.Info("API Call",
		String("password", "secret456"),
		String("token", "xyz789"),
	)

	output1 := buf1.String()
	output2 := buf2.String()

	// Logger1 masks password but not token
	if contains(output1, "secret123") {
		t.Error("password should be masked in logger1")
	}
	if !contains(output1, "token=abc123") {
		t.Error("token should NOT be masked in logger1")
	}

	// Logger2 masks token but not password
	if contains(output2, "xyz789") {
		t.Error("token should be masked in logger2")
	}
	if !contains(output2, "password=secret456") {
		t.Error("password should NOT be masked in logger2")
	}
}

// TestRedactor_ChildLoggerInheritance tests that redactor is inherited by child loggers
func TestRedactor_ChildLoggerInheritance(t *testing.T) {
	parent, buf := newTestLogger(WithRedactor("password"))

	child := parent.With(String("request_id", "123"))

	child.Info("Login",
		String("username", "john"),
		String("password", "secret123"),
	)

	output := buf.String()

	// Child logger should inherit parent's redactor
	if contains(output, "secret123") {
		t.Error("password should be masked in child logger")
	}

	if !contains(output, "request_id=123") {
		t.Error("request_id should appear in output")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
