## Expected

- Exit 0.
- Stdout is absolute worktree path (task slug naming under `{WRK_HOME}/worktrees/`) + `\n`.
- Space invoked once (`CreateAndActivate`).
- iTerm ForceNew at worktree; follow-up contains agent-run + `--dir` + worktree + runner + task.
- Outer agent-run log empty (agent only as iTerm follow-up).

## Side Effects

- Window + terminal + agent follow-up driven by CLI flags while config load is skipped.
- Worktree under default WRK_HOME layout (not a spawn target).

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
