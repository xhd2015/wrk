---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; path printed.
- Space invoked once (`CreateAndActivate`).
- iTerm ForceNew script targets worktree path.
- Outer agent-run not invoked.

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
	assertSpaceInvokedOnce(t, req)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertAgentRunNotInvoked(t, req)
}
```
