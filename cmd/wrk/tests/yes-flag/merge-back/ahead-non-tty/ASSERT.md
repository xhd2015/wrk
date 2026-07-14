## Expected

- Exit code 0.
- Worktree directory still exists.
- Main repo contains the merged commit.
- Worktree branch still exists.

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

	assertFileExists(t, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```
