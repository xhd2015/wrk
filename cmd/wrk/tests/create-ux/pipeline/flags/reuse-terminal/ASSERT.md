---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- iTerm script is reuse-current mode (matchingSession).
- No space; no agent.

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
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeReuse(t, script)
	assertAgentRunNotInvoked(t, req)
}
```
