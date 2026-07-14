## Expected

- Exit 0.
- Outer agent-run invoked with `--dir` targeting worktree.
- Last argv element is exactly `/brainstorm fix "quoted" task` (quotes preserved inside one argv token).

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	args := assertAgentRunInvoked(t, req, wt, req.TaskDesc)
	assertAgentArgvHasDir(t, args, wt)
	last := args[len(args)-1]
	want := `/brainstorm fix "quoted" task`
	if last != want {
		t.Fatalf("prompt argv: want %q, got %q (full=%v)", want, last, args)
	}
}
```
