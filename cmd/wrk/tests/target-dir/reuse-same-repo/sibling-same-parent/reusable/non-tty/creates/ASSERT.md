## Expected

- Exit code 0 (create, **not** refuse).
- Stdout is new path `{WorkRoot}/target/myrepo-main-{date}`.
- New linked worktree exists; prior reusable sibling still present.
- Stderr does **not** contain refuse wording: `refusing non-interactive`, require TTY /
  re-run in a TTY for Policy B.
- Stderr does **not** contain interactive skip prompt tokens (`skip creating`, `would reuse`)
  on the non-TTY create path (create proceeds quietly).

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
		t.Fatalf("expected create (exit 0) on non-TTY with reusable sibling; got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	wantNew := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate)
	assertStdoutExactPath(t, resp.Stdout, wantNew)
	assertFileExists(t, wantNew)
	assertGitFileIsWorktreeLink(t, wantNew)
	assertBranchCheckedOutInWorktree(t, wantNew, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.TargetDir, wantNew)

	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.TargetDir, req.WtDir)

	// Must not refuse or prompt on non-TTY.
	assertNotContains(t, resp.Stderr, "refusing non-interactive")
	assertNotContains(t, resp.Stderr, "re-run in a TTY")
	assertNotContains(t, resp.Stderr, "skip creating")
	assertNotContains(t, resp.Stderr, "would reuse")
	// Soft guard: no hard refuse mentioning TTY for Policy B.
	if strings.Contains(resp.Stderr, "non-interactive") && strings.Contains(resp.Stderr, "skip") {
		t.Fatalf("stderr still looks like legacy non-TTY refuse: %q", resp.Stderr)
	}
}
```
