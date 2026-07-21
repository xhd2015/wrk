## Expected

- Exit code 0.
- Stdout contains `worktree removed:`.
- Worktree directory no longer exists.
- Branch `main-{date}` deleted from main repo.
- Stderr does NOT contain `no go.mod found`.
- Main repo still has its initial commit (`README.md`).

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	assertNotContains(t, resp.Stderr, "no go.mod found")
	assertContains(t, resp.Stdout, "worktree removed:")
	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "README.md"))
}
```
