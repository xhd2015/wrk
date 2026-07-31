# Scenario

**Feature**: Ctrl-C / SIGINT mid --scan-git-repos cancels scan, exits 130, warns, keeps partial progress

```
# mid-scan interrupt (SIGINT after first newly printed main)
wrk --scan-git-repos [--no-cache] ROOT_FIRST ROOT_LATER
  -> OnRepo prints first main (no record)
  -> harness sends SIGINT after first stdout path line
  -> cancel scan context (stop further walk)  # product P2
  -> stderr warning: interrupted; cache may keep progress
  -> exit 130 (ExitCodeError silent body; warning already on stderr)
  -> projects.json unchanged (still empty / pre-seeded only)

# contrast: full scan without signal
wrk --scan-git-repos ROOT...  -> exit 0; no interrupt warning
```

## Preconditions

- Git available (parent scan-git-repos Setup skips if missing).
- Isolated `WRK_HOME` at `{WorkRoot}/.wrk`; cwd non-git `{WorkRoot}`.
- Root `Run` cannot deliver mid-flight signals (runs to completion). Interrupt
  leaves use a subtree helper that runs in-process via `wrkcli.RunWithWriters`
  with `ScanTestPauseAfterFirstPrint` (options down the stack, not env), signals
  `SIGINT` after the first newly printed path line, then waits for exit.
- Product contract (implemented): SIGINT/SIGTERM cancel scan context → exit 130
  via `ExitCodeError{130}`, stderr `warning:` progress saved, projects.json untouched.
  Raw signal death without handler would yield `ExitCode() == -1` (not expected).

## Context

Helper `runScanGitReposSIGINTAfterFirstStdout`:

1. Removes `{WRK_HOME}/projects.json` so partial-record asserts are not polluted
   by a prior full root `Run` of the same Args.
2. Runs in-process via `wrkcli.RunWithWriters` with
   `ScanTestPauseAfterFirstPrint: 200ms` (passed down RunOpts — no env).
3. On the first stdout path line, sends `SIGINT` to this process (product
   `signal.Notify` cancels the scan ctx; default terminate is replaced).
4. Pause after first print keeps the walk open until cancel is observed.
5. Returns stdout, stderr, and exit code from `ExitCodeError` / err.

```go
import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/wrk/wrkcli"
)

// scanInterruptResult is the outcome of a mid-scan SIGINT probe.
type scanInterruptResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// scanInterruptPause is the RunOpts.ScanTestPauseAfterFirstPrint duration used
// by interrupt probes so SIGINT can land while the walk is still open.
const scanInterruptPause = 200 * time.Millisecond

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensureScanGitReposHelpersUsed()
	ensureScanInterruptHelpersUsed()
	return nil
}

// runScanGitReposSIGINTAfterFirstStdout runs --scan-git-repos in-process with
// ScanTestPauseAfterFirstPrint, sends SIGINT after the first printed path line,
// and captures exit/output. Resets projects.json first.
//
// Uses io.Pipe + RunWithWriters (not a custom Write method) so doctest codegen
// stays valid. SIGINT targets this process; product signal.Notify cancels ctx.
func runScanGitReposSIGINTAfterFirstStdout(t *testing.T, req *Request) scanInterruptResult {
	t.Helper()

	// Fresh projects so partial recording is attributable to this probe only.
	// (Root Run may have already completed a full scan with the same Args.)
	if err := os.Remove(scanProjectsJSONPath(req.WrkHome)); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove projects.json: %v", err)
	}

	stdoutR, stdoutW := io.Pipe()
	var stdoutBuf, stderrBuf bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		r := bufio.NewReader(stdoutR)
		line, readErr := r.ReadString('\n')
		if line != "" {
			stdoutBuf.WriteString(line)
			if proc, e := os.FindProcess(os.Getpid()); e == nil {
				_ = proc.Signal(syscall.SIGINT)
			}
		}
		rest, _ := io.ReadAll(r)
		if len(rest) > 0 {
			stdoutBuf.Write(rest)
		}
		_ = readErr
	}()

	env := []string{"WRK_HOME=" + req.WrkHome, "WRK_DATE=" + wrkDate}
	if req.FakeHome != "" {
		env = append(env, "HOME="+req.FakeHome)
	}
	env = appendExtraEnv(env, req)

	err := wrkcli.RunWithWriters(stdoutW, &stderrBuf, wrkcli.CaptureOpts{
		Args:                         buildWrkCLIArgs(req),
		Dir:                          req.RepoDir,
		WrkHome:                      req.WrkHome,
		Env:                          env,
		ScanTestPauseAfterFirstPrint: scanInterruptPause,
	})
	_ = stdoutW.Close()
	<-readDone

	exitCode := 0
	if err != nil {
		var ece wrkcli.ExitCodeError
		if errors.As(err, &ece) {
			exitCode = ece.Code
		} else {
			exitCode = 1
		}
	}

	return scanInterruptResult{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
	}
}

// seedScanPaddingDirs creates many empty sibling dirs under root so walking the
// second scan root stays mid-flight after the first main is printed (SIGINT window).
func seedScanPaddingDirs(t *testing.T, root string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		dir := filepath.Join(root, fmt.Sprintf("a%05d", i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir pad %s: %v", dir, err)
		}
	}
}

// assertInterruptWarning checks stderr for the locked P2 warning shape:
// must contain "warning:" plus interrupt/interrupted and progress/saved (flexible wording).
func assertInterruptWarning(t *testing.T, stderr string) {
	t.Helper()
	if !strings.Contains(stderr, "warning:") {
		t.Fatalf("stderr must contain warning: prefix; got %q", stderr)
	}
	low := strings.ToLower(stderr)
	if !strings.Contains(low, "interrupt") {
		t.Fatalf("stderr warning must mention interrupt/interrupted; got %q", stderr)
	}
	if !strings.Contains(low, "progress") && !strings.Contains(low, "saved") {
		t.Fatalf("stderr warning must mention progress saved (or equivalent); got %q", stderr)
	}
}

func ensureScanInterruptHelpersUsed() {
	_ = runScanGitReposSIGINTAfterFirstStdout
	_ = seedScanPaddingDirs
	_ = assertInterruptWarning
}
```
