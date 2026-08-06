## Expected

- Exit code 0.
- Stdout is `{WorkRoot}/target/myrepo-main-{WRK_DATE}`.
- New linked worktree exists at that path with branch `main-{date}`.
- Stderr does **not** contain Policy B skip/reuse tokens: `would reuse`, `skip creating`,
  `already has a linked worktree`, `refusing non-interactive`.

## Exit Code

- 0

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate)
	assertStdoutExactPath(t, resp.Stdout, wantPath)
	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertBranchCheckedOutInWorktree(t, wantPath, branchName("main", wrkDate, 0))
	assertWorktreeListContains(t, req.TargetDir, wantPath)

	assertNoPolicyBBanner(t, resp.Stdout+resp.Stderr)
}
```
