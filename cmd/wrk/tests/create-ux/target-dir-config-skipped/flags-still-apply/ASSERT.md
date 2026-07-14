## Expected

- Exit 0.
- Stdout is absolute worktree path at `SpawnDir` + `\n` (exact path; task affects
  branch name only, not the spawn path when parent exists / target missing).
- Space invoked once (`CreateAndActivate`).
- iTerm ForceNew at worktree; follow-up contains agent-run + `--dir` + worktree + runner + task.
- Outer agent-run log empty (agent only as iTerm follow-up).

## Side Effects

- Worktree at `{WorkRoot}/wt`.
- Window + terminal + agent follow-up driven solely by CLI flags.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := req.SpawnDir
	if wt == "" {
		wt = filepath.Join(req.WorkRoot, "wt")
	}
	assertNativeCreateOK(t, req, resp, err, wt)
	assertFileNotExists(t, wantCreateUXWorktreeWithTask(req, req.TaskDesc))
	assertSpaceInvokedOnce(t, req)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertItermFollowUpHasAgentRun(t, script, wt, req.TaskDesc)
	assertAgentRunNotInvoked(t, req)
}
```
