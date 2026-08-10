package wrkcli

import (
	"fmt"
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

// paint wraps s with ANSI code when on is true; otherwise returns s unchanged.
func paint(s, code string, on bool) string {
	if !on || s == "" {
		return s
	}
	return colorize(s, code)
}

// resolveStdoutColor is the go-best-practice three-mode policy for stdout:
// --no-color always off; --color always on; else TTY + empty NO_COLOR.
// Callers must reject both --color and --no-color before calling.
func resolveStdoutColor(forceColor, noColor bool) bool {
	if noColor {
		return false
	}
	if forceColor {
		return true
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// paintCount formats n for summary lines: green when n>0 and color on, else plain/gray-0.
func paintCount(n int, colorOn bool) string {
	s := fmt.Sprintf("%d", n)
	if !colorOn {
		return s
	}
	if n > 0 {
		return colorize(s, ansiGreen)
	}
	return colorize(s, ansiGrey)
}

// paintSkippedCount is orange when n>0 (soft skips), grey when 0, plain when color off.
func paintSkippedCount(n int, colorOn bool) string {
	s := fmt.Sprintf("%d", n)
	if !colorOn {
		return s
	}
	if n > 0 {
		return colorize(s, ansiOrange)
	}
	return colorize(s, ansiGrey)
}

// paintFailedCount is red when n>0, grey when 0, plain when color off.
func paintFailedCount(n int, colorOn bool) string {
	s := fmt.Sprintf("%d", n)
	if !colorOn {
		return s
	}
	if n > 0 {
		return colorize(s, ansiRed)
	}
	return colorize(s, ansiGrey)
}

// formatSyncSummaryLine builds "synced: N into main, M into worktrees, K skipped".
// Labels grey when color on; success counts green when >0; skipped orange when >0.
func formatSyncSummaryLine(intoMain, intoWT, skipped int, colorOn bool) string {
	label := "synced:"
	intoMainL, intoWTL, skippedL := "into main,", "into worktrees,", "skipped"
	if colorOn {
		label = colorize(label, ansiGrey)
		intoMainL = colorize(intoMainL, ansiGrey)
		intoWTL = colorize(intoWTL, ansiGrey)
		skippedL = colorize(skippedL, ansiGrey)
	}
	return fmt.Sprintf("%s %s %s %s %s %s %s",
		label,
		paintCount(intoMain, colorOn), intoMainL,
		paintCount(intoWT, colorOn), intoWTL,
		paintSkippedCount(skipped, colorOn), skippedL,
	)
}

// formatReinstallSummaryLine builds "reinstalled N, skipped M, failed F".
func formatReinstallSummaryLine(reinstalled, skipped, failed int, colorOn bool) string {
	reL, skL, faL := "reinstalled", "skipped", "failed"
	if colorOn {
		reL = colorize(reL, ansiGrey)
		skL = colorize(skL, ansiGrey)
		faL = colorize(faL, ansiGrey)
	}
	return fmt.Sprintf("%s %s, %s %s, %s %s",
		reL, paintCount(reinstalled, colorOn),
		skL, paintSkippedCount(skipped, colorOn),
		faL, paintFailedCount(failed, colorOn),
	)
}

// formatGoInstallProgressLine highlights the go install/run verb (green) and
// leaves the package path plain — progress for --reinstall-local.
func formatGoInstallProgressLine(method Method, relPath string, colorOn bool) string {
	verb := "go install"
	if method == MethodGoRunInstall {
		verb = "go run"
	}
	if colorOn {
		verb = colorize(verb, ansiGreen)
	}
	return verb + " " + relPath
}

// formatUnwindSummaryLine builds the end-of-apply rollup for --unwind.
// Only includes stages that were requested / performed (omit unused).
func formatUnwindSummaryLine(stats UnwindApplyStats, flags UnwindFlags, colorOn bool) string {
	var parts []string
	if stats.HadPeels {
		parts = append(parts, fmt.Sprintf("peeled %s", paintCount(stats.Peeled, colorOn)))
	}
	if flags.TagNext {
		parts = append(parts, fmt.Sprintf("tagged %s", paintCount(stats.Tagged, colorOn)))
		parts = append(parts, fmt.Sprintf("pinned %s", paintCount(stats.Pinned, colorOn)))
	}
	if flags.Push {
		parts = append(parts, fmt.Sprintf("pushed %s", paintCount(stats.Pushed, colorOn)))
	}
	if flags.ReinstallLocal {
		parts = append(parts, fmt.Sprintf("reinstalled %s", paintCount(stats.Reinstalled, colorOn)))
	}
	if len(parts) == 0 {
		return ""
	}
	prefix := "unwind:"
	if colorOn {
		prefix = colorize(prefix, ansiGrey)
	}
	return prefix + " " + strings.Join(parts, ", ")
}