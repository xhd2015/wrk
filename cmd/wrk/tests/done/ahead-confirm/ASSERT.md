## Expected

- Exit code 0.
- Stdout contains merge confirmation prompt and success message (`merged branch`).
- Worktree directory no longer exists.
- Main repo has the worktree commit (`feature-work` file).
- Branch `main-{date}` deleted.

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

	assertContains(t, resp.Stdout, "merged branch")
	assertContains(t, resp.Stdout, req.WtBranch)
	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
}
```