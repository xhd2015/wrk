package unwind

import (
	"io"
	"os"
)

// HostIO carries optional stdout/stderr for host-backed apply ops.
// Nil fields default to os.Stdout / os.Stderr so non-TUI call sites stay unchanged.
type HostIO struct {
	Stdout io.Writer
	Stderr io.Writer
}

// Out returns Stdout or os.Stdout.
func (h HostIO) Out() io.Writer {
	if h.Stdout != nil {
		return h.Stdout
	}
	return os.Stdout
}

// Err returns Stderr or os.Stderr.
func (h HostIO) Err() io.Writer {
	if h.Stderr != nil {
		return h.Stderr
	}
	return os.Stderr
}
