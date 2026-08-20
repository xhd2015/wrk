---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; path printed (worktree includes task slug).
- Outer agent-run invoked with `--agent-runner=grok-tty`.
- Last argv element is exactly `/brainstorm fix bug` (unchanged).
- No space; no iterm.

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
	assertSpaceNotInvoked(t, req)
	assertItermNotInvoked(t, req)
	args := assertAgentRunInvokedWith(t, req, wt, req.TaskDesc, "grok-tty", "/brainstorm")
	assertAgentArgvHasDir(t, args, wt)
}
```
