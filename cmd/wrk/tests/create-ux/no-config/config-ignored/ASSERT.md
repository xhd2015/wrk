---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- Stdout is absolute worktree path under `{WRK_HOME}/worktrees/` + trailing `\n`.
- Worktree exists with linked `.git` file.
- Space log empty; iterm script empty; outer agent-run not invoked.
- Config UX is skipped silently (not an error) — stderr empty preferred.
- RED without implementation: if config still applied, space/iterm/agent would fire
  (same full config as `pipeline/config/defaults-match-flags`).

## Side Effects

- Native create only; no Mission Control space, iTerm, or agent-run from config create.*.

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	wt := wantCreateUXWorktree(req)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertSpaceNotInvoked(t, req)
	assertItermNotInvoked(t, req)
	assertAgentRunNotInvoked(t, req)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty (silent config skip via --no-config), got %q", resp.Stderr)
	}
}
```
