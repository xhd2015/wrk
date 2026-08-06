## Expected

- Exit code 0.
- Stdout is new path `{WorkRoot}/target/myrepo-main-{date}` (preferred branch free).
- New linked worktree exists; dirty sibling still present and dirty.
- No Policy B banner (`would reuse`, `skip creating`, `already has a linked worktree`,
  `refusing non-interactive`).

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
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantNew := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate)
	assertStdoutExactPath(t, resp.Stdout, wantNew)
	assertFileExists(t, wantNew)
	assertGitFileIsWorktreeLink(t, wantNew)
	assertBranchCheckedOutInWorktree(t, wantNew, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.TargetDir, wantNew)

	assertFileExists(t, req.WtDir)
	assertFileExists(t, filepath.Join(req.WtDir, "dirty-untracked.txt"))
	assertWorktreeListContains(t, req.TargetDir, req.WtDir)

	if strings.TrimSpace(resp.Stdout) == req.WtDir {
		t.Fatalf("stdout must not reuse dirty sibling %q", req.WtDir)
	}
	assertNoPolicyBBanner(t, resp.Stdout+resp.Stderr)
}
```
