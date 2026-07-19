package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse"
)

// RunDashboard runs the interactive Bubble Tea dashboard until CANCEL.
// Stage / RUN ALL ops run in-process with a loading spinner (no tear-down flash).
func RunDashboard(opts RunDashboardOpts) error {
	m := newTeaDashModel(opts)
	mouseDebugBanner(map[string]any{
		"workDir": opts.WorkDir, "status": opts.Status,
		"addDisabled": m.addDisabled, "addAll": m.addAll,
		"originMode": "tui/mouse-tracker",
	})

	// Shared inline-mouse package: CSI 6n in View + CPR on same stdin as mouse.
	cprCh := make(chan mouse.CPRMsg, 8)
	m.cprCh = cprCh
	in := mouse.NewFilter(os.Stdin, cprCh)
	in.OnDrop = func(msg mouse.CPRMsg) {
		mouseDebugf("cpr_drop", map[string]any{"row1": msg.Row1, "col1": msg.Col1})
	}

	p := tea.NewProgram(
		&m,
		tea.WithInput(in),
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
