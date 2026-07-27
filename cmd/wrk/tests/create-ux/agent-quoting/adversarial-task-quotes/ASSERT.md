---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- Outer agent-run invoked with `--dir` targeting worktree.
- Last argv element is exactly `/brainstorm fix "quoted" task` (quotes preserved inside one argv token).

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
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
