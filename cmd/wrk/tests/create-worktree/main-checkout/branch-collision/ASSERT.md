## Expected

- Exit code 0.
- Date-suffixed path `{WRK_HOME}/worktrees/myrepo-main-2026-06-30` is **not** created (branch `main-2026-06-30` already exists → skip 0).
- Stdout (trimmed) equals `{WRK_HOME}/worktrees/myrepo-main-2026-06-30-1`.
- Branch `main-2026-06-30-1` is a **new** ref checked out in the new worktree (always `worktree add -b`; never reuses the pre-existing branch).
- Pre-existing orphan ref `main-2026-06-30` still exists (not checked out into a second worktree).

## Exit Code

- 0

```go
import (
	"os"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	dateSuffixed := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	if _, err := os.Stat(dateSuffixed); err == nil {
		t.Fatalf("date-suffixed path %q should not exist", dateSuffixed)
	}

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 1)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchExists(t, req.RepoDir, branchName("main", wrkDate, 0))
	assertBranchExists(t, req.RepoDir, branchName("main", wrkDate, 1))
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("main", wrkDate, 1))
	assertWorktreeListContains(t, req.RepoDir, wantPath)
}
```
