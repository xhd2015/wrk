## Expected

- Exit code **0**.
- Stdout is exactly the **Projects** main absolute path (single line + trailing
  `\n`) — **not** `home-main`.
- `projects.json` prints only under Projects; projects.json unchanged
  (count 1 after clear + filtered scan).
- Product **home** universe index still exists at
  `{FakeHome}/.cache/git-repo-scan/home/repos.json` after the Projects scan
  (shared home cache files; not abandoned for a Projects-only universe).
- Home index still lists (or continues to list) fixtures under home universe —
  at least `proj-main`; preferably still retains `home-main` from the full-home
  seed (merge preserve outside filter).
- Must **not** promote Projects-only into a separate primary universe that
  drops home cache: `home/repos.json` remains the durable home-universe file.

## Side Effects

- Product cache under `{FakeHome}/.cache/git-repo-scan/home/` is reused/updated.
- `home-main` is not newly printed or newly added in this filtered run.

## Errors

- None on the success path.

## Exit Code

- 0

```go
import (
	"os"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("Projects filter scan: exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	projMain := resolveScanPath(t, req.MainRepo)
	homeMain := resolveScanPath(t, req.SecondRepo)

	// Stdout: only Projects main (filter emit).
	assertStdoutExactPath(t, resp.Stdout, projMain)
	if strings.Contains(resp.Stdout, homeMain) {
		t.Fatalf("stdout must not include home-main outside Projects filter; stdout=%q",
			resp.Stdout)
	}

	// Print-only: scan never mutates projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)

	// Same home universe cache files still present after subpath scan.
	indexPath := productHomeUniverseReposJSON(req.FakeHome)
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("Projects scan must share home universe cache file %s: %v", indexPath, err)
	}
	if !homeUniverseIndexHasPath(t, req.FakeHome, projMain) {
		t.Fatalf("home universe index should still list proj-main %q after Projects scan",
			projMain)
	}
	// Prefer merge preserve of home-main outside the filter root (library P2 merge).
	if !homeUniverseIndexHasPath(t, req.FakeHome, homeMain) {
		t.Fatalf("home universe index should retain home-main %q from full-home seed "+
			"(shared home cache merge; not wipe outside filter)", homeMain)
	}
}
```
