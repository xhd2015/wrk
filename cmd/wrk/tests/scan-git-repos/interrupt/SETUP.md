# Scenario

**Feature**: Ctrl-C / SIGINT mid --scan-git-repos cancels scan, exits 130, warns, keeps partial progress

```
# mid-scan interrupt (SIGINT after first newly printed main)
wrk --scan-git-repos [--no-cache] ROOT_FIRST ROOT_LATER
  -> OnRepo records + prints first main (source=scan)
  -> harness sends SIGINT after first stdout path line
  -> cancel scan context (stop further walk)  # product P2
  -> stderr warning: interrupted + progress saved
  -> exit 130 (ExitCodeError silent body; warning already on stderr)
  -> projects.json keeps mains already recorded; later unvisited mains absent

# contrast: full scan without signal
wrk --scan-git-repos ROOT...  -> exit 0; no interrupt warning
```

## Preconditions

- Git available (parent scan-git-repos Setup skips if missing).
- Isolated `WRK_HOME` at `{WorkRoot}/.wrk`; cwd non-git `{WorkRoot}`.
- Root `Run` cannot deliver mid-flight signals (runs to completion). Interrupt
  leaves use a subtree helper that spawns `wrk` with stdout pipe, signals
  `SIGINT` after the first newly printed path line, then waits for exit.
- Product contract (implemented): SIGINT/SIGTERM cancel scan context → exit 130
  via `ExitCodeError{130}`, stderr `warning:` progress saved, partial projects kept.
  Raw signal death without handler would yield `ExitCode() == -1` (not expected).

## Context

Helper `runScanGitReposSIGINTAfterFirstStdout`:

1. Removes `{WRK_HOME}/projects.json` so partial-record asserts are not polluted
   by a prior full root `Run` of the same Args.
2. Starts `wrk` with `buildWrkCLIArgs(req)` / `wrkEnv(req)` / `req.RepoDir`.
3. Reads stdout until the first complete line (`…\n`), then
   `Process.Signal(syscall.SIGINT)`.
4. Drains remaining stdout/stderr and waits (30s timeout → Kill).
5. Returns stdout, stderr, and exit code (`-1` when terminated by signal without
   a normal exit status).

```go
import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// scanInterruptResult is the outcome of a mid-scan SIGINT probe.
type scanInterruptResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Setup(t *testing.T, req *Request) error {
	ensureScanGitReposHelpersUsed()
	ensureScanInterruptHelpersUsed()
	return nil
}

// runScanGitReposSIGINTAfterFirstStdout runs wrk with req.Args, sends SIGINT
// after the first newly printed stdout path line, and captures exit/output.
// Resets projects.json first so recorded state reflects only this probe.
func runScanGitReposSIGINTAfterFirstStdout(t *testing.T, req *Request) scanInterruptResult {
	t.Helper()

	// Fresh projects so partial recording is attributable to this probe only.
	// (Root Run may have already completed a full scan with the same Args.)
	if err := os.Remove(scanProjectsJSONPath(req.WrkHome)); err != nil && !os.IsNotExist(err) {
		t.Fatalf("remove projects.json: %v", err)
	}

	bin := getWrkBin(t)
	args := buildWrkCLIArgs(req)
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = wrkEnv(req)

	stdoutR, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("start wrk: %v", err)
	}

	var (
		stdoutMu  sync.Mutex
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)

	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		_, _ = io.Copy(&stderrBuf, stderrR)
	}()

	// Read first line → SIGINT → drain remainder.
	readDone := make(chan error, 1)
	go func() {
		r := bufio.NewReader(stdoutR)
		line, readErr := r.ReadString('\n')
		if line != "" {
			stdoutMu.Lock()
			stdoutBuf.WriteString(line)
			stdoutMu.Unlock()
		}
		if line != "" || readErr == nil {
			// First newly printed path (or empty line) observed — interrupt.
			if cmd.Process != nil {
				_ = cmd.Process.Signal(syscall.SIGINT)
			}
		}
		if readErr != nil && readErr != io.EOF {
			// Still try to drain whatever remains after a partial read.
			rest, _ := io.ReadAll(r)
			if len(rest) > 0 {
				stdoutMu.Lock()
				stdoutBuf.Write(rest)
				stdoutMu.Unlock()
			}
			readDone <- readErr
			return
		}
		rest, _ := io.ReadAll(r)
		if len(rest) > 0 {
			stdoutMu.Lock()
			stdoutBuf.Write(rest)
			stdoutMu.Unlock()
		}
		if readErr == io.EOF && line == "" {
			readDone <- io.EOF
			return
		}
		readDone <- nil
	}()

	waitCh := make(chan error, 1)
	go func() { waitCh <- cmd.Wait() }()

	var waitErr error
	select {
	case waitErr = <-waitCh:
	case <-time.After(30 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		waitErr = <-waitCh
		t.Fatalf("timeout waiting for wrk after SIGINT (30s); stdout=%q stderr=%q",
			stdoutBuf.String(), stderrBuf.String())
	}

	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Logf("stdout reader did not finish within 5s after process exit")
	}
	select {
	case <-stderrDone:
	case <-time.After(5 * time.Second):
		t.Logf("stderr reader did not finish within 5s after process exit")
	}

	exitCode := 0
	if waitErr != nil {
		if ee, ok := waitErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode() // -1 when terminated by signal without normal exit
		} else {
			t.Fatalf("wait wrk: %v", waitErr)
		}
	}

	stdoutMu.Lock()
	out := stdoutBuf.String()
	stdoutMu.Unlock()

	return scanInterruptResult{
		Stdout:   out,
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
