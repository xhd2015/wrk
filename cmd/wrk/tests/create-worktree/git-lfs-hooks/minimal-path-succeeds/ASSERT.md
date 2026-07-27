---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Non-zero exit when process `PATH` is only `/usr/bin:/bin` even though `git-lfs` exists under `$HOME/.local/bin`.
- `wrk` passes the stripped `PATH` through to `git worktree add`; the LFS post-checkout hook fails because `git-lfs` is not on `PATH`.
- Stderr reports `git worktree add: exit status 2` and the missing `git-lfs` message.
- No worktree path printed on stdout (git may leave a partial worktree registered after hook failure).

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
		t.Fatalf("expected non-zero exit when git-lfs is not on stripped PATH, got 0 stdout=%q", resp.Stdout)
	}

	assertContains(t, resp.Stderr, "git worktree add: exit status 2")
	assertContains(t, resp.Stderr, "git-lfs")
	assertContains(t, resp.Stderr, "not found on your path")

	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout should be empty on failure, got %q", resp.Stdout)
	}
}
```