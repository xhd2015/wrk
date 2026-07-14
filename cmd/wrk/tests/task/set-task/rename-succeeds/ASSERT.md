## Expected

- Exit code 0.
- Stdout equals new worktree directory path.
- Old worktree directory no longer exists.
- New worktree directory exists and is a linked git worktree.
- Old branch `main-{date}-original-task` removed from main repo.
- New branch `main-{date}-new-task` exists in main repo.

```go
import (
	"path/filepath"
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
	wantPath := worktreePathWithTask(req.WrkHome, "myrepo", "main", wrkDate, slug, 0)
	gotPath := strings.TrimSpace(resp.Stdout)
	if gotPath != wantPath {
		t.Fatalf("stdout: expected %q, got %q", wantPath, gotPath)
	}

	// Old path should no longer exist
	assertFileNotExists(t, req.WtDir)

	// New path should exist and be a linked worktree
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	// Branch renamed in main repo
	oldBranch := branchNameWithTask("main", wrkDate, slugify("original task"), 0)
	newBranch := branchNameWithTask("main", wrkDate, slug, 0)
	assertBranchNotExists(t, req.MainRepo, oldBranch)
	assertBranchExists(t, req.MainRepo, newBranch)

	// New worktree listed in main repo
	assertWorktreeListContains(t, req.MainRepo, wantPath)
}
```
