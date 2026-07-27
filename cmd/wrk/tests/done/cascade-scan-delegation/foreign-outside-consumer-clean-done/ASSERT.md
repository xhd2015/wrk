## Expected

- Exit code 0 (own merge-back succeeds; foreign dirty tree did not block).
- Consumer linked worktree removed; branch deleted from main.
- Stdout contains merge-back / worktree removed language (`merged branch` or
  `worktree removed:`).
- Stderr/stdout do **not** mention the foreign path or `other/external/agent-pro`.
- No `Error:` dirty preflight for the foreign tree.
- Foreign dirty tree still exists on disk (untouched).

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
		t.Fatalf("expected exit 0 (foreign dirty outside must not block --done), got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	assertNoForeignPathInOutput(t, req, resp)

	// Own success: worktree gone; branch removed.
	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)

	stdoutLower := strings.ToLower(resp.Stdout)
	if !strings.Contains(stdoutLower, "merged branch") &&
		!strings.Contains(stdoutLower, "worktree removed") {
		t.Fatalf("expected own merge-back success language; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}

	// Foreign untouched (still dirty on disk).
	assertFileExists(t, req.SecondRepo)
	assertFileExists(t, filepath.Join(req.SecondRepo, "dirty-foreign"))
}
```
