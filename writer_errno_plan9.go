//go:build plan9

package go_logs

import "syscall"

// ignorableSyncErrnos are the errors that syncing a terminal, a pipe or a
// closed descriptor returns; see isIgnorableSyncErr. Plan 9 defines no
// ENOTTY or EBADF in package syscall.
var ignorableSyncErrnos = []error{syscall.EINVAL}
