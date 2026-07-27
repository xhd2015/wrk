## Expected

- Exit code **0**.
- Stdout prints the already-known main path exactly once (always-print).
- projects.json has zero entries (scan is print-only).
- **Stderr contains zero occurrences of the substring `scan:`** (case-sensitive greppable Debug prefix).
- Non-vacuous control: same seed+cache path as on-leaves; silence proves Debug was not wired, not that Scan was skipped.
- Root-error `warning: scan root …` lines (if any) are out of scope; they must not introduce `scan:` phase logs.

## Side Effects

- Cache/projects may still update; discovery proceeds without Debug spam.

## Errors

- None on the success path.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("debug-off second scan: exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := resolveScanPath(t, req.MainRepo)
	n := countScanStdoutPathLines(t, resp.Stdout, want)
	if n != 1 {
		t.Fatalf("second scan must always-print known main once; count=%d stdout=%q want=%q", n, resp.Stdout, want)
	}
	// Print-only: scan never mutates projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)

	for _, a := range req.Args {
		if a == "-v" || a == "--verbose" {
			t.Fatalf("off leaf must not pass verbose flags, got Args=%v", req.Args)
		}
	}

	if strings.Contains(resp.Stderr, "scan:") {
		t.Fatalf("without -v and without WRK_SCAN_DEBUG, stderr must have zero scan: markers, got:\n%s",
			resp.Stderr)
	}
}
```
