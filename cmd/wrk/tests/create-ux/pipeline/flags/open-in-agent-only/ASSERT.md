---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; path printed (worktree includes task slug).
- Outer agent-run invoked with default argv/prompt **and** `--dir` targeting the worktree.
- Process cwd of agent-run is not required to equal worktree.
- No space; no iterm.

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
	assertSpaceNotInvoked(t, req)
	assertItermNotInvoked(t, req)
	args := assertAgentRunInvoked(t, req, wt, req.TaskDesc)
	// Explicit: --dir is present and points at worktree (also checked in helper).
	assertAgentArgvHasDir(t, args, wt)
}
```
