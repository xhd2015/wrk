## Expected

- Exit code 0.
- A **new** worktree is created under `{WorkRoot}/target/` (named subdir).
- Because preferred branch `main-{date}` is already checked out in the WRK_HOME prior WT,
  create under target uses joint path+branch suffix: `…/target/myrepo-main-{date}-1`
  and branch `main-{date}-1`.
- Prior WRK_HOME worktree still present; stdout is the **new** path (not the WRK_HOME path).
- No Policy B banner/prompt (`would reuse`, `skip creating`, `already has a linked worktree`,
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

	// Branch main-{date} taken by WRK_HOME prior → joint -1 under existing target.
	wantNew := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-1")
	assertStdoutExactPath(t, resp.Stdout, wantNew)
	assertFileExists(t, wantNew)
	assertGitFileIsWorktreeLink(t, wantNew)
	assertBranchCheckedOutInWorktree(t, wantNew, branchName("main", wrkDate, 1))
	assertWorktreeListContains(t, req.TargetDir, wantNew)

	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.TargetDir, req.WtDir)

	// Must not treat other-parent WT as Policy B reuse primary.
	if strings.TrimSpace(resp.Stdout) == req.WtDir {
		t.Fatalf("stdout must not be other-parent path %q", req.WtDir)
	}
	assertNoPolicyBBanner(t, resp.Stdout+resp.Stderr)
}
```
