---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected Output

```text
Multiple projects match "mydep":
  1) <sorted-path-1>
  2) <sorted-path-2>
Select [1-2]:
```

## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External worktree created from the **selected** saved dep (`zzz/mydep`, index 2).
- No external worktree registered under the unselected dep repo.

## Side Effects

- TTY prompt shown (or simulated via `WRK_BASENAME_CONFIRM=1`); stdin index selects candidate.

## Exit Code

- 0

```go
import (
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

	sorted := sortedSavedPaths(t, req.DepPath, req.SecondRepo)
	tmpl := `<contains>
Multiple projects match "mydep":
  1) ` + sorted[0] + `
  2) ` + sorted[1] + `
Select [1-2]:
</contains>`
	assert.Output(t, resp.Stderr, tmpl)

	wantPath := externalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
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