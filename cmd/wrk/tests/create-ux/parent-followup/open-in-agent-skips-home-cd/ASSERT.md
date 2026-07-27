---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; stdout is worktree path + trailing `\n`.
- Outer agent-run invoked with `--dir` targeting worktree.
- Follow-up file empty (agent UX skips home-gated parent cd).

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
	assertAgentRunInvoked(t, req, wt, req.TaskDesc)
	assertFollowupEmptyUX(t, req)
}
```
