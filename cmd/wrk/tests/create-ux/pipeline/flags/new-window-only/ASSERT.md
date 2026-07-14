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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantCreateUXWorktree(req)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertSpaceInvokedOnce(t, req)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertAgentRunNotInvoked(t, req)
}
```
