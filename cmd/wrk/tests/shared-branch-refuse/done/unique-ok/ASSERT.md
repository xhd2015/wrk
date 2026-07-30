## Expected

- Exit code 0.
- Worktree directory removed; branch deleted; `feature-work` present on main.
- Stdout indicates successful done (e.g. `worktree removed:` or `merged branch`).

## Side Effects

- Merge-back with remove completed for a unique branch (happy path pin).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("unique-branch --done should succeed; exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	// Must not look like shared-branch refuse.
	if strings.Contains(resp.Stderr, "multiple worktrees") ||
		(strings.Contains(resp.Stderr, "Error:") && strings.Contains(strings.ToLower(resp.Stderr), "refuse")) {
		t.Fatalf("unique-branch must not refuse as shared; stderr=%q", resp.Stderr)
	}
	assertFileNotExists(t, req.WtDir)
	assertBranchNotExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListNotContains(t, req.MainRepo, req.WtDir)
	assertFileExists(t, filepath.Join(req.MainRepo, "feature-work"))
	combined := resp.Stdout + resp.Stderr
	if !strings.Contains(combined, "worktree removed:") && !strings.Contains(combined, "merged branch") {
		t.Fatalf("expected done success message; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
