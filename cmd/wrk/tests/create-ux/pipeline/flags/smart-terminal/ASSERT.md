---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- iTerm smart script (matchingWindow + create tab).
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
	assertItermModeSmart(t, script)
	assertAgentRunNotInvoked(t, req)
}
```
