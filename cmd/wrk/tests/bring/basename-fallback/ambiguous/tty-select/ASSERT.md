## Expected Output

```text
Multiple projects match "mydep":
  1) <sorted-path-1>
  2) <sorted-path-2>
Select [1-2]:
```

## Expected

- Exit code 0.
- Stderr shows the candidate listing and **exactly one** `Select [1-2]:` (preflight resolve-once; apply must not re-prompt).
- Stderr contains **exactly one** `Multiple projects match "mydep":` block.
- Stdin is a single line `2\n` — sufficient for success (no second read).
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}` (one external path line only).
- External worktree created from the **selected** saved dep (`zzz/mydep`, index 2).
- No external worktree registered under the unselected dep repo.

## Side Effects

- TTY prompt shown once (or simulated via `WRK_BASENAME_CONFIRM=1`); stdin index selects candidate.
- Resolved abs path is reused for materialize — no second basename prompt.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	// P1 contract: preflight resolves once — one Select, one listing (not checkDuplicate + bringOne).
	if n := strings.Count(resp.Stderr, "Select ["); n != 1 {
		t.Fatalf("expected exactly one Select prompt, got %d stderr=%q", n, resp.Stderr)
	}
	if n := strings.Count(resp.Stderr, `Multiple projects match "mydep":`); n != 1 {
		t.Fatalf("expected exactly one match listing, got %d stderr=%q", n, resp.Stderr)
	}
	// P2 multi-only: single-arg bring must not print will bring: plan (noise reduction).
	assertNotContains(t, resp.Stderr, "will bring:")

	sorted := sortedSavedPaths(t, req.DepPath, req.SecondRepo)
	tmpl := `<contains>
Multiple projects match "mydep":
  1) ` + sorted[0] + `
  2) ` + sorted[1] + `
Select [1-2]:
</contains>`
	assert.Output(t, resp.Stderr, tmpl)

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.SelectedSavedRepo, wantPath)

	unselected := req.DepPath
	if req.SelectedSavedRepo == req.DepPath {
		unselected = req.SecondRepo
	}
	assertWorktreeListNotContains(t, unselected, wantPath)
}
```
