package go_logs

import (
	"sync/atomic"
)

// MetricsSnapshot represents an immutable snapshot of metrics at a point in time.
type MetricsSnapshot struct {
	Total   int64
	ByLevel map[Level]int64
}

// Metrics collects statistics about logging activity.
// It provides thread-safe counters for monitoring logging behavior in production.
//
// Metrics are always enabled (zero overhead) and include:
//   - Total log count
//   - Count by level (trace, debug, info, warn, error, fatal)
//   - Dropped logs (when async buffer is full)
//
// Example:
//
//	logger, _ := go_logs.New()
//	metrics := logger.GetMetrics()
//	fmt.Printf("Total logs: %d\n", metrics.Total())
//	fmt.Printf("Errors: %d\n", metrics.Count(go_logs.ErrorLevel))
type Metrics struct {
	// total is the total number of log entries processed
	total atomic.Int64

	// byLevel stores counts per log level using atomic operations
	// Index: 0=Trace, 1=Debug, 2=Info, 3=Warn, 4=Error, 5=Fatal
	byLevel [6]atomic.Int64

	// dropped counts logs dropped due to async buffer overflow
	dropped atomic.Int64
}

// NewMetrics creates a new Metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// Increment atomically increments the counter for the given level.
func (m *Metrics) Increment(level Level) {
	m.total.Add(1)
	if idx := levelToIndex(level); idx >= 0 && idx < 6 {
		m.byLevel[idx].Add(1)
	}
}

// Count returns the count of logs at the given level.
func (m *Metrics) Count(level Level) int64 {
	if idx := levelToIndex(level); idx >= 0 && idx < 6 {
		return m.byLevel[idx].Load()
	}
	return 0
}

// Total returns the total number of logs across all levels.
func (m *Metrics) Total() int64 {
	return m.total.Load()
}

// Dropped returns the number of logs dropped due to buffer overflow.
// This is only relevant when using async logging.
func (m *Metrics) Dropped() int64 {
	return m.dropped.Load()
}

// IncrementDropped increments the dropped counter.
// This is called by async logging when the buffer is full.
func (m *Metrics) IncrementDropped() {
	m.dropped.Add(1)
}

// Reset resets all counters to zero.
// This is useful for testing or periodic metrics collection.
func (m *Metrics) Reset() {
	m.total.Store(0)
	m.dropped.Store(0)
	for i := range m.byLevel {
		m.byLevel[i].Store(0)
	}
}

// Snapshot returns an immutable copy of the current metrics.
// This is useful for capturing metrics at a point in time.
func (m *Metrics) Snapshot() MetricsSnapshot {
	byLevel := make(map[Level]int64)
	levels := []Level{TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel}
	for _, level := range levels {
		byLevel[level] = m.Count(level)
	}
	return MetricsSnapshot{
		Total:   m.Total(),
		ByLevel: byLevel,
	}
}

// levelToIndex converts a Level to an array index.
// Returns -1 for invalid levels.
func levelToIndex(level Level) int {
	switch level {
	case TraceLevel:
		return 0
	case DebugLevel:
		return 1
	case InfoLevel:
		return 2
	case WarnLevel:
		return 3
	case ErrorLevel:
		return 4
	case FatalLevel:
		return 5
	default:
		return -1
	}
}

// MetricsGetter is an interface for getting metrics from a logger.
type MetricsGetter interface {
	GetMetrics() *Metrics
}
