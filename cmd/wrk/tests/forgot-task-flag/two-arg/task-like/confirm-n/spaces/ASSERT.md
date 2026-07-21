## Expected

- Exit 0.
- Worktree created at `{WorkRoot}/out with spaces` (fixed target-dir).
- Stdout may include `script(1)` TTY noise; path may appear in combined output.
- Linked worktree there; branch default `main-{date}` (no task slug required).
- No WRK_HOME default-named worktree for promoted task.

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
		t.Fatalf("exit %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := filepath.Join(req.WorkRoot, "out with spaces")
	// script(1) may pollute stdout; require the worktree on disk (primary contract).
	assertFileExists(t, want)
	assertGitFileIsWorktreeLink(t, want)
	if !strings.Contains(resp.Stdout, want) && strings.TrimSpace(resp.Stdout) != want {
		// Soft check only when path never appears; still OK if disk state is correct.
		_ = resp.Stdout
	}
	br := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.MainRepo, br)
	assertBranchCheckedOutInWorktree(t, want, br)
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))
	assertFileNotExists(t, wantPromotedWorktree(req, "out with spaces"))
}
```
