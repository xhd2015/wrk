## Expected

- Exit 0; stdout is worktree path + `\n`.
- Worktree exists under `{WRK_HOME}/worktrees/`.
- Space, iterm, and agent-run mocks are not invoked.

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
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
