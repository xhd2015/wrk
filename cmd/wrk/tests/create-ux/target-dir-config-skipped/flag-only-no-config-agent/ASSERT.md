## Expected

- Exit 0.
- Worktree at exact `SpawnDir` path; stdout path + `\n`.
- iTerm ForceNew at worktree path.
- iTerm script must **not** contain agent-run follow-up (config agent not base-merged).
- Outer agent-run not invoked.
- Space not invoked (no window CLI flag; config window not applied).

## Side Effects

- Distinguishes config-skip from flags-apply: terminal yes from flag; agent no despite config.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	wt := req.SpawnDir
	if wt == "" {
		wt = filepath.Join(req.WorkRoot, "wt")
	}
	assertNativeCreateOK(t, req, resp, err, wt)
	assertFileNotExists(t, wantCreateUXWorktree(req))
	assertSpaceNotInvoked(t, req)
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	if strings.Contains(script, "agent-run") {
		t.Fatalf("config agent must not apply with <target-dir>; iterm follow-up must lack agent-run; script:\n%s", script)
	}
	assertAgentRunNotInvoked(t, req)
}
```
