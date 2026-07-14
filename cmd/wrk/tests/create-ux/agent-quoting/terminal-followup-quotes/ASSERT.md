## Expected

- Exit 0.
- iTerm script contains agent-run follow-up with `--dir` + worktree and shell-safe encoding of the prompt (POSIX single-quoted word containing the `"` characters).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermFollowUpHasAgentRun(t, script, wt, req.TaskDesc)
	wantSafe := shellSafeQuoteUX(`/brainstorm ` + req.TaskDesc)
	if !strings.Contains(script, wantSafe) && !strings.Contains(script, req.TaskDesc) {
		t.Fatalf("follow-up should embed shell-safe or raw task prompt; wantSafe=%q script:\n%s", wantSafe, script)
	}
	assertAgentRunNotInvoked(t, req)
}
```
