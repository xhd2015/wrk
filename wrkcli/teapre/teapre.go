// Package teapre must be imported before bubbletea so OSC background probes
// do not hang on PTYs that never answer (tty-watch, some CI harnesses).
//
// bubbletea's init() calls lipgloss.HasDarkBackground(); if we set the value
// explicitly first, that call skips termenv's terminal query.
package teapre

import "github.com/charmbracelet/lipgloss"

func init() {
	lipgloss.SetHasDarkBackground(true)
}
