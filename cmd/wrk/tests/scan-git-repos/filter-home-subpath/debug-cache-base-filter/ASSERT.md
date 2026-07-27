## Expected

- Exit code **0**.
- Stdout prints the already-known Projects main path exactly once (always-print; not a new record).
- Stderr contains greppable **`scan:`** phase lines (Debug wired).
- Stderr contains greppable **`cache_base`** (product/library debug for the
  home-universe cache base used for this root).
- Stderr contains greppable **`filter`** (emit filter for the Projects subpath
  under home).
- Prefer also `mode=` warm|cold (optional soft — do not fail if only cache_base
  + filter + scan: are present).
- Phase-level volume: modest count of `scan:` lines (not per-directory spam).

## Side Effects

- Product cache under FakeHome may update; projects stay at 2 entries.

## Errors

- None on the success path.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("Projects -v debug: exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	// Always-print: seed already recorded both mains; this leaf scans ~/Projects only.
	// Expect proj-main once; home-main is outside the Projects filter and must not appear.
	wantProj := resolveScanPath(t, filepath.Join(req.FakeHome, "Projects", "proj-main"))
	nProj := countScanStdoutPathLines(t, resp.Stdout, wantProj)
	if nProj != 1 {
		t.Fatalf("second scan must always-print known Projects main once; count=%d stdout=%q want=%q",
			nProj, resp.Stdout, wantProj)
	}
	assertScanProjectsCount(t, req.WrkHome, 2)

	stderr := resp.Stderr
	if !strings.Contains(stderr, "scan:") {
		t.Fatalf("-v must enable Debug so stderr contains scan: prefix, got:\n%s", stderr)
	}
	// P5: product/library debug must surface cache_base + filter for two-base mapping.
	if !strings.Contains(stderr, "cache_base") {
		t.Fatalf("-v debug stderr must include cache_base (home universe cache base), got:\n%s",
			stderr)
	}
	if !strings.Contains(stderr, "filter") {
		t.Fatalf("-v debug stderr must include filter (Projects emit filter), got:\n%s",
			stderr)
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
