package wrkcli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// RunOptions configures an in-process wrk invocation.
//
// Writers default to os.Stdout / os.Stderr when nil. Dir defaults to the process
// working directory. WrkHome / WrkDate override process env for this call only
// (parallel-safe; no os.Setenv).
type RunOptions struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Dir     string // process effective cwd for this invocation (origWd)
	WrkHome string // overrides WRK_HOME when non-empty
	WrkDate string // overrides WRK_DATE when non-empty
}

// Run executes wrk with process stdout/stderr and the current working directory.
func Run(args []string) error {
	return RunWithOptions(args, RunOptions{})
}

// RunWithWriters executes wrk writing success/help output to the given writers.
// Product failures are returned as errors (same as Run); they are not printed.
// Use RunCLI to mirror cmd/wrk main (format Error:/warning: onto stderr + exit code).
func RunWithWriters(args []string, stdout, stderr io.Writer) error {
	return RunWithOptions(args, RunOptions{Stdout: stdout, Stderr: stderr})
}

// RunWithOptions executes wrk with the given options.
func RunWithOptions(args []string, opts RunOptions) error {
	origWd := strings.TrimSpace(opts.Dir)
	if origWd == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get cwd: %w", err)
		}
		origWd = wd
	} else {
		abs, err := filepath.Abs(origWd)
		if err != nil {
			return fmt.Errorf("resolve dir: %w", err)
		}
		origWd = abs
	}

	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	ctx := newInvocationContext(origWd, args)
	ctx.stdout = stdout
	ctx.stderr = stderr
	ctx.wrkHomeOverride = strings.TrimSpace(opts.WrkHome)
	ctx.wrkDateOverride = strings.TrimSpace(opts.WrkDate)

	var runErr error
	defer func() {
		exitCode := 0
		if runErr != nil {
			var ece ExitCodeError
			if errors.As(runErr, &ece) {
				exitCode = ece.Code
			} else {
				exitCode = 1
			}
		}
		ctx.finish(exitCode)
	}()
	runErr = run(origWd, args, ctx)
	return runErr
}

// RunCLI mirrors cmd/wrk main for in-process tests: runs wrk with opts, writes
// formatted Error:/warning: lines to opts.Stderr on failure, and returns the
// process exit code. Setup failures (e.g. bad Dir) still surface as non-zero
// exit with the error message on stderr.
func RunCLI(args []string, opts RunOptions) int {
	err := RunWithOptions(args, opts)
	if err == nil {
		return 0
	}
	code := 1
	var ece ExitCodeError
	if errors.As(err, &ece) {
		code = ece.Code
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	msg := err.Error()
	msg = FormatStderrError(msg)
	msg = FormatStderrWarning(msg)
	fmt.Fprintln(stderr, msg)
	return code
}
