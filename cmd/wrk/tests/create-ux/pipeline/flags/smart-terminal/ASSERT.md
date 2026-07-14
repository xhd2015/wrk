## Expected

- Exit 0.
- iTerm smart script (matchingWindow + create tab).
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
	assertItermModeSmart(t, script)
	assertAgentRunNotInvoked(t, req)
}
```
