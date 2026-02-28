package go_logs

import (
	"context"
)

// contextKey is a custom type for context keys to prevent collisions.
// Using an unexported type ensures that other packages cannot create
// conflicting context keys.
type contextKey struct {
	name string
}

// Context keys for storing values in context.Context.
// These are unexported to prevent collisions with other packages.
// Each key has a unique name to prevent collisions even within this package.
var (
	traceIDKey   = contextKey{name: "trace_id"}
	spanIDKey    = contextKey{name: "span_id"}
	requestIDKey = contextKey{name: "request_id"}
	userIDKey    = contextKey{name: "user_id"}
)

// WithTraceID adds a trace ID to the context.
//
// Trace IDs are used in distributed tracing to correlate log entries
// across multiple services. The trace ID should be the same for all
// log entries within a single request/transaction flow.
//
// Parameters:
//   ctx     - The parent context
//   traceID - The trace ID to associate with this context
//
// Returns:
//   A new context with the trace ID stored
//
// Example:
//
//	ctx := go_logs.WithTraceID(context.Background(), "abc-123-def-456")
//	logger.LogCtx(ctx, go_logs.InfoLevel, "Request received")
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// WithSpanID adds a span ID to the context.
//
// Span IDs represent individual operations within a trace. A single trace
// can have multiple spans (e.g., HTTP request, database query, external API call).
//
// Parameters:
//   ctx    - The parent context
//   spanID - The span ID to associate with this context
//
// Returns:
//   A new context with the span ID stored
//
// Example:
//
//	ctx := go_logs.WithSpanID(context.Background(), "span-xyz")
//	logger.LogCtx(ctx, go_logs.InfoLevel, "Database query executed")
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

// WithRequestID adds a request ID to the context.
//
// Request IDs are used to track individual requests through a system.
// Unlike trace IDs (which are distributed), request IDs are typically
// scoped to a single service or application.
//
// Parameters:
//   ctx       - The parent context
//   requestID - The request ID to associate with this context
//
// Returns:
//   A new context with the request ID stored
//
// Example:
//
//	ctx := go_logs.WithRequestID(context.Background(), "req-123-456")
//	logger.LogCtx(ctx, go_logs.InfoLevel, "Processing request")
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithUserID adds a user ID to the context.
//
// User IDs are useful for associating log entries with specific users,
// which is valuable for auditing and user-specific debugging.
//
// Parameters:
//   ctx    - The parent context
//   userID - The user ID to associate with this context
//
// Returns:
//   A new context with the user ID stored
//
// Example:
//
//	ctx := go_logs.WithUserID(context.Background(), "user-789")
//	logger.LogCtx(ctx, go_logs.InfoLevel, "User action performed")
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetTraceID extracts the trace ID from the context.
//
// If no trace ID is set in the context, returns an empty string.
//
// Parameters:
//   ctx - The context to extract from
//
// Returns:
//   The trace ID if set, or empty string
//
// Example:
//
//	traceID := go_logs.GetTraceID(ctx)
//	if traceID != "" {
//	    logger.Info("Processing with trace", go_logs.String("trace_id", traceID))
//	}
func GetTraceID(ctx context.Context) string {
	if v := ctx.Value(traceIDKey); v != nil {
		if traceID, ok := v.(string); ok {
			return traceID
		}
	}
	return ""
}

// GetSpanID extracts the span ID from the context.
//
// If no span ID is set in the context, returns an empty string.
//
// Parameters:
//   ctx - The context to extract from
//
// Returns:
//   The span ID if set, or empty string
func GetSpanID(ctx context.Context) string {
	if v := ctx.Value(spanIDKey); v != nil {
		if spanID, ok := v.(string); ok {
			return spanID
		}
	}
	return ""
}

// GetRequestID extracts the request ID from the context.
//
// If no request ID is set in the context, returns an empty string.
//
// Parameters:
//   ctx - The context to extract from
//
// Returns:
//   The request ID if set, or empty string
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
}

// GetUserID extracts the user ID from the context.
//
// If no user ID is set in the context, returns an empty string.
//
// Parameters:
//   ctx - The context to extract from
//
// Returns:
//   The user ID if set, or empty string
func GetUserID(ctx context.Context) string {
	if v := ctx.Value(userIDKey); v != nil {
		if userID, ok := v.(string); ok {
			return userID
		}
	}
	return ""
}

// ExtractFieldsFromContext extracts structured fields from the context.
//
// This function automatically extracts trace_id, span_id, request_id, and user_id
// from the context and returns them as Field structs. This is used internally
// by LogCtx to automatically include context information in log entries.
//
// Parameters:
//   ctx - The context to extract from
//
// Returns:
//   A slice of Field structs containing the extracted context values
//
// Example:
//
//	fields := go_logs.ExtractFieldsFromContext(ctx)
//	// fields may contain: String("trace_id", "..."), String("span_id", "..."), etc.
func ExtractFieldsFromContext(ctx context.Context) []Field {
	var fields []Field

	if traceID := GetTraceID(ctx); traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}

	if spanID := GetSpanID(ctx); spanID != "" {
		fields = append(fields, String("span_id", spanID))
	}

	if requestID := GetRequestID(ctx); requestID != "" {
		fields = append(fields, String("request_id", requestID))
	}

	if userID := GetUserID(ctx); userID != "" {
		fields = append(fields, String("user_id", userID))
	}

	return fields
}
