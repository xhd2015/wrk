---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; default-named worktree under WRK_HOME (no task slug).
- Stderr has no treat-as-task messaging.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	want := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertStdoutExactPath(t, resp.Stdout, want)
	assertFileExists(t, want)
	assertGitFileIsWorktreeLink(t, want)
	br := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.MainRepo, br)
	assertBranchCheckedOutInWorktree(t, want, br)
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "treat as") || strings.Contains(low, "looks like a task") {
		t.Fatalf("existing source must not trigger treat-as-task; stderr=%q", resp.Stderr)
	}
}
```
