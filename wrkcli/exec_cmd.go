package wrkcli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// runExecInDir runs argv in dir (absolute) after a successful mode action.
// Empty argv is a no-op. Non-zero child exit becomes ExitCodeError.
// Child stdout/stderr default to process streams; use runExecInDirTo for L2 capture.
func runExecInDir(dir string, argv []string) error {
	return runExecInDirTo(os.Stdout, os.Stderr, dir, argv)
}

// runExecInDirTo is like runExecInDir but attaches child streams to out/errw
// (nil → process defaults).
func runExecInDirTo(out, errw io.Writer, dir string, argv []string) error {
	if len(argv) == 0 {
		return nil
	}
	if out == nil {
		out = os.Stdout
	}
	if errw == nil {
		errw = os.Stderr
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve exec dir: %w", err)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = absDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = errw
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 0 {
				code = 1
			}
			return ExitCodeError{Code: code}
		}
		return fmt.Errorf("wrk: exec %s: %w", argv[0], err)
	}
	return nil
}
