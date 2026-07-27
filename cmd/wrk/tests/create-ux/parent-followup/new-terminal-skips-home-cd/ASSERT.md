---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; stdout is worktree path + trailing `\n`.
- iTerm invoked at worktree (ForceNew).
- Outer agent-run not invoked.
- Follow-up file empty (terminal UX skips home-gated parent cd).

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
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertAgentRunNotInvoked(t, req)
	assertFollowupEmptyUX(t, req)
}
```
