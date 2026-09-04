package unwind

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

func paint(s, code string, on bool) string {
	if !on || s == "" {
		return s
	}
	return colorize(s, code)
}

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
