## Expected

- Exit code 0.
- Stdout is the free path with `-1` suffix: `{WRK_HOME}/worktrees/myrepo-main-{date}-new-task-1`.
- Old worktree path is gone.
- Branch renamed to `main-{date}-new-task-1`.
- Preferred unsuffixed path still exists as the pre-created blocker (not a worktree).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	slug := slugify("new task")
	blockedPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slug, 0)
	wantPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slug, 1)
	gotPath := strings.TrimSpace(resp.Stdout)
	if gotPath != wantPath {
		t.Fatalf("stdout: expected %q, got %q", wantPath, gotPath)
	}

	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	// Blocker dir still present (not consumed as the worktree).
	assertFileExists(t, blockedPath)

	oldBranch := branchNameWithTask("main", wrkDate, slugify("original task"), 0)
	newBranch := branchNameWithTask("main", wrkDate, slug, 1)
	assertBranchNotExists(t, req.MainRepo, oldBranch)
	assertBranchExists(t, req.MainRepo, newBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, newBranch)
	assertWorktreeListContains(t, req.MainRepo, wantPath)
}
```
