# Scenario

**Feature**: --scan-git-repos streams every valid main in discovery order, with first path before finish

```
# multi-root scan: print as each main is found (discovery order; print-only)
wrk --scan-git-repos ROOT_B ROOT_A
  -> scan_repo discovers mains under roots in root order
  -> stdout always prints valid main abs paths in discovery order (not post-scan sort)
  -> projects.json never written

# timing / incremental stdout (first path arrives before process exit)
wrk --scan-git-repos --no-cache ROOT_FIRST ROOT_LATER
  -> OnRepo prints first main quickly
  -> walk continues over pad dirs + late main
  -> first stdout path line arrives with measurable gap before process exit
```

## Preconditions

- Explicit multi-root fixtures under `{WorkRoot}`; cwd remains non-git `{WorkRoot}`.
- Discovery order must differ from lexicographic path sort so order leaves fail under a batch-sorted product.
- Timing leaves need a slow second root (pad dirs + late main) so first-path lead is measurable on a pipe.

## Steps

- Descendants place one main under each scan root and set Args with roots in intentional order.
- Timing leaves may seed pad dirs under the later root and re-run via stream probe in Assert.

## Context

Helper `runScanGitReposStreamProbe`:

1. Removes `{WRK_HOME}/projects.json` so the probe run starts with a clean project list
   (root `Run` may have already completed a full scan with the same Args; always-print
   still lists paths either way, but record side effects stay attributable to the probe).
2. Starts `wrk` with `buildWrkCLIArgs(req)` / `wrkEnv(req)` / `req.RepoDir` and a stdout pipe.
3. Records `firstByteMS` at the first non-empty read and `totalMS` at process exit.
4. Captures first chunk prefix + full stdout and exit code.

```go
import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"github.com/xhd2015/doctest/session"
)

// scanStreamProbe is the outcome of an incremental-stdout timing probe for --scan-git-repos.
type scanStreamProbe struct {
	FirstByteMS int64
	TotalMS     int64
	FirstChunk  string
	FullStdout  string
	ExitCode    int
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: streaming / discovery-order / first-path-before-finish share scan helpers.
	ensureScanGitReposHelpersUsed()
	ensureScanStreamProbeHelpersUsed()
	return nil
}

// seedScanPaddingDirs creates many empty sibling dirs under root so walking the
// second scan root continues after the first main is printed (timing / SIGINT window).
func seedScanPaddingDirs(t *testing.T, root string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		dir := filepath.Join(root, fmt.Sprintf("a%05d", i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir pad %s: %v", dir, err)
		}
	}
}

// runScanGitReposStreamProbe runs wrk with req.Args on a stdout pipe, measuring
// first-byte time vs total duration. Resets projects.json first so stdout is newly recorded paths.
func runScanGitReposStreamProbe(t *testing.T, req *Request) scanStreamProbe {
	t.Helper()

	// Fresh projects so record side effects are attributable to this probe only.
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
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	start := time.Now()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start wrk: %v", err)
	}

	type readResult struct {
		firstByteMS int64
		firstChunk  string
		fullStdout  string
		readErr     error
	}
	readDone := make(chan readResult, 1)
	go func() {
		var firstByteMS int64 = -1
		var firstChunk, rest bytes.Buffer
		buf := make([]byte, 4096)
		for {
			n, readErr := stdoutR.Read(buf)
			if n > 0 {
				if firstByteMS < 0 {
					firstByteMS = time.Since(start).Milliseconds()
					firstChunk.Write(buf[:n])
				} else {
					rest.Write(buf[:n])
				}
			}
			if readErr == io.EOF {
				readDone <- readResult{
					firstByteMS: firstByteMS,
					firstChunk:  firstChunk.String(),
					fullStdout:  firstChunk.String() + rest.String(),
				}
				return
			}
			if readErr != nil {
				readDone <- readResult{readErr: readErr}
				return
			}
		}
	}()

	waitErr := cmd.Wait()
	totalMS := time.Since(start).Milliseconds()

	var rr readResult
	select {
	case rr = <-readDone:
	case <-time.After(60 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		t.Fatalf("timeout reading wrk --scan-git-repos stdout (60s); stderr=%q", stderr.String())
	}
	if rr.readErr != nil {
		t.Fatalf("read stdout: %v", rr.readErr)
	}

	exitCode := 0
	if waitErr != nil {
		if ee, ok := waitErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			t.Fatalf("wait wrk: %v", waitErr)
		}
	}

	return scanStreamProbe{
		FirstByteMS: rr.firstByteMS,
		TotalMS:     totalMS,
		FirstChunk:  rr.firstChunk,
		FullStdout:  rr.fullStdout,
		ExitCode:    exitCode,
	}
}

// assertScanStreamsIncrementally requires first path bytes before process exit and
// that the first chunk begins with the expected first main path line.
func assertScanStreamsIncrementally(t *testing.T, probe scanStreamProbe, firstMainPath string) {
	t.Helper()
	// Fast CI runners (ubuntu-latest) finish the second root quickly; keep
	// thresholds low enough for wall-clock noise while still rejecting batch-at-end.
	const minTotalMS = int64(20)
	const minLeadMS = int64(10)

	if probe.FirstByteMS < 0 {
		t.Fatalf("no stdout until process exit (buffered); total_ms=%d full=%q", probe.TotalMS, probe.FullStdout)
	}
	if probe.TotalMS < minTotalMS {
		t.Fatalf("total_ms=%d too short for streaming probe (want >=%d)", probe.TotalMS, minTotalMS)
	}
	gap := probe.TotalMS - probe.FirstByteMS
	if gap < minLeadMS {
		t.Fatalf("stdout not incremental: first_byte_ms=%d total_ms=%d gap_ms=%d (want gap >= %d; batch-at-end would buffer until scan finishes)",
			probe.FirstByteMS, probe.TotalMS, gap, minLeadMS)
	}

	want := resolveScanPath(t, firstMainPath)
	// First printed content must be the first discovered main (path line prefix).
	firstLine := probe.FirstChunk
	if i := strings.IndexByte(firstLine, '\n'); i >= 0 {
		firstLine = firstLine[:i]
	}
	firstLine = strings.TrimSpace(firstLine)
	if firstLine != want {
		t.Fatalf("first stdout path should be first main %q, got first_line=%q first_chunk=%q",
			want, firstLine, probe.FirstChunk)
	}
}

func ensureScanStreamProbeHelpersUsed() {
	_ = seedScanPaddingDirs
	_ = runScanGitReposStreamProbe
	_ = assertScanStreamsIncrementally
}
```
