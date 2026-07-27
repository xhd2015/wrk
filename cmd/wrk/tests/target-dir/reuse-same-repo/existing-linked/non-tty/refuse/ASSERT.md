## Expected

- Non-zero exit.
- Stdout empty.
- Stderr contains refuse wording: `already has a linked worktree`, basename, existing path, non-interactive / TTY, and that default is skip.
- Shape: `wrk: myrepo already has a linked worktree at <absPath>; refusing non-interactive create (default is skip; re-run in a TTY)`
- No new worktree under `{WorkRoot}/target/`.
- Prior worktree unchanged.
- No ANSI color required (non-TTY harness; plain refuse string).

## Exit Code

- non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit on non-TTY named bring with existing linked WT; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout must be empty, got %q", resp.Stdout)
	}

	assertContains(t, resp.Stderr, "already has a linked worktree")
	assertContains(t, resp.Stderr, req.WtDir)
	assertContains(t, resp.Stderr, "myrepo")
	// Non-interactive refuse (stable substrings from design).
	if !strings.Contains(resp.Stderr, "non-interactive") && !strings.Contains(resp.Stderr, "TTY") && !strings.Contains(resp.Stderr, "tty") {
		t.Fatalf("stderr should mention non-interactive/TTY refuse; got %q", resp.Stderr)
	}
	assertContains(t, resp.Stderr, "skip")

	assertFileExists(t, req.WtDir)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate))
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-1"))
}
```
