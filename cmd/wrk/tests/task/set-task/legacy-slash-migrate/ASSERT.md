## Expected

- Exit code 0.
- Stdout is `{WRK_HOME}/worktrees/myrepo-feature-foo-{date}-bar`.
- New branch is `feature-foo-{date}-bar` (**no** `/`).
- Legacy branch `feature/foo-{date}` is gone.
- Old path is gone; invariant holds on the new path.

## Exit Code

- 0

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

	token := sanitizeBranchToken("feature/foo")
	slug := slugify("bar")
	wantPath := worktreePathWithTask(req.WrkHome, "myrepo", token, wrkDate, slug, 0)
	// New branch uses sanitized token (no slash), not legacy feature/foo-…
	wantBranch := branchNameWithTask(token, wrkDate, slug, 0)

	gotPath := strings.TrimSpace(resp.Stdout)
	if gotPath != wantPath {
		t.Fatalf("stdout: expected %q, got %q", wantPath, gotPath)
	}

	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	assertBranchNotExists(t, req.MainRepo, "feature/foo-"+wrkDate)
	assertBranchNotExists(t, req.MainRepo, "feature/foo-"+wrkDate+"-"+slug)
	assertBranchExists(t, req.MainRepo, wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)
	assertWorktreeListContains(t, req.MainRepo, wantPath)

	if filepath.Base(wantPath) != "myrepo-"+wantBranch {
		t.Fatalf("invariant broken: base(%q) != myrepo-%s", wantPath, wantBranch)
	}
}
```
