---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; worktree created.
- Outer agent-run invoked with `--dir` (in-process path).
- Follow-up empty (no `--here` emit; agent skips home-gated cd).
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
	assertAgentRunInvoked(t, req, wt, req.TaskDesc)
	assertFollowupEmptyUX(t, req)
}
```
