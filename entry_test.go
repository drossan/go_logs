package go_logs

import (
	"strings"
	"testing"
	"time"
)

// TestEntry_String verifies the String() method formats entries correctly
func TestEntry_String(t *testing.T) {
	entry := &Entry{
		Level:     InfoLevel,
		Message:   "Test message",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Fields:    []Field{String("key", "value")},
	}

	output := entry.String()

	if !strings.Contains(output, "INFO") {
		t.Error("String() should contain level")
	}
	if !strings.Contains(output, "Test message") {
		t.Error("String() should contain message")
	}
	if !strings.Contains(output, "key=value") {
		t.Error("String() should contain fields")
	}
}

// TestEntry_StringWithZeroTimestamp verifies String() handles zero timestamps
func TestEntry_StringWithZeroTimestamp(t *testing.T) {
	entry := &Entry{
		Level:     ErrorLevel,
		Message:   "Error occurred",
		Timestamp: time.Time{}, // Zero timestamp
	}

	output := entry.String()

	if !strings.Contains(output, "ERROR") {
		t.Error("String() should contain level even with zero timestamp")
	}
	if !strings.Contains(output, "Error occurred") {
		t.Error("String() should contain message")
	}
}

// TestEntry_AddFields verifies addFields appends fields correctly
func TestEntry_AddFields(t *testing.T) {
	entry := &Entry{
		Level:   InfoLevel,
		Message: "test",
		Fields:  []Field{String("key1", "value1")},
	}

	entry.addFields(String("key2", "value2"))

	if entry.FieldCount() != 2 {
		t.Errorf("FieldCount() = %d, want 2", entry.FieldCount())
	}

	if !entry.HasField("key2") {
		t.Error("addFields() should add new field")
	}
}

// TestEntry_HasField verifies HasField checks for field existence
func TestEntry_HasField(t *testing.T) {
	entry := &Entry{
		Fields: []Field{
			String("username", "john"),
			Int("age", 30),
		},
	}

	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"Existing field", "username", true},
		{"Existing field 2", "age", true},
		{"Non-existing field", "email", false},
		{"Empty key", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := entry.HasField(tt.key)
			if got != tt.want {
				t.Errorf("HasField(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

// TestEntry_GetField verifies GetField returns the correct field
func TestEntry_GetField(t *testing.T) {
	entry := &Entry{
		Fields: []Field{
			String("username", "john"),
			Int("age", 30),
		},
	}

	t.Run("Existing field", func(t *testing.T) {
		field := entry.GetField("username")
		if field.Key() != "username" {
			t.Errorf("GetField() key = %v, want 'username'", field.Key())
		}
		if field.StringValue() != "john" {
			t.Errorf("GetField() value = %v, want 'john'", field.StringValue())
		}
	})

	t.Run("Non-existing field", func(t *testing.T) {
		field := entry.GetField("email")
		if field.Key() != "" {
			t.Errorf("GetField() non-existing should return zero-value Field, got key = %v", field.Key())
		}
	})
}

// TestEntry_GetFieldValue verifies GetFieldValue returns correct values
func TestEntry_GetFieldValue(t *testing.T) {
	entry := &Entry{
		Fields: []Field{
			String("name", "test"),
			Int("count", 42),
		},
	}

	tests := []struct {
		name  string
		key   string
		check func(interface{}) bool
	}{
		{
			name: "String value",
			key:  "name",
			check: func(v interface{}) bool {
				str, ok := v.(string)
				return ok && str == "test"
			},
		},
		{
			name: "Int value",
			key:  "count",
			check: func(v interface{}) bool {
				i, ok := v.(int)
				return ok && i == 42
			},
		},
		{
			name: "Non-existing value",
			key:  "missing",
			check: func(v interface{}) bool {
				return v == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := entry.GetFieldValue(tt.key)
			if !tt.check(value) {
				t.Errorf("GetFieldValue(%q) check failed", tt.key)
			}
		})
	}
}

// TestEntry_Clone verifies Clone creates a deep copy
func TestEntry_Clone(t *testing.T) {
	original := &Entry{
		Level:     ErrorLevel,
		Message:   "Original message",
		Timestamp: time.Now(),
		Fields:    []Field{String("key", "value")},
	}

	cloned := original.Clone()

	// Verify values are equal
	if cloned.Level != original.Level {
		t.Error("Clone() level should match")
	}
	if cloned.Message != original.Message {
		t.Error("Clone() message should match")
	}
	if len(cloned.Fields) != len(original.Fields) {
		t.Error("Clone() fields length should match")
	}

	// Verify it's a deep copy (not same slices)
	cloned.Message = "Modified message"
	cloned.Fields[0] = String("modified", "field")

	if original.Message == "Modified message" {
		t.Error("Clone() should create deep copy, not share reference")
	}
	if original.Fields[0].Key() == "modified" {
		t.Error("Clone() should create deep copy of fields slice")
	}
}

// TestEntry_WithFields verifies WithFields doesn't mutate original
func TestEntry_WithFields(t *testing.T) {
	original := &Entry{
		Level:  InfoLevel,
		Fields: []Field{String("key1", "value1")},
	}

	modified := original.WithFields(String("key2", "value2"))

	if original.FieldCount() != 1 {
		t.Errorf("WithFields() should not modify original, got %d fields", original.FieldCount())
	}

	if modified.FieldCount() != 2 {
		t.Errorf("WithFields() should add field to clone, got %d fields", modified.FieldCount())
	}
}

// TestEntry_WithLevel verifies WithLevel doesn't mutate original
func TestEntry_WithLevel(t *testing.T) {
	original := &Entry{Level: InfoLevel}
	modified := original.WithLevel(ErrorLevel)

	if original.Level != InfoLevel {
		t.Error("WithLevel() should not modify original")
	}
	if modified.Level != ErrorLevel {
		t.Error("WithLevel() should change clone's level")
	}
}

// TestEntry_WithMessage verifies WithMessage doesn't mutate original
func TestEntry_WithMessage(t *testing.T) {
	original := &Entry{Message: "original"}
	modified := original.WithMessage("modified")

	if original.Message != "original" {
		t.Error("WithMessage() should not modify original")
	}
	if modified.Message != "modified" {
		t.Error("WithMessage() should change clone's message")
	}
}

// TestEntry_WithTimestamp verifies WithTimestamp doesn't mutate original
func TestEntry_WithTimestamp(t *testing.T) {
	original := &Entry{Timestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	newTime := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	modified := original.WithTimestamp(newTime)

	if !original.Timestamp.Equal(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Error("WithTimestamp() should not modify original")
	}
	if !modified.Timestamp.Equal(newTime) {
		t.Error("WithTimestamp() should change clone's timestamp")
	}
}

// TestEntry_FieldCount verifies FieldCount returns correct count
func TestEntry_FieldCount(t *testing.T) {
	tests := []struct {
		name   string
		fields []Field
		want   int
	}{
		{"No fields", []Field{}, 0},
		{"One field", []Field{String("k", "v")}, 1},
		{"Multiple fields", []Field{
			String("k1", "v1"),
			Int("k2", 42),
			Bool("k3", true),
		}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &Entry{Fields: tt.fields}
			got := entry.FieldCount()
			if got != tt.want {
				t.Errorf("FieldCount() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestEntry_FormatBasic verifies FormatBasic creates correct format
func TestEntry_FormatBasic(t *testing.T) {
	entry := &Entry{
		Level:     WarnLevel,
		Message:   "Warning message",
		Timestamp: time.Date(2025, 6, 15, 14, 30, 0, 0, time.UTC),
		Fields:    []Field{String("service", "api")},
	}

	output := entry.FormatBasic()

	if !strings.Contains(output, "2025-06-15T14:30:00Z") {
		t.Error("FormatBasic() should contain RFC3339 timestamp")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("FormatBasic() should contain level")
	}
	if !strings.Contains(output, "Warning message") {
		t.Error("FormatBasic() should contain message")
	}
	if !strings.Contains(output, "service=api") {
		t.Error("FormatBasic() should contain fields")
	}
}

// TestEntry_GetFieldsByPrefix verifies field filtering by prefix
func TestEntry_GetFieldsByPrefix(t *testing.T) {
	entry := &Entry{
		Fields: []Field{
			String("http.method", "GET"),
			String("http.url", "/api/users"),
			String("user.id", "123"),
			String("service.name", "api"),
		},
	}

	httpFields := entry.GetFieldsByPrefix("http.")

	if len(httpFields) != 2 {
		t.Errorf("GetFieldsByPrefix('http.') = %d fields, want 2", len(httpFields))
	}

	for _, field := range httpFields {
		if !strings.HasPrefix(field.Key(), "http.") {
			t.Errorf("GetFieldsByPrefix() returned non-matching field: %s", field.Key())
		}
	}
}

// TestEntry_RemoveField verifies field removal
func TestEntry_RemoveField(t *testing.T) {
	entry := &Entry{
		Fields: []Field{
			String("key1", "value1"),
			String("key2", "value2"),
			String("key3", "value3"),
		},
	}

	t.Run("Remove existing field", func(t *testing.T) {
		initialCount := entry.FieldCount()
		removed := entry.RemoveField("key2")

		if !removed {
			t.Error("RemoveField() should return true for existing field")
		}
		if entry.FieldCount() != initialCount-1 {
			t.Errorf("RemoveField() should decrease count, got %d", entry.FieldCount())
		}
		if entry.HasField("key2") {
			t.Error("RemoveField() should actually remove the field")
		}
	})

	t.Run("Remove non-existing field", func(t *testing.T) {
		initialCount := entry.FieldCount()
		removed := entry.RemoveField("missing")

		if removed {
			t.Error("RemoveField() should return false for non-existing field")
		}
		if entry.FieldCount() != initialCount {
			t.Error("RemoveField() should not change count for non-existing field")
		}
	})
}

// TestEntry_ReplaceFieldValue verifies field value replacement
func TestEntry_ReplaceFieldValue(t *testing.T) {
	entry := &Entry{
		Fields: []Field{
			String("username", "john"),
			Int("age", 30),
		},
	}

	t.Run("Replace existing field", func(t *testing.T) {
		replaced := entry.ReplaceFieldValue("username", "jane")

		if !replaced {
			t.Error("ReplaceFieldValue() should return true for existing field")
		}

		field := entry.GetField("username")
		if field.StringValue() != "jane" {
			t.Errorf("ReplaceFieldValue() should update value, got %s", field.StringValue())
		}
	})

	t.Run("Replace non-existing field", func(t *testing.T) {
		replaced := entry.ReplaceFieldValue("email", "test@example.com")

		if replaced {
			t.Error("ReplaceFieldValue() should return false for non-existing field")
		}
	})
}

// BenchmarkEntry_String measures String() performance
func BenchmarkEntry_String(b *testing.B) {
	entry := &Entry{
		Level:     InfoLevel,
		Message:   "Test message",
		Timestamp: time.Now(),
		Fields:    []Field{String("key", "value"), Int("count", 42)},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = entry.String()
	}
}

// BenchmarkEntry_Clone measures Clone() performance
func BenchmarkEntry_Clone(b *testing.B) {
	entry := &Entry{
		Level:     InfoLevel,
		Message:   "Test message",
		Timestamp: time.Now(),
		Fields:    []Field{String("key", "value")},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = entry.Clone()
	}
}

// BenchmarkEntry_AddFields measures addFields() performance
func BenchmarkEntry_AddFields(b *testing.B) {
	fields := []Field{String("key", "value")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := &Entry{Fields: []Field{}}
		entry.addFields(fields...)
	}
}
