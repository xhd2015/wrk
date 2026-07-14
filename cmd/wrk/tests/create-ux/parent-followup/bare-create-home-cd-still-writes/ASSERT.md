## Expected

- Exit 0; stdout is worktree path + trailing `\n`.
- No space / iterm / agent-run.
- Follow-up file contains exactly `cd <worktree-abs>` with trailing newline (home gate still open for bare create).

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
	assertItermNotInvoked(t, req)
	assertAgentRunNotInvoked(t, req)
	assertFollowupCDUX(t, req, wt)
}
```
