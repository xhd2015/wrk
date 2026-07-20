package tui

import (
	"strings"
	"testing"
)

func TestSectionHRLinesPresent(t *testing.T) {
	for _, color := range []bool{false, true} {
		m := newTeaDashModel(RunDashboardOpts{WorkDir: "/tmp"})
		m.width = 100
		m.color = color
		out := m.renderView()
		for _, name := range []string{"Pre", "Main", "After", "Batch", "Log"} {
			found := false
			for _, ln := range strings.Split(out, "\n") {
				// frameSection / colored HR: contains section name and horizontal rule dashes.
				if strings.Contains(ln, name) && strings.Contains(ln, "─") &&
					(strings.Contains(ln, "├") || strings.Contains(ln, "\x1b")) {
					found = true
					break
				}
				// colorless exact form
				if strings.Contains(ln, "─ "+name+" ") && strings.Contains(ln, "├") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("color=%v missing section HR for %q in:\n%s", color, name, out)
			}
		}
	}
}
