---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; stdout is worktree path + trailing `\n`.
- No space / iterm / agent-run.
- Follow-up file contains exactly `cd <worktree-abs>` with trailing newline (home gate still open for bare create).

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
	assertFollowupCDUX(t, req, wt)
}
```
