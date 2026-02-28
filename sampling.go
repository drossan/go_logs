package go_logs

import (
	"io"
	"sync"
	"time"
)

// Sampler controls the rate of log messages to prevent flooding
type Sampler struct {
	// threshold is the maximum number of messages allowed per interval
	threshold int

	// interval is the time window for counting
	interval time.Duration

	// mu protects counter state
	mu sync.Mutex

	// count tracks messages in current interval
	count int

	// lastReset tracks when the interval started
	lastReset time.Time
}

// NewSampler creates a new Sampler with the specified threshold and interval.
//
// Example: NewSampler(1000, time.Minute) allows 1000 messages per minute
func NewSampler(threshold int, interval time.Duration) *Sampler {
	return &Sampler{
		threshold: threshold,
		interval:  interval,
		lastReset: time.Now(),
	}
}

// Allow returns true if the message should be logged (within rate limit)
// Returns false if the message should be dropped (rate limit exceeded)
func (s *Sampler) Allow() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Reset counter if interval has passed
	if now.Sub(s.lastReset) > s.interval {
		s.count = 0
		s.lastReset = now
	}

	// Check if within threshold
	if s.count < s.threshold {
		s.count++
		return true
	}

	return false
}

// AllowN allows N messages
func (s *Sampler) AllowN(n int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Reset counter if interval has passed
	if now.Sub(s.lastReset) > s.interval {
		s.count = 0
		s.lastReset = now
	}

	// Check if within threshold
	remaining := s.threshold - s.count
	if remaining <= 0 {
		return 0
	}

	allowed := n
	if allowed > remaining {
		allowed = remaining
	}

	s.count += allowed
	return allowed
}

// GetCount returns the current count
func (s *Sampler) GetCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count
}

// GetThreshold returns the configured threshold
func (s *Sampler) GetThreshold() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.threshold
}

// Reset resets the counter
func (s *Sampler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count = 0
	s.lastReset = time.Now()
}

// SamplingWriter wraps an io.Writer with sampling
type SamplingWriter struct {
	writer   io.Writer
	sampler  *Sampler
	onDrop   func(dropped int) // Called when messages are dropped
}

// NewSamplingWriter creates a writer with sampling
func NewSamplingWriter(writer io.Writer, threshold int, interval time.Duration) *SamplingWriter {
	return &SamplingWriter{
		writer:  writer,
		sampler: NewSampler(threshold, interval),
	}
}

// NewSamplingWriterWithCallback creates a writer with sampling and drop callback
func NewSamplingWriterWithCallback(writer io.Writer, threshold int, interval time.Duration, onDrop func(dropped int)) *SamplingWriter {
	return &SamplingWriter{
		writer:  writer,
		sampler: NewSampler(threshold, interval),
		onDrop:  onDrop,
	}
}

// Write implements io.Writer with sampling
func (sw *SamplingWriter) Write(p []byte) (n int, err error) {
	if sw.sampler.Allow() {
		return sw.writer.Write(p)
	}

	// Message dropped due to sampling
	if sw.onDrop != nil {
		sw.onDrop(1)
	}

	return len(p), nil
}

// Close closes the underlying writer if it implements io.Closer
func (sw *SamplingWriter) Close() error {
	if closer, ok := sw.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
