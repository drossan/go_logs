package go_logs

import (
	"errors"
	"testing"
)

// TestField_String verifies string field creation and access
func TestField_String(t *testing.T) {
	f := String("key", "value")

	if f.Key() != "key" {
		t.Errorf("String() field key = %v, want 'key'", f.Key())
	}
	if f.Type() != StringType {
		t.Errorf("String() field type = %v, want StringType", f.Type())
	}
	if f.StringValue() != "value" {
		t.Errorf("String() field value = %v, want 'value'", f.StringValue())
	}
}

// TestField_Int verifies int field creation and access
func TestField_Int(t *testing.T) {
	f := Int("count", 42)

	if f.Key() != "count" {
		t.Errorf("Int() field key = %v, want 'count'", f.Key())
	}
	if f.Type() != IntType {
		t.Errorf("Int() field type = %v, want IntType", f.Type())
	}
	if f.IntValue() != 42 {
		t.Errorf("Int() field value = %v, want 42", f.IntValue())
	}
}

// TestField_Int64 verifies int64 field creation and access
func TestField_Int64(t *testing.T) {
	f := Int64("bytes", 1536299520)

	if f.Key() != "bytes" {
		t.Errorf("Int64() field key = %v, want 'bytes'", f.Key())
	}
	if f.Type() != Int64Type {
		t.Errorf("Int64() field type = %v, want Int64Type", f.Type())
	}
	if f.Int64Value() != 1536299520 {
		t.Errorf("Int64() field value = %v, want 1536299520", f.Int64Value())
	}
}

// TestField_Float64 verifies float64 field creation and access
func TestField_Float64(t *testing.T) {
	f := Float64("cpu", 45.7)

	if f.Key() != "cpu" {
		t.Errorf("Float64() field key = %v, want 'cpu'", f.Key())
	}
	if f.Type() != Float64Type {
		t.Errorf("Float64() field type = %v, want Float64Type", f.Type())
	}
	if f.Float64Value() != 45.7 {
		t.Errorf("Float64() field value = %v, want 45.7", f.Float64Value())
	}
}

// TestField_Bool verifies bool field creation and access
func TestField_Bool(t *testing.T) {
	f := Bool("enabled", true)

	if f.Key() != "enabled" {
		t.Errorf("Bool() field key = %v, want 'enabled'", f.Key())
	}
	if f.Type() != BoolType {
		t.Errorf("Bool() field type = %v, want BoolType", f.Type())
	}
	if f.BoolValue() != true {
		t.Errorf("Bool() field value = %v, want true", f.BoolValue())
	}
}

// TestField_Err verifies error field creation and access
func TestField_Err(t *testing.T) {
	err := errors.New("test error")
	f := Err(err)

	if f.Key() != "error" {
		t.Errorf("Err() field key = %v, want 'error'", f.Key())
	}
	if f.Type() != ErrorType {
		t.Errorf("Err() field type = %v, want ErrorType", f.Type())
	}
	if f.ErrorValue() == nil {
		t.Errorf("Err() field value = nil, want error")
	}
	if f.ErrorValue().Error() != "test error" {
		t.Errorf("Err() field error = %v, want 'test error'", f.ErrorValue().Error())
	}
}

// TestField_Any verifies any field creation and access
func TestField_Any(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	user := User{Name: "John", Age: 30}
	f := Any("user", user)

	if f.Key() != "user" {
		t.Errorf("Any() field key = %v, want 'user'", f.Key())
	}
	if f.Type() != AnyType {
		t.Errorf("Any() field type = %v, want AnyType", f.Type())
	}
	if f.Value() == nil {
		t.Errorf("Any() field value = nil, want user struct")
	}
}

// TestField_TypeSafety verifies type-safe getters return defaults for wrong types
func TestField_TypeSafety(t *testing.T) {
	tests := []struct {
		name       string
		field      Field
		wantString string
		wantInt    int
		wantInt64  int64
		wantFloat  float64
		wantBool   bool
	}{
		{
			name:       "String field",
			field:      String("key", "value"),
			wantString: "value",
			wantInt:    0,
			wantInt64:  0,
			wantFloat:  0,
			wantBool:   false,
		},
		{
			name:       "Int field",
			field:      Int("key", 42),
			wantString: "",
			wantInt:    42,
			wantInt64:  0,
			wantFloat:  0,
			wantBool:   false,
		},
		{
			name:       "Bool field",
			field:      Bool("key", true),
			wantString: "",
			wantInt:    0,
			wantInt64:  0,
			wantFloat:  0,
			wantBool:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.field.StringValue(); got != tt.wantString {
				t.Errorf("StringValue() = %v, want %v", got, tt.wantString)
			}
			if got := tt.field.IntValue(); got != tt.wantInt {
				t.Errorf("IntValue() = %v, want %v", got, tt.wantInt)
			}
			if got := tt.field.Int64Value(); got != tt.wantInt64 {
				t.Errorf("Int64Value() = %v, want %v", got, tt.wantInt64)
			}
			if got := tt.field.Float64Value(); got != tt.wantFloat {
				t.Errorf("Float64Value() = %v, want %v", got, tt.wantFloat)
			}
			if got := tt.field.BoolValue(); got != tt.wantBool {
				t.Errorf("BoolValue() = %v, want %v", got, tt.wantBool)
			}
		})
	}
}

// TestField_NilError verifies Err() handles nil error gracefully
func TestField_NilError(t *testing.T) {
	f := Err(nil)

	if f.Key() != "error" {
		t.Errorf("Err(nil) key = %v, want 'error'", f.Key())
	}
	if f.Type() != ErrorType {
		t.Errorf("Err(nil) type = %v, want ErrorType", f.Type())
	}
	// nil error should store nil value
	if f.ErrorValue() != nil {
		t.Errorf("Err(nil) value = %v, want nil", f.ErrorValue())
	}
}

// BenchmarkField_StringAllocation measures allocations for String()
// CRITICAL: Should be 0 allocations
func BenchmarkField_StringAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = String("key", "value")
	}
}

// BenchmarkField_IntAllocation measures allocations for Int()
func BenchmarkField_IntAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Int("count", 42)
	}
}

// BenchmarkField_Multiple measures creating multiple fields
func BenchmarkField_Multiple(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = []Field{
			String("key1", "value1"),
			Int("key2", 42),
			Float64("key3", 45.7),
			Bool("key4", true),
		}
	}
}

// TestField_ZeroValue verifies zero-value Field behaves correctly
func TestField_ZeroValue(t *testing.T) {
	var f Field

	if f.Key() != "" {
		t.Errorf("zero-value field key = %v, want empty string", f.Key())
	}
	if f.Value() != nil {
		t.Errorf("zero-value field value = %v, want nil", f.Value())
	}
}

// TestFieldType_ConstValues verifies FieldType constants are unique
func TestFieldType_ConstValues(t *testing.T) {
	types := []FieldType{
		StringType,
		IntType,
		Int64Type,
		Float64Type,
		BoolType,
		ErrorType,
		AnyType,
	}

	// Verify all values are unique
	seen := make(map[FieldType]bool)
	for _, ft := range types {
		if seen[ft] {
			t.Errorf("Duplicate FieldType value: %v", ft)
		}
		seen[ft] = true
	}
}

// TestField_Getters verifies all getter methods work correctly
func TestField_Getters(t *testing.T) {
	f := String("test_key", "test_value")

	if got := f.Key(); got != "test_key" {
		t.Errorf("Key() = %v, want 'test_key'", got)
	}

	if got := f.Type(); got != StringType {
		t.Errorf("Type() = %v, want StringType", got)
	}

	if got := f.Value(); got != "test_value" {
		t.Errorf("Value() = %v, want 'test_value'", got)
	}
}

// TestField_BoolValues verifies both true and false bool values
func TestField_BoolValues(t *testing.T) {
	trueField := Bool("enabled", true)
	falseField := Bool("disabled", false)

	if !trueField.BoolValue() {
		t.Error("Bool(true) should return true")
	}

	if falseField.BoolValue() {
		t.Error("Bool(false) should return false")
	}
}

// TestField_NumericValues verifies various numeric values
func TestField_NumericValues(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		check func(Field) bool
	}{
		{
			name:  "Positive int",
			field: Int("pos", 42),
			check: func(f Field) bool { return f.IntValue() == 42 },
		},
		{
			name:  "Negative int",
			field: Int("neg", -10),
			check: func(f Field) bool { return f.IntValue() == -10 },
		},
		{
			name:  "Zero int",
			field: Int("zero", 0),
			check: func(f Field) bool { return f.IntValue() == 0 },
		},
		{
			name:  "Large int64",
			field: Int64("large", 9223372036854775807),
			check: func(f Field) bool { return f.Int64Value() == 9223372036854775807 },
		},
		{
			name:  "Small float64",
			field: Float64("small", 0.001),
			check: func(f Field) bool { return f.Float64Value() == 0.001 },
		},
		{
			name:  "Negative float64",
			field: Float64("negative", -123.456),
			check: func(f Field) bool { return f.Float64Value() == -123.456 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.field) {
				t.Errorf("Field value check failed for %s", tt.name)
			}
		})
	}
}
