---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; stdout is worktree path + `\n`.
- Worktree exists under `{WRK_HOME}/worktrees/`.
- Space, iterm, and agent-run mocks are not invoked.

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	wt := wantCreateUXWorktree(req)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertSpaceNotInvoked(t, req)
	assertItermNotInvoked(t, req)
	assertAgentRunNotInvoked(t, req)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
