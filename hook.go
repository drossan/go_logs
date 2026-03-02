package go_logs

// HookFunc is an adapter that allows ordinary functions to be used as hooks.
//
// This is useful for simple hooks that don't need to maintain state or
// implement the full Hook interface.
//
// Example:
//
//	hook := go_logs.HookFunc(func(entry *go_logs.Entry) error {
//	    fmt.Printf("Log: %s\n", entry.Message)
//	    return nil
//	})
type HookFunc func(entry *Entry) error

// Run implements Hook.Run for HookFunc.
func (f HookFunc) Run(entry *Entry) error {
	return f(entry)
}

// NewFuncHook creates a Hook from a function.
//
// This is a convenience constructor for HookFunc that provides better
// type inference and documentation.
//
// Parameters:
//
//	fn - A function that processes log entries
//
// Returns:
//
//	A Hook that calls the provided function
//
// Example:
//
//	hook := go_logs.NewFuncHook(func(entry *go_logs.Entry) error {
//	    // Collect metrics
//	    metrics.Inc("logs.total", 1)
//	    return nil
//	})
//
//	logger, _ := go_logs.New(
//	    go_logs.WithHooks(hook),
//	)
func NewFuncHook(fn func(entry *Entry) error) Hook {
	return HookFunc(fn)
}
