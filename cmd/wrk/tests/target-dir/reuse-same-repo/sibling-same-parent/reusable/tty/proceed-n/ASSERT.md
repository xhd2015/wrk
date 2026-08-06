---
label: e2e, tty
explanation: requires `script` fake TTY for Policy B skip prompt; platform-specific
---

## Expected

- Exit code 0.
- A **new** worktree is created at `{WorkRoot}/target/myrepo-main-{date}` (preferred
  branch free — sibling used a distinct branch name).
- Prior reusable sibling still exists.
- Combined output showed the Policy B skip prompt before create (`would reuse`,
  `skip creating`, `wrk: warning:`, sibling path).

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
	got := strings.TrimSpace(resp.Stdout)
	if got != wantNew && !strings.Contains(resp.Stdout, wantNew) {
		t.Fatalf("stdout should be/include new path %q; stdout=%q stderr=%q", wantNew, resp.Stdout, resp.Stderr)
	}

	assertFileExists(t, wantNew)
	assertGitFileIsWorktreeLink(t, wantNew)
	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.TargetDir, wantNew)
	assertWorktreeListContains(t, req.TargetDir, req.WtDir)
	assertBranchCheckedOutInWorktree(t, wantNew, branchName("main", wrkDate, 0))

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "wrk: warning:")
	assertContains(t, combined, "would reuse")
	assertContains(t, combined, "skip creating")
	assertContains(t, combined, req.WtDir)
}
```
