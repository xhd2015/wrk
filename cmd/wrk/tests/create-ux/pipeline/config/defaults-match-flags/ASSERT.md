## Expected

- Exit 0.
- Same side effects as `--new-window --new-terminal --open-in-agent`.

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertSpaceInvokedOnce(t, req)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertItermFollowUpHasAgentRun(t, script, wt, req.TaskDesc)
	assertAgentRunNotInvoked(t, req)
}
```
