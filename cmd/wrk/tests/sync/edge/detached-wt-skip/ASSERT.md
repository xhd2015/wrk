---
label: slow
explanation: detached linked worktree skip
---

## Expected Output

```
synced: 0 into main, 0 into worktrees, 1 skipped
```

## Expected

- Exit code 0 (partial skip).
- Stdout is exactly zero-action summary with skipped=1.
- Stderr contains `warning: skip` and `detached HEAD`, and mentions the worktree
  path (absolute or basename of `req.WtPath`).
- Main HEAD unchanged; worktree still at `req.WtSHA`.

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

	wantOut := buildSyncStdout(nil, 0, 0, 1, false)
	assertOutputExact(t, resp.Stdout, syncStdoutV2(wantOut))

	assertContains(t, resp.Stderr, "warning: skip")
	assertContains(t, resp.Stderr, "detached HEAD")
	base := filepath.Base(req.WtPath)
	if !strings.Contains(resp.Stderr, req.WtPath) && !strings.Contains(resp.Stderr, base) {
		t.Fatalf("stderr should mention worktree path %q or base %q; got %q", req.WtPath, base, resp.Stderr)
	}

	assertHEADUnchanged(t, req.MainRepo, req.MainSHA)
	assertHEADUnchanged(t, req.WtPath, req.WtSHA)
}
```
