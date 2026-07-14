## Expected

- Exit 0.
- iTerm script is reuse-current mode (matchingSession).
- No space; no agent.

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantCreateUXWorktree(req)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertSpaceNotInvoked(t, req)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeReuse(t, script)
	assertAgentRunNotInvoked(t, req)
}
```
