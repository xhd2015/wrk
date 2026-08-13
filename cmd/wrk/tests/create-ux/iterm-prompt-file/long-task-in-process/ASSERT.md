---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- Outer agent-run invoked with `--dir` = worktree and `--prompt-file=<abs>`.
- Spill file body is `/brainstorm` + full original TaskDesc.
- Positional prompt token is not the long body.
- No iTerm / no space.

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
	wantPrompt := "/brainstorm " + req.TaskDesc
	assertAgentRunInvokedWithPromptFile(t, req, wt, wantPrompt)
}
```
