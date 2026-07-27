## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-linked-side-2026-06-30` (basename from main repo `myrepo`, not `linked-wt`).
- Branch `linked-side-2026-06-30` is created (distinct from source worktree branch `linked-side`).
- New worktree is registered in `git worktree list` from the main repo.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", "linked-side", wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	assertBranchExists(t, req.RepoDir, branchName("linked-side", wrkDate, 0))
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("linked-side", wrkDate, 0))

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	assertWorktreeListContains(t, mainRepo, wantPath)
}
```