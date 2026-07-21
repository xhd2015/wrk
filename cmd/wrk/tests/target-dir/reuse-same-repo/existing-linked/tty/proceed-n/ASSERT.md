---
label: tty
explanation: requires `script` fake TTY for skip prompt; platform-specific
---

## Expected

- Exit code 0.
- A **new** worktree is created under `{WorkRoot}/target/myrepo-main-{date}-1` (preferred branch taken).
- Prior WRK_HOME worktree still exists.
- Combined output showed the Policy B skip prompt before create (`already has a linked worktree`, `skip creating`).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	// Preferred branch main-{date} is taken by the prior WRK_HOME worktree, so
	// create under existing target dir walks to path+branch -1.
	wantNew := filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-1")
	got := strings.TrimSpace(resp.Stdout)
	if got != wantNew && !strings.Contains(resp.Stdout, wantNew) {
		// Some implementations may put path on its own clean stdout line after prompt noise.
		t.Fatalf("stdout should be/include new path %q; stdout=%q stderr=%q", wantNew, resp.Stdout, resp.Stderr)
	}

	assertFileExists(t, wantNew)
	assertGitFileIsWorktreeLink(t, wantNew)
	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.TargetDir, wantNew)
	assertWorktreeListContains(t, req.TargetDir, req.WtDir)

	assertBranchExists(t, req.TargetDir, branchName("main", wrkDate, 1))
	assertBranchCheckedOutInWorktree(t, wantNew, branchName("main", wrkDate, 1))

	combined := resp.Stdout + resp.Stderr
	assertContains(t, combined, "already has a linked worktree")
	assertContains(t, combined, "skip creating")
	assertContains(t, combined, "wrk: warning:")
	assertContains(t, combined, req.WtDir)
	assertContains(t, combined, "myrepo")
}
```
