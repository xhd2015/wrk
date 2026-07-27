---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- Stdout is exactly `{WorkRoot}/out with spaces` (fixed target-dir).
- Linked worktree there; branch default `main-{date}` (no task slug required).
- No WRK_HOME default-named worktree for promoted task.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	want := filepath.Join(req.WorkRoot, "out with spaces")
	assertStdoutExactPath(t, resp.Stdout, want)
	assertFileExists(t, want)
	assertGitFileIsWorktreeLink(t, want)
	br := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.MainRepo, br)
	assertBranchCheckedOutInWorktree(t, want, br)
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))
	assertFileNotExists(t, wantPromotedWorktree(req, "out with spaces"))
}
```
