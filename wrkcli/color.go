package wrkcli

import (
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiOrange = "\x1b[33m"
	ansiGrey   = "\x1b[90m"
	ansiReset  = "\x1b[0m"
)

func colorize(s, code string) string {
	return code + s + ansiReset
}

// forceStderrColor is set when --color is passed (forces ANSI even on non-TTY).
var forceStderrColor bool

// SetForceStderrColor records --color for top-level Error:/warning: printing.
func SetForceStderrColor(v bool) {
	forceStderrColor = v
}

// stderrColorEnabled reports whether stderr diagnostic prefixes should use ANSI.
// --color forces on; otherwise on only when stderr is a TTY and NO_COLOR is empty.
func stderrColorEnabled() bool {
	if forceStderrColor {
		return true
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}

// FormatStderrError colors a leading "Error:" prefix red when color is on.
// Prefer prefix-only coloring so the body stays plain.
func FormatStderrError(msg string) string {
	if !stderrColorEnabled() {
		return msg
	}
	return colorDiagnosticPrefix(msg, "Error:", ansiRed)
}

// FormatStderrWarning colors a leading "warning:" prefix yellow/orange when color is on.
func FormatStderrWarning(msg string) string {
	if !stderrColorEnabled() {
		return msg
	}
	return colorDiagnosticPrefix(msg, "warning:", ansiOrange)
}

// colorDiagnosticPrefix colorizes prefix if msg starts with it (after optional wrk: ).
func colorDiagnosticPrefix(msg, prefix, code string) string {
	if strings.HasPrefix(msg, prefix) {
		return colorize(prefix, code) + msg[len(prefix):]
	}
	// Allow "wrk: Error: …" style frames.
	const wrk = "wrk: "
	if strings.HasPrefix(msg, wrk) && strings.HasPrefix(msg[len(wrk):], prefix) {
		rest := msg[len(wrk):]
		return wrk + colorize(prefix, code) + rest[len(prefix):]
	}
	return msg
}