## Expected

- Exit 0; worktree exactly at `{WorkRoot}/wt with spaces`.
- Branch includes slug from `-t "other task"` (`other-task`), not from the path text.
- Stderr has no treat-as-task prompt for the second positional.

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
	want := filepath.Join(req.WorkRoot, "wt with spaces")
	assertStdoutExactPath(t, resp.Stdout, want)
	assertFileExists(t, want)
	assertGitFileIsWorktreeLink(t, want)
	br := branchNameWithTask("main", wrkDate, slugify("other task"), 0)
	assertBranchExists(t, req.MainRepo, br)
	assertBranchCheckedOutInWorktree(t, want, br)
	low := strings.ToLower(resp.Stderr)
	if strings.Contains(low, "treat as") {
		t.Fatalf("-t already set: no treat-as-task for second positional; stderr=%q", resp.Stderr)
	}
	// Must not also create a WRK_HOME path from promoting the second arg.
	assertFileNotExists(t, wantPromotedWorktree(req, "wt with spaces"))
}
```
