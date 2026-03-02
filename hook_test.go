package go_logs

import (
	"errors"
	"testing"
)

// TestHookFunc tests that HookFunc adapter works
func TestHookFunc(t *testing.T) {
	called := false
	testError := errors.New("test error")

	// Create a hook from a function
	hook := HookFunc(func(entry *Entry) error {
		called = true
		return testError
	})

	entry := &Entry{Message: "test"}

	// Run the hook
	err := hook.Run(entry)

	if !called {
		t.Error("HookFunc was not called")
	}

	if err != testError {
		t.Errorf("HookFunc returned wrong error, got %v, want %v", err, testError)
	}
}

// TestNewFuncHook tests creating hooks from functions
func TestNewFuncHook(t *testing.T) {
	callCount := 0

	hook := NewFuncHook(func(entry *Entry) error {
		callCount++
		return nil
	})

	entry := &Entry{Message: "test"}

	// Run multiple times
	hook.Run(entry)
	hook.Run(entry)
	hook.Run(entry)

	if callCount != 3 {
		t.Errorf("FuncHook called %d times, want 3", callCount)
	}
}

// TestHookWithEntry tests that hooks receive the correct entry
func TestHookWithEntry(t *testing.T) {
	var receivedEntry *Entry

	hook := NewFuncHook(func(entry *Entry) error {
		receivedEntry = entry
		return nil
	})

	testEntry := &Entry{
		Message: "test message",
		Level:   InfoLevel,
		Fields:  []Field{String("key", "value")},
	}

	hook.Run(testEntry)

	if receivedEntry == nil {
		t.Fatal("Hook did not receive entry")
	}

	if receivedEntry.Message != testEntry.Message {
		t.Errorf("Hook received wrong message, got %s, want %s", receivedEntry.Message, testEntry.Message)
	}

	if receivedEntry.Level != testEntry.Level {
		t.Errorf("Hook received wrong level, got %v, want %v", receivedEntry.Level, testEntry.Level)
	}

	if len(receivedEntry.Fields) != len(testEntry.Fields) {
		t.Errorf("Hook received wrong number of fields, got %d, want %d", len(receivedEntry.Fields), len(testEntry.Fields))
	}
}

// TestHookCanModifyEntry tests that hooks can modify entry fields
func TestHookCanModifyEntry(t *testing.T) {
	hook := NewFuncHook(func(entry *Entry) error {
		// Modify entry fields
		entry.Fields = append(entry.Fields, String("added_by_hook", "yes"))
		return nil
	})

	entry := &Entry{
		Message: "test",
		Fields:  []Field{String("original", "field")},
	}

	hook.Run(entry)

	// Entry should be modified
	if len(entry.Fields) != 2 {
		t.Errorf("Entry should have 2 fields after hook, got %d", len(entry.Fields))
	}

	found := false
	for _, f := range entry.Fields {
		if f.Key() == "added_by_hook" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Hook did not add field to entry")
	}
}

// TestHookErrorHandling tests that hook errors are handled gracefully
func TestHookErrorHandling(t *testing.T) {
	testError := errors.New("hook failed")

	hook := NewFuncHook(func(entry *Entry) error {
		return testError
	})

	entry := &Entry{Message: "test"}
	err := hook.Run(entry)

	if err != testError {
		t.Errorf("Hook should return error, got %v, want %v", err, testError)
	}
}

// TestMultipleHooks tests that multiple hooks can be executed
func TestMultipleHooks(t *testing.T) {
	callOrder := []string{}

	hook1 := NewFuncHook(func(entry *Entry) error {
		callOrder = append(callOrder, "hook1")
		return nil
	})

	hook2 := NewFuncHook(func(entry *Entry) error {
		callOrder = append(callOrder, "hook2")
		return nil
	})

	hook3 := NewFuncHook(func(entry *Entry) error {
		callOrder = append(callOrder, "hook3")
		return nil
	})

	entry := &Entry{Message: "test"}

	// Execute hooks in order
	hooks := []Hook{hook1, hook2, hook3}
	for _, hook := range hooks {
		_ = hook.Run(entry)
	}

	if len(callOrder) != 3 {
		t.Fatalf("Expected 3 hook calls, got %d", len(callOrder))
	}

	if callOrder[0] != "hook1" || callOrder[1] != "hook2" || callOrder[2] != "hook3" {
		t.Errorf("Hooks called in wrong order: %v", callOrder)
	}
}

// TestHookWithCondition tests conditional hook execution
func TestHookWithCondition(t *testing.T) {
	errorCount := 0

	hook := NewFuncHook(func(entry *Entry) error {
		if entry.Level >= ErrorLevel {
			errorCount++
		}
		return nil
	})

	tests := []struct {
		level     Level
		shouldInc bool
	}{
		{InfoLevel, false},
		{WarnLevel, false},
		{ErrorLevel, true},
		{FatalLevel, true},
	}

	for _, tt := range tests {
		entry := &Entry{Level: tt.level}
		hook.Run(entry)
	}

	if errorCount != 2 {
		t.Errorf("Expected 2 error level logs, got %d", errorCount)
	}
}
