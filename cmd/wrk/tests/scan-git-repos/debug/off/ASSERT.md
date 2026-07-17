## Expected

- Exit code **0**.
- Stdout empty (already known main).
- One `source=scan` projects.json entry for `myrepo`.
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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("debug-off second scan: exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout != "" {
		t.Fatalf("second scan should print no newly recorded paths, got stdout=%q", resp.Stdout)
	}
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")

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
