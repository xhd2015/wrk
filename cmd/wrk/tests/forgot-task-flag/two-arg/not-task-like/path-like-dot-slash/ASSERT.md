---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; worktree at `{WorkRoot}/real-target`.
- Stderr has no task-like treat-as-task messaging.
- Not under WRK_HOME default path.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	want := filepath.Join(req.WorkRoot, "real-target")
	assertStdoutExactPath(t, resp.Stdout, want)
	assertFileExists(t, want)
	assertGitFileIsWorktreeLink(t, want)
	br := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.MainRepo, br)
	assertBranchCheckedOutInWorktree(t, want, br)
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "treat as") || strings.Contains(low, "looks like a task") {
		t.Fatalf("path-like must not trigger treat-as-task; stderr=%q", resp.Stderr)
	}
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))
}
```
