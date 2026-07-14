## Expected

- Exit 0; stdout is worktree path + trailing `\n`.
- Outer agent-run invoked with `--dir` targeting worktree.
- Follow-up file contains `cd <worktree-abs>` (`--force-cd` still lands parent).

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertAgentRunInvoked(t, req, wt, req.TaskDesc)
	assertFollowupCDUX(t, req, wt)
}
```
