## Expected

- Exit code 0.
- Stdout (trimmed) equals `{WorkRoot}/wt` exactly — fixed path is **not** renamed.
- Branch checked out in the worktree is `main-2026-06-30-1` (suffix on branch only).
- Branch `main-2026-06-30` still exists as the pre-created ref (not checked out into this worktree).
- Always-new-branch: worktree uses a **new** `-b` branch, not a second checkout of the pre-existing branch.
- Worktree is NOT under `{WRK_HOME}/worktrees`.

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

	wantPath := filepath.Join(req.WorkRoot, "wt")
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	preBranch := branchName("main", wrkDate, 0)
	wantBranch := branchName("main", wrkDate, 1)
	assertBranchExists(t, req.TargetDir, preBranch)
	assertBranchExists(t, req.TargetDir, wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)
	assertWorktreeListContains(t, req.TargetDir, wantPath)

	// worktree must NOT live under WRK_HOME
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 1))
}
```
