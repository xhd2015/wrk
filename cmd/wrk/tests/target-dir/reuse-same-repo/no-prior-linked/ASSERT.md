## Expected

- Exit code 0.
- Stdout is `{WorkRoot}/target/myrepo-main-{WRK_DATE}`.
- New linked worktree exists at that path with branch `main-{date}`.
- Stderr does **not** contain Policy B skip prompt tokens: `already has a linked worktree`, `skip creating`, or legacy `already exists` / `skip?`.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.TargetDir, wantPath)

	// Policy B must not fire when source has no prior linked worktree.
	assertNotContains(t, resp.Stderr, "already has a linked worktree")
	assertNotContains(t, resp.Stderr, "skip creating")
	assertNotContains(t, resp.Stderr, "already exists")
	assertNotContains(t, resp.Stderr, "skip?")
}
```
