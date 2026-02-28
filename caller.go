package go_logs

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

// CallerInfo holds information about the calling function
type CallerInfo struct {
	File     string // Short file name (e.g., "main.go")
	FileFull string // Full file path
	Line     int    // Line number
	Func     string // Function name (e.g., "main.myFunction")
	Package  string // Package name
}

// GetCaller retrieves caller information by skipping a number of frames.
// skip=0 returns the caller of GetCaller, skip=1 returns its caller, etc.
func GetCaller(skip int) *CallerInfo {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return nil
	}

	// Extract function and package name
	fn := runtime.FuncForPC(pc)
	var fnName, pkgName string
	if fn != nil {
		fnName = fn.Name()
		// Extract package from function name (e.g., "github.com/user/pkg.Func" -> "pkg")
		if idx := strings.LastIndex(fnName, "/"); idx >= 0 {
			fnName = fnName[idx+1:]
		}
		if idx := strings.Index(fnName, "."); idx >= 0 {
			pkgName = fnName[:idx]
			fnName = fnName[idx+1:]
		}
	}

	return &CallerInfo{
		File:     filepath.Base(file),
		FileFull: file,
		Line:     line,
		Func:     fnName,
		Package:  pkgName,
	}
}

// String returns a formatted string representation (file:line)
func (c *CallerInfo) String() string {
	if c == nil {
		return "???"
	}
	return fmt.Sprintf("%s:%d", c.File, c.Line)
}

// FullString returns a detailed string representation (file:line function)
func (c *CallerInfo) FullString() string {
	if c == nil {
		return "???"
	}
	return fmt.Sprintf("%s:%d %s", c.File, c.Line, c.Func)
}

// GetStackTrace captures the current stack trace
func GetStackTrace(skip int) []byte {
	// Allocate buffer for stack trace (4KB should be enough for most cases)
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	if n == 0 {
		return nil
	}
	return buf[:n]
}

// GetStackTraceAll captures stack trace of all goroutines
func GetStackTraceAll() []byte {
	buf := make([]byte, 65536) // 64KB for all goroutines
	n := runtime.Stack(buf, true)
	if n == 0 {
		return nil
	}
	return buf[:n]
}
