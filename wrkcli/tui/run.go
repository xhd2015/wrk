package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// RunDashboard runs the interactive Bubble Tea dashboard until CANCEL.
// Stage / RUN ALL ops run in-process with a loading spinner (no tear-down flash).
func RunDashboard(opts RunDashboardOpts) error {
	m := newTeaDashModel(opts)
	// No alt-screen (inline UI). Mouse Y is terminal-absolute; mapMouseY
	// converts using height - viewLines (inline paint sits at bottom).
	p := tea.NewProgram(
		&m,
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
		tea.WithMouseCellMotion(),
	)
	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("wrk: dashboard tui: %w", err)
	}
	if _, ok := final.(*teaDashModel); !ok {
		return fmt.Errorf("wrk: unexpected tea model type %T", final)
	}
	return nil
}
