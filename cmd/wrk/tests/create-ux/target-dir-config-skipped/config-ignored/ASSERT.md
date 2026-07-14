## Expected

- Exit 0.
- Stdout is absolute worktree path at `SpawnDir` + trailing `\n` (exact spawn;
  not under `{WRK_HOME}/worktrees`).
- Worktree exists at `SpawnDir` with linked `.git` file.
- Space log empty; iterm script empty; outer agent-run not invoked.
- Silent skip of config (not an error) — stderr empty preferred.

## Side Effects

- Worktree created exactly at `{WorkRoot}/wt`.
- No Mission Control space, iTerm, or agent-run from config create.*.

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
	// Must not land under default WRK_HOME layout.
	assertFileNotExists(t, wantCreateUXWorktree(req))
	assertSpaceNotInvoked(t, req)
	assertItermNotInvoked(t, req)
	assertAgentRunNotInvoked(t, req)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty (silent config skip), got %q", resp.Stderr)
	}
}
```
