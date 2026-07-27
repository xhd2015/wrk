package wrkcli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// runCaptureMu serializes Capture / RunWithWriters because product code writes
// to os.Stdout/os.Stderr and may optionally Chdir for Dir.
//
// WRK_HOME / WRK_DATE are injected via package overrides (not os.Setenv) so
// Capture can run in the same suite process as parallel binary leaves that
// build child env from os.Environ().
var runCaptureMu sync.Mutex

// captureWrkHome / captureWrkDate / captureDir / captureUserHome are set only
// while runCaptureMu is held. resolveWrkHome / resolveWrkDate / userHomeDir
// prefer these overrides over process env. captureDir is a virtual process cwd
// (no os.Chdir); captureUserHome peels HOME= from Capture Env without Setenv so
// parallel doctest Setup/git keep a valid real getcwd and real $HOME.
var (
	captureWrkHome   string
	captureWrkDate   string
	captureDir       string
	captureUserHome  string
)

// processCwd returns the effective process working directory. During Capture
// with Dir set, this is the virtual capture Dir (no os.Chdir). Otherwise it is
// os.Getwd(). Prefer this over os.Getwd() in product code so InProcess tests
// do not mutate process cwd under t.Parallel().
func processCwd() (string, error) {
	if captureDir != "" {
		return captureDir, nil
	}
	return os.Getwd()
}

// absAgainstProcessCwd is filepath.Abs but uses processCwd for relative paths
// so Capture Dir is honored without os.Chdir.
func absAgainstProcessCwd(path string) (string, error) {
	if path == "" {
		return processCwd()
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}
	wd, err := processCwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(wd, path)), nil
}

// userHomeDir returns Capture HOME override when set; otherwise os.UserHomeDir.
// Prefer this over os.UserHomeDir in product code so FakeHome InProcess leaves
// do not Setenv("HOME") under t.Parallel().
func userHomeDir() (string, error) {
	if captureUserHome != "" {
		return captureUserHome, nil
	}
	return os.UserHomeDir()
}

// CaptureOpts configures an in-process CLI invocation that mirrors main's
// stdout/stderr/exit-code contract without spawning the wrk binary.
type CaptureOpts struct {
	// Args are CLI args after the program name (same as wrkcli.Run).
	Args []string
	// Dir, when non-empty, is the virtual process cwd for the duration of the
	// call (via processCwd / captureDir under the capture mutex — no os.Chdir,
	// so parallel test Setup/git keep a valid real getcwd). Prefer absolute
	// paths. Omit when cwd is irrelevant (most short paths).
	Dir string
	// WrkHome, when non-empty, overrides WRK_HOME without os.Setenv (Parallel-safe).
	WrkHome string
	// Env are additional KEY=VAL entries. WRK_HOME and WRK_DATE are peeled into
	// capture overrides (no Setenv). Other keys still use temporary Setenv under
	// the capture mutex — avoid those when mixing with parallel Environ readers.
	Env []string
	// Stdin, when non-empty, replaces os.Stdin with a pipe containing this
	// data for the duration of the call (restored afterward). Use for
	// --confirm-from-stdin and similar non-TTY prompt paths. Empty leaves
	// the process stdin unchanged (typically not a pipe under go test).
	Stdin string
}

// CaptureResult is the CLI-shaped outcome of Capture / RunWithWriters.
type CaptureResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Capture runs wrk in-process, capturing stdout and stderr like a subprocess.
// Errors are formatted onto stderr the same way as cmd/wrk main (except
// ExitCodeError, which stays silent and only sets ExitCode).
//
// Capture holds an internal mutex for the duration of the call (stdout/stderr
// swap, optional virtual cwd). Prefer for short paths (help, skill, version,
// flag reject) and dual-mode InProcess leaves.
func Capture(opts CaptureOpts) CaptureResult {
	runCaptureMu.Lock()
	defer runCaptureMu.Unlock()

	restoreCapture := applyCaptureOverrides(opts)
	defer restoreCapture()

	restoreDir := applyCaptureDir(opts.Dir)
	defer restoreDir()

	restoreStdin, errStdin := applyStdin(opts.Stdin)
	defer restoreStdin()
	if errStdin != nil {
		return CaptureResult{Stderr: errStdin.Error(), ExitCode: 1}
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	finishIO, errSetup := beginStdoutStderrCapture(&stdoutBuf, &stderrBuf)
	if errSetup != nil {
		return CaptureResult{Stderr: errSetup.Error(), ExitCode: 1}
	}

	err := Run(opts.Args)
	if err != nil {
		var ece ExitCodeError
		if !errors.As(err, &ece) {
			// Match cmd/wrk main: format Error:/warning: prefixes, print to stderr, exit 1.
			msg := err.Error()
			msg = FormatStderrError(msg)
			msg = FormatStderrWarning(msg)
			if !strings.HasSuffix(msg, "\n") {
				msg += "\n"
			}
			_, _ = fmt.Fprint(os.Stderr, msg)
		}
	}

	finishIO() // close pipes and wait for copy goroutines before reading buffers

	res := CaptureResult{
		Stdout: stdoutBuf.String(),
		Stderr: stderrBuf.String(),
	}
	if err == nil {
		return res
	}
	var ece ExitCodeError
	if errors.As(err, &ece) {
		res.ExitCode = ece.Code
		return res
	}
	res.ExitCode = 1
	return res
}

// RunWithWriters runs wrk in-process, writing primary output to the given
// writers (nil → discard). Env/Dir isolation matches Capture. Prefer Capture
// when you need the full CLI exit-code + formatted-stderr contract.
//
// Note: non-ExitCodeError return values are not auto-printed to stderr (unlike
// Capture / main); callers that need main's formatting should use Capture.
func RunWithWriters(stdout, stderr io.Writer, opts CaptureOpts) error {
	runCaptureMu.Lock()
	defer runCaptureMu.Unlock()

	restoreCapture := applyCaptureOverrides(opts)
	defer restoreCapture()

	restoreDir := applyCaptureDir(opts.Dir)
	defer restoreDir()

	restoreStdin, errStdin := applyStdin(opts.Stdin)
	defer restoreStdin()
	if errStdin != nil {
		return errStdin
	}

	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}
	finishIO, errSetup := beginStdoutStderrCapture(stdout, stderr)
	if errSetup != nil {
		return errSetup
	}
	err := Run(opts.Args)
	finishIO()
	return err
}

// RunCapture is a convenience wrapper: Capture with only Args set.
func RunCapture(args ...string) CaptureResult {
	return Capture(CaptureOpts{Args: args})
}

// applyCaptureOverrides installs Parallel-safe WRK_HOME/WRK_DATE/HOME overrides
// and optionally Setenv for remaining Env keys (under runCaptureMu).
func applyCaptureOverrides(opts CaptureOpts) (restore func()) {
	wrkHome := opts.WrkHome
	wrkDate := ""
	userHome := ""
	var rest []string
	for _, p := range opts.Env {
		key, val, ok := strings.Cut(p, "=")
		if !ok || key == "" {
			continue
		}
		switch key {
		case "WRK_HOME":
			if wrkHome == "" {
				wrkHome = val
			}
		case "WRK_DATE":
			wrkDate = val
		case "HOME":
			userHome = val
		default:
			rest = append(rest, p)
		}
	}

	prevHome, prevDate, prevUser := captureWrkHome, captureWrkDate, captureUserHome
	captureWrkHome = wrkHome
	captureWrkDate = wrkDate
	captureUserHome = userHome

	// HOME is also Setenv under the capture mutex so product code that still
	// uses os.Getenv("HOME") / expand~ (not only userHomeDir) sees FakeHome.
	// Other keys stay in rest as before.
	if userHome != "" {
		rest = append([]string{"HOME=" + userHome}, rest...)
	}

	restoreEnv := applyEnvPairs(rest)

	return func() {
		restoreEnv()
		captureWrkHome = prevHome
		captureWrkDate = prevDate
		captureUserHome = prevUser
	}
}

func applyEnvPairs(pairs []string) (restore func()) {
	if len(pairs) == 0 {
		return func() {}
	}
	type saved struct {
		key    string
		val    string
		hadVal bool
	}
	var prev []saved
	for _, p := range pairs {
		key, val, ok := strings.Cut(p, "=")
		if !ok || key == "" {
			continue
		}
		old, had := os.LookupEnv(key)
		prev = append(prev, saved{key: key, val: old, hadVal: had})
		_ = os.Setenv(key, val)
	}
	return func() {
		for i := len(prev) - 1; i >= 0; i-- {
			s := prev[i]
			if s.hadVal {
				_ = os.Setenv(s.key, s.val)
			} else {
				_ = os.Unsetenv(s.key)
			}
		}
	}
}

// applyCaptureDir installs a virtual cwd for processCwd() without os.Chdir.
// Real process cwd stays untouched so parallel leaves can still os.Getwd/git
// and TempDir cleanup cannot leave Capture holding a deleted getcwd.
// Callers that need third-party os.Getwd/filepath.Abs behavior must pass
// absolute paths or pin --dir (see runGenCommitMsg).
func applyCaptureDir(dir string) (restore func()) {
	prev := captureDir
	captureDir = dir
	return func() { captureDir = prev }
}

// applyStdin replaces os.Stdin with a pipe of data when data is non-empty.
// The pipe is a non-TTY character device (ModeNamedPipe), so confirm-from-stdin
// and similar paths treat it as piped input. Empty data is a no-op.
func applyStdin(data string) (restore func(), err error) {
	if data == "" {
		return func() {}, nil
	}
	r, w, err := os.Pipe()
	if err != nil {
		return func() {}, err
	}
	old := os.Stdin
	os.Stdin = r
	var once sync.Once
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.WriteString(w, data)
		_ = w.Close()
	}()
	return func() {
		once.Do(func() {
			<-done
			_ = r.Close()
			os.Stdin = old
		})
	}, nil
}

// beginStdoutStderrCapture redirects os.Stdout/os.Stderr to pipes that copy into
// the given writers. finish must be called to restore and wait for copy.
func beginStdoutStderrCapture(stdout, stderr io.Writer) (finish func(), err error) {
	oldOut, oldErr := os.Stdout, os.Stderr

	outR, outW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		_ = outR.Close()
		_ = outW.Close()
		return nil, err
	}

	os.Stdout = outW
	os.Stderr = errW

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(stdout, outR)
		_ = outR.Close()
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(stderr, errR)
		_ = errR.Close()
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			_ = outW.Close()
			_ = errW.Close()
			wg.Wait()
			os.Stdout = oldOut
			os.Stderr = oldErr
		})
	}, nil
}
