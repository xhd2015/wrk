## Expected

- Exit code **0**.
- Stdout empty (main already recorded by seed; no newly-added paths).
- `projects.json` still exactly one `source=scan` entry for `myrepo`.
- **Stderr contains greppable `scan:`** (library Debug wired through wrk).
- **Stderr contains `mode=`** with warm or cold (`mode=warm` preferred after seed; `mode=cold` acceptable if product still cold-walks).
- Optional (do not fail if absent): `record` / `known=` / `newly=` wrk-side summary lines.
- Phase-level volume: count of lines containing `scan:` stays modest (not per-directory spam).

## Side Effects

- Product cache under `{FakeHome}/.cache/git-repo-scan` may be updated by the warm/cold path.
- `-v` may also emit wrk major-git log lines on stderr; those are orthogonal and must not replace the need for `scan:`.

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
		t.Fatalf("via-verbose second scan: exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stdout != "" {
		t.Fatalf("second scan should print no newly recorded paths, got stdout=%q", resp.Stdout)
	}
	assertScanProjectsCount(t, req.WrkHome, 1)
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")

	stderr := resp.Stderr
	if !strings.Contains(stderr, "scan:") {
		t.Fatalf("-v must wire Options.Debug so stderr contains scan: prefix, got:\n%s", stderr)
	}
	if !strings.Contains(stderr, "mode=") {
		t.Fatalf("-v debug stderr must include mode= (warm|cold), got:\n%s", stderr)
	}
	// Prefer warm after seed, but accept cold — mode= presence is the seal.
	if !strings.Contains(stderr, "mode=warm") && !strings.Contains(stderr, "mode=cold") {
		t.Fatalf("-v debug stderr mode= value should be warm or cold, got:\n%s", stderr)
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
