---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Non-zero exit when cwd differs from the source repo and `PATH` is stripped (same as `wrk $X/agent-pro` with missing `git-lfs` on `PATH`).
- Foreign cwd does not change the failure: `git worktree add` still inherits the stripped `PATH` and the LFS post-checkout hook errors.
- Stderr reports `git worktree add: exit status 2` and the missing `git-lfs` message.
- No worktree path printed on stdout.

## Exit Code

- 1

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit from foreign cwd with stripped PATH, got 0 stdout=%q", resp.Stdout)
	}

	assertContains(t, resp.Stderr, "git worktree add: exit status 2")
	assertContains(t, resp.Stderr, "git-lfs")
	assertContains(t, resp.Stderr, "not found on your path")

	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout should be empty on failure, got %q", resp.Stdout)
	}
}
```