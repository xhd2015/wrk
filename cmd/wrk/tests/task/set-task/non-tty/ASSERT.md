## Expected

- Exit code 0 (default auto-yes; no TTY / no WRK_SET_TASK_CONFIRM required).
- Stdout is the new worktree path.
- Old path gone; new path exists with renamed branch.

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
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	slug := slugify("new task desc")
	wantPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slug, 0)
	gotPath := strings.TrimSpace(resp.Stdout)
	if gotPath != wantPath {
		t.Fatalf("stdout: expected %q, got %q", wantPath, gotPath)
	}

	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	oldBranch := branchNameWithTask("main", wrkDate, slugify("original task"), 0)
	newBranch := branchNameWithTask("main", wrkDate, slug, 0)
	assertBranchNotExists(t, req.MainRepo, oldBranch)
	assertBranchExists(t, req.MainRepo, newBranch)
}
```
