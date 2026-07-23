package wrkcli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/interactive"
)

// followupChannelOpen reports whether the bash-integration follow-up channel is open
// via invocation override or process env.
func followupChannelOpen(followupFile string) bool {
	if strings.TrimSpace(followupFile) != "" {
		return true
	}
	return strings.TrimSpace(os.Getenv("WRK_FOLLOWUP_FILE")) != ""
}

// runCd jumps into absDir: in-place follow-up when the follow-up channel is open
// (RunOptions.FollowupFile or WRK_FOLLOWUP_FILE), otherwise prints install hint +
// abs path and launches an interactive shell.
// When execArgs is non-empty, runs the command in absDir after a successful jump
// (follow-up is still written when the channel is open).
func runCd(out, errw io.Writer, absDir string, execArgs []string, followupFile string) error {
	if out == nil {
		out = os.Stdout
	}
	if errw == nil {
		errw = os.Stderr
	}
	info, err := os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wrk: %s does not exist", absDir)
		}
		return fmt.Errorf("stat dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("wrk: %s is not a directory", absDir)
	}

	// Branch A — bash integration channel open: in-place follow-up only.
	if followupChannelOpen(followupFile) {
		if err := writeFollowupCDTo(false, absDir, followupFile); err != nil {
			return err
		}
		return runExecInDirTo(out, errw, absDir, execArgs)
	}

	// Branch B — fallback: warn, print abs path, launch interactive shell.
	fmt.Fprintf(errw, "warning: bash integration not active; install with: wrk --bash-integration --install\n")
	fmt.Fprintln(out, absDir)

	err = interactive.LoginInteractive(absDir, filepath.Base(absDir), "WRK_SHELL=1")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExitCodeError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return runExecInDirTo(out, errw, absDir, execArgs)
}

// forceLandInDir lands the user in dest after successful create/--done/--set-task
// with --force-cd. Gates are already bypassed by the caller. Dual path like runCd:
// follow-up file when channel open; otherwise install-hint + interactive shell.
// Does not re-print dest on stdout (mode already printed path/message).
// followupFile is an optional invocation override (same as runCd).
func forceLandInDir(errw io.Writer, dest, followupFile string) error {
	if errw == nil {
		errw = os.Stderr
	}
	// Branch A — bash integration channel open: in-place follow-up only.
	if followupChannelOpen(followupFile) {
		return writeFollowupCDTo(false, dest, followupFile)
	}

	// Branch B — fallback: warn and launch interactive shell (no stdout path).
	fmt.Fprintf(errw, "warning: bash integration not active; install with: wrk --bash-integration --install\n")
	err := interactive.LoginInteractive(dest, filepath.Base(dest), "WRK_SHELL=1")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExitCodeError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return nil
}
