
## Expected

- Exit code 0.
- Stdout is the new worktree path.
- Old worktree path gone; new path exists with renamed branch.

## Exit Code

- 0

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	slug := slugify("new task")
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
