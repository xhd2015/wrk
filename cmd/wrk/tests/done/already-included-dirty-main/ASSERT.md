## Expected

- Exit code 0.
- Stdout contains `worktree removed:`.
- Stderr does not mention `main-sync`.
- Worktree directory no longer exists; branch deleted.
- Main still has the merged `feature-work` file and uncommitted `dirty-main.txt`.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	assertContains(t, resp.Stdout, "worktree removed:")
	if strings.Contains(resp.Stderr, "main-sync") {
		t.Fatalf("did not expect main-sync; stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}
	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	assertFileExists(t, filepath.Join(req.MainRepo, "dirty-main.txt"))

	status := gitOutputIsolated(t, req.MainRepo, "status", "--porcelain")
	if !strings.Contains(status, "dirty-main.txt") {
		t.Fatalf("main should stay dirty; porcelain=%q", status)
	}
}
```
