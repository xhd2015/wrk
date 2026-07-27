---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- iTerm ForceNew at wt.
- Script must not contain agent-run follow-up.
- Outer agent-run not invoked.
- No space.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertSpaceNotInvoked(t, req)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	if strings.Contains(script, "agent-run") {
		t.Fatalf("--no-open-in-agent should suppress agent follow-up; script:\n%s", script)
	}
	assertAgentRunNotInvoked(t, req)
}
```
