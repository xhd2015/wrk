## Expected

- Exit code 0 (operation aborted cleanly).
- Stdout contains `merge-back aborted`.
- Worktree directory still exists.
- Main repo does NOT have the worktree commit.
- Branch still exists.

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assertContains(t, resp.Stdout, "merge-back aborted")
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```
