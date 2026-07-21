## Expected

- Exit code 0.
- Combined output contains `Proceed?` (opt-in interactive path).
- Worktree removed; main has `feature-work`; branch deleted.

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

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "Proceed?")
	assertContains(t, resp.Stdout, "merged branch")
	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
}
```
