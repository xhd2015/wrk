---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0.
- Stdout is absolute worktree path under `{WRK_HOME}/worktrees/` + trailing `\n`.
- Worktree exists with linked `.git` file.
- Space / iterm / outer agent-run not invoked.
- Stderr must not mention JSON/parse/config load failure (config must not be opened).
- Preferred: stderr empty.

## Side Effects

- Corrupt file left untouched on disk; never read under `--no-config`.

## Errors

- No config parse error when `--no-config` is set.

## Exit Code

- 0

```go
import (
	"strings"
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
	se := strings.ToLower(resp.Stderr)
	for _, bad := range []string{"json", "parse", "unmarshal", "invalid character", "config.json"} {
		if strings.Contains(se, bad) {
			t.Fatalf("--no-config must not open/read corrupt config; stderr mentions %q: %q", bad, resp.Stderr)
		}
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty under --no-config bare create, got %q", resp.Stderr)
	}
}
```
