## Expected Output

```text
Multiple projects match "myrepo":
  1) <sorted-path-1>
  2) <sorted-path-2>
Select [1-2]:
```

## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-main-2026-06-30`.
- Worktree created from the **selected** saved repo (`zzz/myrepo`, index 2).
- No worktree registered under the unselected repo.

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
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	sorted := sortedSavedPaths(t, req.MainRepo, req.SecondRepo)
	tmpl := `<contains>
Multiple projects match "myrepo":
  1) ` + sorted[0] + `
  2) ` + sorted[1] + `
Select [1-2]:
</contains>`
	assert.Output(t, resp.Stderr, tmpl)

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchExists(t, req.SelectedSavedRepo, branchName("main", wrkDate, 0))
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.SelectedSavedRepo, wantPath)

	unselected := req.MainRepo
	if req.SelectedSavedRepo == req.MainRepo {
		unselected = req.SecondRepo
	}
	assertWorktreeListNotContains(t, unselected, wantPath)
}
```