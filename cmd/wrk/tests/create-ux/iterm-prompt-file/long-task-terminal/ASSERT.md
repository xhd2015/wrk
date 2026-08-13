---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- iTerm ForceNew at worktree; follow-up has `agent-run`, `--dir`, runner, `--color`.
- Follow-up uses `--prompt-file=<abs>`; file body is `/brainstorm` + full TaskDesc.
- AppleScript does not embed the 700-x body.
- Follow-up write-text payload is under write-text SafeMax.
- Outer agent-run is not exec'd.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertSpaceNotInvoked(t, req)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertItermFollowUpHasAgentRun(t, script, wt, "")
	wantPrompt := "/brainstorm " + req.TaskDesc
	assertItermFollowUpUsesPromptFile(t, script, wantPrompt)
	if strings.Contains(script, strings.Repeat("x", 700)) {
		t.Fatalf("script must not embed 700-x task body:\n%s", script)
	}
	// The write-text payload (follow-up line) must be write-text safe.
	path, ok := itermFollowUpPromptFilePath(script)
	if !ok {
		t.Fatal("missing --prompt-file after helper asserted it")
	}
	fu := "agent-run run --dir " + wt + " --session-id-from-prompt --no-submit --open --color --agent-runner=grok-tty --prompt-file=" + path
	if !applescript.CheckWriteText(fu).OK {
		t.Fatalf("follow-up still over write-text limit: len=%d", len(fu))
	}
	assertAgentRunNotInvoked(t, req)
}
```
