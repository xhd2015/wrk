## Expected

- Exit code 0.
- Stdout is `{WRK_HOME}/worktrees/myrepo-main-{date}-new-task-1`.
- Branch is `main-{date}-new-task-1` (preferred branch ref was already taken).
- Old worktree path and old branch are gone.
- Pre-existing preferred branch `main-{date}-new-task` still exists.

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
	wantPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slug, 1)
	gotPath := strings.TrimSpace(resp.Stdout)
	if gotPath != wantPath {
		t.Fatalf("stdout: expected %q, got %q", wantPath, gotPath)
	}

	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	preferredBranch := branchNameWithTask("main", wrkDate, slug, 0)
	newBranch := branchNameWithTask("main", wrkDate, slug, 1)
	oldBranch := branchNameWithTask("main", wrkDate, slugify("original task"), 0)
	assertBranchExists(t, req.MainRepo, preferredBranch)
	assertBranchExists(t, req.MainRepo, newBranch)
	assertBranchNotExists(t, req.MainRepo, oldBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, newBranch)
	assertWorktreeListContains(t, req.MainRepo, wantPath)
}
```
