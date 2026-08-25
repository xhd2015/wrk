---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; worktree created.
- Follow-up has agent-run line only (no `cd`).
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
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertAgentRunNotInvoked(t, req)
	assertFollowupHereUX(t, req, wt, req.TaskDesc, false)
}
```
