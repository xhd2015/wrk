## Expected

- Exit code 0.
- Stdout contains `worktree removed:`.
- Last stdout line is the main repo absolute path (`pwd` ran in main, not the removed worktree).
- Worktree directory no longer exists; branch deleted; main still has merged content.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assertContains(t, resp.Stdout, "worktree removed:")
	assertLastStdoutLine(t, resp.Stdout, req.MainRepo)

	// pwd must not be the removed worktree path alone as last line.
	if strings.TrimSpace(resp.Stdout) == req.WtDir {
		t.Fatalf("stdout is only removed worktree path; exec should run in main")
	}

	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
}
```
