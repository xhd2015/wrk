package wrkcli

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
)

// Package-level CLI I/O used by product print paths. When a RunWithOptions
// call installs custom writers (tests), these point at the capture buffers for
// the duration of that call. Default is process stdout/stderr.
//
// Custom-writer / env-isolated runs take captureMu for the whole invocation so
// concurrent in-process tests do not interleave capture or Setenv.

var (
	captureMu     sync.Mutex
	activeStdout  atomic.Pointer[io.Writer]
	activeStderr  atomic.Pointer[io.Writer]
)

// cliStdout returns the active CLI stdout writer (test capture or os.Stdout).
func cliStdout() io.Writer {
	if p := activeStdout.Load(); p != nil && *p != nil {
		return *p
	}
	return os.Stdout
}

// cliStderr returns the active CLI stderr writer (test capture or os.Stderr).
func cliStderr() io.Writer {
	if p := activeStderr.Load(); p != nil && *p != nil {
		return *p
	}
	return os.Stderr
}

func installCLIWriters(stdout, stderr io.Writer) (restore func()) {
	// Heap-allocate so atomic pointers stay valid for the call duration.
	var outHold, errHold io.Writer
	if stdout != nil {
		outHold = stdout
		activeStdout.Store(&outHold)
	}
	if stderr != nil {
		errHold = stderr
		activeStderr.Store(&errHold)
	}
	return func() {
		if stdout != nil {
			activeStdout.Store(nil)
		}
		if stderr != nil {
			activeStderr.Store(nil)
		}
	}
}
