package wrkcli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/interactive"
)

// runCd jumps into absDir: in-place follow-up when WRK_FOLLOWUP_FILE is set,
// otherwise prints install hint + abs path and launches an interactive shell.
// When execArgs is non-empty, runs the command in absDir after a successful jump
// (follow-up is still written when the channel is open).
func runCd(absDir string, execArgs []string) error {
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
	if strings.TrimSpace(os.Getenv("WRK_FOLLOWUP_FILE")) != "" {
		if err := writeFollowupCD(false, absDir); err != nil {
			return err
		}
		return runExecInDir(absDir, execArgs)
	}

	// Branch B — fallback: warn, print abs path, launch interactive shell.
	fmt.Fprintf(os.Stderr, "warning: bash integration not active; install with: wrk --bash-integration --install\n")
	fmt.Fprintln(os.Stdout, absDir)

	err = interactive.LoginInteractive(absDir, filepath.Base(absDir), "WRK_SHELL=1")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return ExitCodeError{Code: exitErr.ExitCode()}
		}
		return err
	}
	return runExecInDir(absDir, execArgs)
}

// forceLandInDir lands the user in dest after successful create/--done/--set-task
// with --force-cd. Gates are already bypassed by the caller. Dual path like runCd:
// follow-up file when WRK_FOLLOWUP_FILE is set; otherwise install-hint + interactive
// shell. Does not re-print dest on stdout (mode already printed path/message).
func forceLandInDir(dest string) error {
	// Branch A — bash integration channel open: in-place follow-up only.
	if strings.TrimSpace(os.Getenv("WRK_FOLLOWUP_FILE")) != "" {
		return writeFollowupCD(false, dest)
	}

	// Branch B — fallback: warn and launch interactive shell (no stdout path).
	fmt.Fprintf(os.Stderr, "warning: bash integration not active; install with: wrk --bash-integration --install\n")
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
