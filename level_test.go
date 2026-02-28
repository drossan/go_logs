package go_logs

import (
	"strings"
	"testing"
)

// TestLevel_String verifies the String() method returns correct representations
func TestLevel_String(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		expected string
	}{
		{"Trace", TraceLevel, "TRACE"},
		{"Debug", DebugLevel, "DEBUG"},
		{"Info", InfoLevel, "INFO"},
		{"Warn", WarnLevel, "WARN"},
		{"Error", ErrorLevel, "ERROR"},
		{"Fatal", FatalLevel, "FATAL"},
		{"Success", SuccessLevel, "SUCCESS"},
		{"Silent", SilentLevel, "SILENT"},
		{"Unknown", Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.expected {
				t.Errorf("Level.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestLevel_ShouldLog verifies fast-path filtering behavior
func TestLevel_ShouldLog(t *testing.T) {
	tests := []struct {
		name      string
		level     Level
		threshold Level
		expected  bool
	}{
		{"Trace logs at Trace", TraceLevel, TraceLevel, true},
		{"Debug logs at Debug", DebugLevel, DebugLevel, true},
		{"Info logs at Info", InfoLevel, InfoLevel, true},
		{"Warn logs at Warn", WarnLevel, WarnLevel, true},
		{"Error logs at Error", ErrorLevel, ErrorLevel, true},
		{"Fatal logs at Fatal", FatalLevel, FatalLevel, true},

		{"Debug filtered at Info", DebugLevel, InfoLevel, false},
		{"Info logs at Info threshold", InfoLevel, InfoLevel, true},
		{"Warn logs at Info threshold", WarnLevel, InfoLevel, true},
		{"Error logs at Info threshold", ErrorLevel, InfoLevel, true},

		{"Silent filters everything", InfoLevel, SilentLevel, true},
		{"Silent level doesn't log", SilentLevel, InfoLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.shouldLog(tt.threshold)
			if got != tt.expected {
				t.Errorf("Level.shouldLog(%v) = %v, want %v",
					tt.threshold, got, tt.expected)
			}
		})
	}
}

// TestParseLevel verifies string to Level conversion
func TestParseLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Level
	}{
		{"TRACE uppercase", "TRACE", TraceLevel},
		{"TRACE lowercase", "trace", TraceLevel},
		{"DEBUG uppercase", "DEBUG", DebugLevel},
		{"DEBUG lowercase", "debug", DebugLevel},
		{"INFO uppercase", "INFO", InfoLevel},
		{"INFO lowercase", "info", InfoLevel},
		{"WARN uppercase", "WARN", WarnLevel},
		{"WARN lowercase", "warn", WarnLevel},
		{"WARNING uppercase", "WARNING", WarnLevel},
		{"WARNING lowercase", "warning", WarnLevel},
		{"ERROR uppercase", "ERROR", ErrorLevel},
		{"ERROR lowercase", "error", ErrorLevel},
		{"FATAL uppercase", "FATAL", FatalLevel},
		{"FATAL lowercase", "fatal", FatalLevel},
		{"SUCCESS uppercase", "SUCCESS", SuccessLevel},
		{"SUCCESS lowercase", "success", SuccessLevel},
		{"SILENT uppercase", "SILENT", SilentLevel},
		{"SILENT lowercase", "silent", SilentLevel},
		{"NONE", "none", SilentLevel},
		{"DISABLE", "disable", SilentLevel},
		{"Unknown defaults to Info", "unknown", InfoLevel},
		{"Empty defaults to Info", "", InfoLevel},
		{"Mixed case", "InFo", InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// BenchmarkLevel_ShouldLog measures fast-path filtering performance
// CRITICAL: Must be < 5ns per call
func BenchmarkLevel_ShouldLog(b *testing.B) {
	threshold := InfoLevel
	level := DebugLevel

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = level.shouldLog(threshold)
	}
}

// BenchmarkParseLevel measures level parsing performance
func BenchmarkParseLevel(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseLevel("info")
	}
}

// BenchmarkLevel_String measures string conversion performance
func BenchmarkLevel_String(b *testing.B) {
	level := InfoLevel
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = level.String()
	}
}

// TestLevelNumericValues verifies syslog-style numeric levels
func TestLevelNumericValues(t *testing.T) {
	tests := []struct {
		level    Level
		expected int
	}{
		{SilentLevel, 0},
		{TraceLevel, 10},
		{DebugLevel, 20},
		{SuccessLevel, 25},
		{InfoLevel, 30},
		{WarnLevel, 40},
		{ErrorLevel, 50},
		{FatalLevel, 60},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if int(tt.level) != tt.expected {
				t.Errorf("%s = %d, want %d", tt.level.String(), int(tt.level), tt.expected)
			}
		})
	}
}

// TestLevelOrdering verifies levels are properly ordered
func TestLevelOrdering(t *testing.T) {
	// Verify ascending order
	levels := []Level{
		SilentLevel,
		TraceLevel,
		DebugLevel,
		SuccessLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel,
		FatalLevel,
	}

	for i := 1; i < len(levels); i++ {
		if levels[i] <= levels[i-1] {
			t.Errorf("Level ordering broken: %s (%d) <= %s (%d)",
				levels[i].String(), int(levels[i]),
				levels[i-1].String(), int(levels[i-1]))
		}
	}
}

// TestParseLevel_CaseInsensitive verifies case insensitivity
func TestParseLevel_CaseInsensitive(t *testing.T) {
	variations := []string{
		"TRACE", "Trace", "trace", "TrAcE",
		"DEBUG", "Debug", "debug", "DeBuG",
		"INFO", "Info", "info", "InFo",
	}

	for _, variation := range variations {
		level := strings.ToLower(variation)
		parsed := ParseLevel(variation)
		expected := ParseLevel(level)

		if parsed != expected {
			t.Errorf("ParseLevel(%q) = %v, ParseLevel(%q) = %v (should be equal)",
				variation, parsed, level, expected)
		}
	}
}
