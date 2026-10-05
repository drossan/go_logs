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
		pkgName, fnName = packageFromFuncName(fn.Name())
	}

	return &CallerInfo{
		File:     filepath.Base(file),
		FileFull: file,
		Line:     line,
		Func:     fnName,
		Package:  pkgName,
	}
}

// packageFromFuncName splits a fully qualified runtime function name
// (as returned by runtime.Func.Name) into its package name and the function
// name relative to that package.
//
// Examples:
//
//	"github.com/user/pkg.Func"                      -> ("pkg", "Func")
//	"github.com/drossan/go_logs/v3.(*T).Log"        -> ("go_logs", "(*T).Log")
//	"github.com/drossan/go_logs/v3/async.(*L).With" -> ("async", "(*L).With")
//	"main.main"                                     -> ("main", "main")
//
// When the last path element is a module major-version suffix (vN with
// N >= 2, per Go's semantic import versioning), the package name is taken
// from the preceding path element. Generic instantiations such as
// "pkg.List[go.shape.int].Get" are handled by ignoring everything from the
// first '[' when locating the package path. If the name contains no '.',
// pkg is empty and fn is the last path element.
func packageFromFuncName(fnName string) (pkg, fn string) {
	path := fnName
	if idx := strings.IndexByte(path, '['); idx >= 0 {
		path = path[:idx]
	}
	lastSlash := strings.LastIndexByte(path, '/')
	rest := fnName[lastSlash+1:]

	dot := strings.IndexByte(rest, '.')
	if dot < 0 {
		return "", rest
	}
	pkg, fn = rest[:dot], rest[dot+1:]

	if lastSlash >= 0 && isMajorVersionSuffix(pkg) {
		prefix := fnName[:lastSlash]
		pkg = prefix[strings.LastIndexByte(prefix, '/')+1:]
	}
	return pkg, fn
}

// isMajorVersionSuffix reports whether s is a module major-version path
// element "vN" with N >= 2 and no leading zero (e.g. "v2", "v10", not "v1",
// "v01" or "v2tools").
func isMajorVersionSuffix(s string) bool {
	if len(s) < 2 || s[0] != 'v' || s[1] == '0' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != "v1"
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
