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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantCreateUXWorktree(req)
	assertNativeCreateOK(t, req, resp, err, wt)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertAgentRunNotInvoked(t, req)
	assertFollowupEmptyUX(t, req)
}
```
