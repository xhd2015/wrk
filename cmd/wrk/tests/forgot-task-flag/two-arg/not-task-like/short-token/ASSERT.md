## Expected

- Exit 0; worktree at `{WorkRoot}/out`.
- No task-like messaging; not WRK_HOME default naming.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	want := filepath.Join(req.WorkRoot, "out")
	assertStdoutExactPath(t, resp.Stdout, want)
	assertFileExists(t, want)
	assertGitFileIsWorktreeLink(t, want)
	br := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.MainRepo, br)
	assertBranchCheckedOutInWorktree(t, want, br)
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "treat as") || strings.Contains(low, "looks like a task") {
		t.Fatalf("short token must not trigger treat-as-task; stderr=%q", resp.Stderr)
	}
}
```
