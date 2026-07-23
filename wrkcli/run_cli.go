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
// working directory. WrkHome / WrkDate / ExtraEnv / PathPrepend are applied for
// this call only (serialized when any isolation is needed).
type RunOptions struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Stdin   io.Reader // when non-nil, installed as os.Stdin for the call (serialized)
	Dir     string    // process cwd for this invocation (os.Chdir when set + isolation)
	WrkHome string    // overrides WRK_HOME when non-empty
	WrkDate string    // overrides WRK_DATE when non-empty
	// ExtraEnv is KEY=VAL entries applied via Setenv for the call duration.
	ExtraEnv []string
	// PathPrepend dirs are prepended to PATH for the call duration (order preserved).
	PathPrepend []string
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
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	// Serialize all in-process runs that capture I/O or mutate process env/cwd.
	// Prevents parallel doctest leaves from interleaving stdout capture or Setenv.
	needsIsolation := opts.Stdout != nil || opts.Stderr != nil || opts.Stdin != nil ||
		strings.TrimSpace(opts.Dir) != "" ||
		strings.TrimSpace(opts.WrkHome) != "" ||
		strings.TrimSpace(opts.WrkDate) != "" ||
		len(opts.ExtraEnv) > 0 ||
		len(opts.PathPrepend) > 0

	if needsIsolation {
		captureMu.Lock()
		defer captureMu.Unlock()
	}

	// Install writers for product paths that call cliStdout/cliStderr.
	if opts.Stdout != nil || opts.Stderr != nil {
		restoreIO := installCLIWriters(stdout, stderr)
		defer restoreIO()
	}

	// Optional stdin (confirm prompts, etc.).
	if opts.Stdin != nil {
		restoreStdin, err := installStdin(opts.Stdin)
		if err != nil {
			return err
		}
		defer restoreStdin()
	}

	// Resolve Dir as origWd for run() (effective work dir). Match binary cmd.Dir
	// semantics by chdir for the duration so os.Getwd()-based paths match.
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
		if needsIsolation {
			prev, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get cwd: %w", err)
			}
			if err := os.Chdir(origWd); err != nil {
				return fmt.Errorf("chdir %s: %w", origWd, err)
			}
			defer func() { _ = os.Chdir(prev) }()
		}
	}

	// Env isolation (serialized with captureMu when needsIsolation).
	var envPairs []string
	if v := strings.TrimSpace(opts.WrkHome); v != "" {
		envPairs = append(envPairs, "WRK_HOME="+v)
	}
	if v := strings.TrimSpace(opts.WrkDate); v != "" {
		envPairs = append(envPairs, "WRK_DATE="+v)
	}
	envPairs = append(envPairs, opts.ExtraEnv...)
	if len(opts.PathPrepend) > 0 {
		path := os.Getenv("PATH")
		for i := len(opts.PathPrepend) - 1; i >= 0; i-- {
			d := strings.TrimSpace(opts.PathPrepend[i])
			if d == "" {
				continue
			}
			if path == "" {
				path = d
			} else {
				path = d + string(os.PathListSeparator) + path
			}
		}
		envPairs = append(envPairs, "PATH="+path)
	}
	if len(envPairs) > 0 {
		restoreEnv := applyEnvPairs(envPairs)
		defer restoreEnv()
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
// process exit code.
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

func applyEnvPairs(pairs []string) (restore func()) {
	type undo struct {
		key    string
		val    string
		existed bool
	}
	var undos []undo
	for _, kv := range pairs {
		key, val, ok := strings.Cut(kv, "=")
		if !ok || key == "" {
			continue
		}
		prev, existed := os.LookupEnv(key)
		undos = append(undos, undo{key: key, val: prev, existed: existed})
		_ = os.Setenv(key, val)
	}
	return func() {
		for i := len(undos) - 1; i >= 0; i-- {
			u := undos[i]
			if u.existed {
				_ = os.Setenv(u.key, u.val)
			} else {
				_ = os.Unsetenv(u.key)
			}
		}
	}
}

func installStdin(r io.Reader) (restore func(), err error) {
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	old := os.Stdin
	os.Stdin = pr
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer pw.Close()
		_, _ = io.Copy(pw, r)
	}()
	return func() {
		_ = pr.Close()
		<-done
		os.Stdin = old
	}, nil
}
