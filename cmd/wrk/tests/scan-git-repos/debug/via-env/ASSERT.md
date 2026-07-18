## Expected

- Exit code **0**.
- Stdout prints the already-known main path exactly once (always-print).
- One `source=scan` projects.json entry for `myrepo`.
- **Stderr contains `scan:`** and **`mode=`** (`mode=warm` preferred after seed; `mode=cold` OK).
- No `-v` was passed: any major-git `[timestamp] $ git …` lines are not required; `scan:` must still appear from env-driven Debug alone.
- Optional (do not fail if absent): `record known=N newly=M`.
- Phase-level `scan:` line volume (not per-dir spam).

## Side Effects

- Product cache under FakeHome may update; projects stay idempotent.

## Errors

- None on the success path.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("via-env second scan: exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolveScanPath(t, req.MainRepo)
	n := countScanStdoutPathLines(t, resp.Stdout, want)
	if n != 1 {
		t.Fatalf("second scan must always-print known main once; count=%d stdout=%q want=%q", n, resp.Stdout, want)
	}
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")

	// Sanity: Args must not include -v/--verbose for this leaf.
	for _, a := range req.Args {
		if a == "-v" || a == "--verbose" {
			t.Fatalf("via-env must not pass verbose flags, got Args=%v", req.Args)
		}
	}

	stderr := resp.Stderr
	if !strings.Contains(stderr, "scan:") {
		t.Fatalf("WRK_SCAN_DEBUG=1 must wire Options.Debug so stderr contains scan:, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "mode=") {
		t.Fatalf("env debug stderr must include mode= (warm|cold), got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "mode=warm") && !strings.Contains(stderr, "mode=cold") {
		t.Fatalf("env debug stderr mode= value should be warm or cold, got:\n%s", stderr)
	}

	scanLines := 0
	for _, line := range strings.Split(stderr, "\n") {
		if strings.Contains(line, "scan:") {
			scanLines++
		}
	}
	if scanLines > 40 {
		t.Fatalf("too many scan: lines (%d); want phase-level logs, got:\n%s", scanLines, stderr)
	}
}
```
