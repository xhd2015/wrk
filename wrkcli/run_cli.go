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
// Parallel-safe contract (DOCTEST_LINT.md §1):
//   - Stdout/Stderr are attached to the invocation context only (ctx.out/errw).
//   - WrkHome / WrkDate / Home / Env are per-invocation overrides (no os.Setenv).
//   - Dir is the logical work dir passed into run() as origWd (no os.Chdir).
//   - PathPrepend and Stdin are not supported in-process: use a product-binary
//     subprocess with cmd.Env / cmd.Dir instead (L3). ExtraEnv is rejected;
//     pass known keys via Env (or named fields) for dual-mode L2.
type RunOptions struct {
	Stdout  io.Writer // nil → os.Stdout for ctx.out()
	Stderr  io.Writer // nil → os.Stderr for ctx.errw()
	Dir     string    // logical work dir (origWd); not os.Chdir
	WrkHome string    // overrides WRK_HOME for this call only
	WrkDate string    // overrides WRK_DATE for this call only
	// Home overrides os.UserHomeDir for this call only (FakeHome for L2).
	Home string
	// Gobin overrides GOBIN for --reinstall-local only (no os.Setenv). Empty
	// falls back to process GOBIN / $(go env GOPATH)/bin.
	Gobin string
	// FollowupFile overrides WRK_FOLLOWUP_FILE for --cd / force-cd land paths
	// without os.Setenv (parallel-safe L2).
	FollowupFile string
	// Env is a per-invocation KEY→VAL overlay read via ctx.getenv (no Setenv).
	// Typical keys: WRK_DASHBOARD_*, WRK_SCAN_DEBUG, WRK_SET_TASK_CONFIRM,
	// WRK_BASENAME_CONFIRM, WRK_TASK_LIKE_CONFIRM, WRK_PROJECTS_PERF_LOG.
	Env map[string]string

	// Unsupported in-process (hard error). Prefer binary + cmd.Env/cmd.Dir.
	Stdin       io.Reader
	ExtraEnv    []string
	PathPrepend []string
}

// ErrProcessIsolationRequired is returned when RunOptions request process-global
// isolation (env/stdin) that is not parallel-safe in-process.
var ErrProcessIsolationRequired = errors.New("wrkcli: ExtraEnv/PathPrepend/Stdin require product-binary isolation (cmd.Env/cmd.Dir); not supported in-process")

// Run executes wrk with process stdout/stderr and the current working directory.
func Run(args []string) error {
	return RunWithOptions(args, RunOptions{})
}

// RunWithWriters executes wrk with capturable stdout/stderr on the invocation
// context. Parallel-safe when product paths print via ctx.out()/ctx.errw().
func RunWithWriters(args []string, stdout, stderr io.Writer) error {
	return RunWithOptions(args, RunOptions{Stdout: stdout, Stderr: stderr})
}

// RunWithOptions executes wrk with the given options.
func RunWithOptions(args []string, opts RunOptions) error {
	if opts.Stdin != nil || len(opts.ExtraEnv) > 0 || len(opts.PathPrepend) > 0 {
		return fmt.Errorf("%w", ErrProcessIsolationRequired)
	}

	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

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

	ctx := newInvocationContext(origWd, args)
	ctx.stdout = stdout
	ctx.stderr = stderr
	ctx.wrkHomeOverride = strings.TrimSpace(opts.WrkHome)
	ctx.wrkDateOverride = strings.TrimSpace(opts.WrkDate)
	ctx.gobinOverride = strings.TrimSpace(opts.Gobin)
	ctx.followupFileOverride = strings.TrimSpace(opts.FollowupFile)
	ctx.homeOverride = strings.TrimSpace(opts.Home)
	if len(opts.Env) > 0 {
		ctx.envOverride = make(map[string]string, len(opts.Env))
		for k, v := range opts.Env {
			ctx.envOverride[k] = v
		}
	}

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
// process exit code. Isolation fields (ExtraEnv/PathPrepend/Stdin) fail closed.
func RunCLI(args []string, opts RunOptions) int {
	err := RunWithOptions(args, opts)
	if err == nil {
		return 0
	}
	if errors.Is(err, ErrProcessIsolationRequired) {
		// Setup mistake: surface clearly on stderr.
		stderr := opts.Stderr
		if stderr == nil {
			stderr = os.Stderr
		}
		fmt.Fprintln(stderr, err.Error())
		return 2
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
