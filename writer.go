package go_logs

import "errors"

// Flusher is implemented by outputs that buffer data in user space.
//
// The logger calls Flush AFTER EVERY ENTRY, so no line is held in memory
// between writes and a crash does not lose already-logged entries still
// sitting in a buffer. Flush must not fsync: forcing data to stable storage
// is the job of Sync, which the logger only calls from Logger.Sync. A writer
// that does not want a per-entry flush simply does not implement Flusher.
//
// Errors returned by Flush during logging are ignored: a failing output must
// not break or block the caller of Log.
//
// RotatingFileWriter, EnhancedRotatingFileWriter, MultiWriter and
// SamplingWriter implement Flusher; *bufio.Writer does too.
type Flusher interface {
	Flush() error
}

// isIgnorableSyncErr reports whether err comes only from syncing an output
// that cannot be synced, such as a terminal, a pipe or a wrapped os.Stdout:
// the errno values EINVAL, ENOTTY and EBADF. They are matched by errno with
// errors.Is (os.File.Sync wraps them in *os.PathError), not by the identity of
// the writer, so stdout wrapped by another writer is covered too.
//
// For an error that joins several errors (errors.Join, as returned by
// MultiWriter.Sync) it reports true only if every joined error is ignorable,
// so a real failure of one writer is never hidden by a terminal of another.
// A nil error is not ignorable: there is nothing to ignore.
func isIgnorableSyncErr(err error) bool {
	if err == nil {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		errs := joined.Unwrap()
		if len(errs) == 0 {
			return false
		}
		for _, e := range errs {
			if !isIgnorableSyncErr(e) {
				return false
			}
		}
		return true
	}
	for _, errno := range ignorableSyncErrnos {
		if errors.Is(err, errno) {
			return true
		}
	}
	return false
}
