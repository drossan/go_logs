package hooks

import (
	"fmt"
	"log/syslog"
	"sync"

	"github.com/drossan/go_logs"
)

// SyslogHook sends logs to the local syslog daemon
type SyslogHook struct {
	writer    *syslog.Writer
	minLevel  go_logs.Level
	mu        sync.RWMutex
	formatter func(entry *go_logs.Entry) string
}

// SyslogConfig holds configuration for the syslog hook
type SyslogConfig struct {
	// Network and address for remote syslog (e.g., "tcp", "logs.example.com:514")
	Network string
	Addr    string

	// Priority is the syslog priority (facility | severity)
	Priority syslog.Priority

	// Tag is the syslog tag
	Tag string

	// Flags for the syslog writer
	Flags int
}

// NewSyslogHook creates a new syslog hook with default configuration
func NewSyslogHook(tag string) (*SyslogHook, error) {
	return NewSyslogHookWithConfig(SyslogConfig{
		Priority: syslog.LOG_LOCAL0 | syslog.LOG_INFO,
		Tag:      tag,
	})
}

// NewSyslogHookWithConfig creates a new syslog hook with custom configuration
func NewSyslogHookWithConfig(config SyslogConfig) (*SyslogHook, error) {
	var writer *syslog.Writer
	var err error

	if config.Network != "" && config.Addr != "" {
		// Remote syslog
		writer, err = syslog.Dial(config.Network, config.Addr, config.Priority, config.Tag)
	} else {
		// Local syslog
		writer, err = syslog.New(config.Priority, config.Tag)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create syslog writer: %w", err)
	}

	return &SyslogHook{
		writer:    writer,
		minLevel:  go_logs.InfoLevel,
		formatter: defaultSyslogFormatter,
	}, nil
}

// defaultSyslogFormatter formats an entry for syslog
func defaultSyslogFormatter(entry *go_logs.Entry) string {
	msg := entry.Message

	// Add fields
	for _, f := range entry.Fields {
		msg += fmt.Sprintf(" %s=%v", f.Key(), f.Value())
	}

	// Add caller info if present
	if entry.Caller != nil {
		msg += fmt.Sprintf(" caller=%s:%d", entry.Caller.File, entry.Caller.Line)
	}

	return msg
}

// Run implements go_logs.Hook interface
func (h *SyslogHook) Run(entry *go_logs.Entry) error {
	h.mu.RLock()
	minLevel := h.minLevel
	h.mu.RUnlock()

	if entry.Level < minLevel {
		return nil
	}

	msg := h.formatter(entry)

	// Send to syslog based on level
	switch entry.Level {
	case go_logs.TraceLevel, go_logs.DebugLevel:
		return h.writer.Debug(msg)
	case go_logs.InfoLevel:
		return h.writer.Info(msg)
	case go_logs.WarnLevel:
		return h.writer.Warning(msg)
	case go_logs.ErrorLevel:
		return h.writer.Err(msg)
	case go_logs.FatalLevel:
		return h.writer.Crit(msg)
	default:
		return h.writer.Info(msg)
	}
}

// SetLevel sets the minimum level for this hook
func (h *SyslogHook) SetLevel(level go_logs.Level) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.minLevel = level
}

// SetFormatter sets a custom formatter
func (h *SyslogHook) SetFormatter(fn func(entry *go_logs.Entry) string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.formatter = fn
}

// Close closes the syslog writer
func (h *SyslogHook) Close() error {
	return h.writer.Close()
}

// NetworkSyslogHook sends logs to a remote syslog server
type NetworkSyslogHook struct {
	*SyslogHook
}

// NewNetworkSyslogHook creates a hook that sends to a remote syslog server
func NewNetworkSyslogHook(network, addr, tag string) (*NetworkSyslogHook, error) {
	hook, err := NewSyslogHookWithConfig(SyslogConfig{
		Network:  network,
		Addr:     addr,
		Priority: syslog.LOG_LOCAL0 | syslog.LOG_INFO,
		Tag:      tag,
	})
	if err != nil {
		return nil, err
	}
	return &NetworkSyslogHook{SyslogHook: hook}, nil
}

// RFC5424Formatter formats logs according to RFC5424
func RFC5424Formatter(appName string) func(entry *go_logs.Entry) string {
	return func(entry *go_logs.Entry) string {
		// Structured data format
		sd := fmt.Sprintf(`[meta@12345 app="%s" level="%s"`, appName, entry.Level.String())

		if entry.Caller != nil {
			sd += fmt.Sprintf(` file="%s" line="%d" func="%s"`,
				entry.Caller.File, entry.Caller.Line, entry.Caller.Func)
		}

		for _, f := range entry.Fields {
			sd += fmt.Sprintf(` %s="%v"`, f.Key(), f.Value())
		}

		sd += "]"

		return sd + " " + entry.Message
	}
}
