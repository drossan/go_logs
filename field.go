package go_logs

// Field represents a structured key-value pair for logging.
// Fields are designed to be zero-allocation - they are simple structs
// that can be created and passed around efficiently.
//
// The valueType allows the formatter to handle different types appropriately
// (e.g., quoting strings, formatting numbers, handling errors specially).
type Field struct {
	key       string
	valueType FieldType
	value     interface{}
}

// FieldType represents the type of value stored in a Field.
// This enables type-specific formatting in formatters.
type FieldType int

const (
	// StringType represents a string value
	StringType FieldType = iota

	// IntType represents an int value
	IntType

	// Int64Type represents an int64 value
	Int64Type

	// Float64Type represents a float64 value
	Float64Type

	// BoolType represents a bool value
	BoolType

	// ErrorType represents an error value (formatted specially)
	ErrorType

	// AnyType represents an arbitrary interface{} value
	AnyType
)

// String creates a Field with a string value.
// This is the most common field type for log messages.
//
// Example:
//
//	logger.Info("User logged in",
//	    go_logs.String("username", "john"),
//	    go_logs.String("ip", "192.168.1.1"))
func String(key, val string) Field {
	return Field{key: key, valueType: StringType, value: val}
}

// Int creates a Field with an int value.
// Useful for numeric data like counts, IDs, ports, etc.
//
// Example:
//
//	logger.Info("Request processed",
//	    go_logs.Int("status_code", 200),
//	    go_logs.Int("response_time_ms", 45))
func Int(key string, val int) Field {
	return Field{key: key, valueType: IntType, value: val}
}

// Int64 creates a Field with an int64 value.
// Use this for large integers or when you need explicit 64-bit precision.
//
// Example:
//
//	logger.Info("File size",
//	    go_logs.Int64("bytes", 1536299520))
func Int64(key string, val int64) Field {
	return Field{key: key, valueType: Int64Type, value: val}
}

// Float64 creates a Field with a float64 value.
// Useful for measurements, percentages, etc.
//
// Example:
//
//	logger.Info("Memory usage",
//	    go_logs.Float64("cpu_percent", 45.7),
//	    go_logs.Float64("memory_gb", 1.23))
func Float64(key string, val float64) Field {
	return Field{key: key, valueType: Float64Type, value: val}
}

// Bool creates a Field with a bool value.
// Useful for flags and boolean state.
//
// Example:
//
//	logger.Info("Feature flags",
//	    go_logs.Bool("debug_mode", true),
//	    go_logs.Bool("cache_enabled", false))
func Bool(key string, val bool) Field {
	return Field{key: key, valueType: BoolType, value: val}
}

// Err creates a Field with an error value.
// The key is automatically set to "error" for consistency.
// Errors are formatted specially by formatters.
//
// Example:
//
//	logger.Error("Database connection failed",
//	    go_logs.Err(err),
//	    go_logs.String("host", "localhost"))
func Err(err error) Field {
	return Field{key: "error", valueType: ErrorType, value: err}
}

// Any creates a Field with an arbitrary value.
// Use this when the type doesn't match the typed helpers.
// The formatter will use fmt.Sprintf("%v") to format the value.
//
// Example:
//
//	logger.Info("Complex data",
//	    go_logs.Any("user", User{Name: "John", Age: 30}))
func Any(key string, val interface{}) Field {
	return Field{key: key, valueType: AnyType, value: val}
}

// Key returns the field's key.
// This is useful for formatters and hooks that need to inspect fields.
func (f Field) Key() string {
	return f.key
}

// Type returns the field's value type.
// This is useful for type-specific formatting.
func (f Field) Type() FieldType {
	return f.valueType
}

// Value returns the field's value.
// This is useful for formatters and hooks that need to access the value.
func (f Field) Value() interface{} {
	return f.value
}

// String returns the string representation of the field's value.
// For non-string types, it uses fmt.Sprintf("%v").
func (f Field) StringValue() string {
	if f.valueType == StringType {
		if str, ok := f.value.(string); ok {
			return str
		}
	}
	return ""
}

// IntValue returns the int value.
// Returns 0 if the field is not an IntType.
func (f Field) IntValue() int {
	if f.valueType == IntType {
		if i, ok := f.value.(int); ok {
			return i
		}
	}
	return 0
}

// Int64Value returns the int64 value.
// Returns 0 if the field is not an Int64Type.
func (f Field) Int64Value() int64 {
	if f.valueType == Int64Type {
		if i, ok := f.value.(int64); ok {
			return i
		}
	}
	return 0
}

// Float64Value returns the float64 value.
// Returns 0 if the field is not a Float64Type.
func (f Field) Float64Value() float64 {
	if f.valueType == Float64Type {
		if f, ok := f.value.(float64); ok {
			return f
		}
	}
	return 0
}

// BoolValue returns the bool value.
// Returns false if the field is not a BoolType.
func (f Field) BoolValue() bool {
	if f.valueType == BoolType {
		if b, ok := f.value.(bool); ok {
			return b
		}
	}
	return false
}

// ErrorValue returns the error value.
// Returns nil if the field is not an ErrorType.
func (f Field) ErrorValue() error {
	if f.valueType == ErrorType {
		if err, ok := f.value.(error); ok {
			return err
		}
	}
	return nil
}
