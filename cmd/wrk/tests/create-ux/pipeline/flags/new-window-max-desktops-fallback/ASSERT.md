---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; path printed (native create succeeds).
- Space invoked once (`CreateAndActivate` logged).
- Stderr contains `warning:` and mentions maximum Desktops / current Desktop.
- Stderr does **not** hard-fail with only a fatal window error (exit remains 0).
- iTerm ForceNew still targets worktree path.
- Outer agent-run not invoked.

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
	assertSpaceInvokedOnce(t, req)
	s := strings.ToLower(resp.Stderr)
	if !strings.Contains(s, "warning:") {
		t.Fatalf("expected warning: on stderr, got %q", resp.Stderr)
	}
	if !strings.Contains(s, "maximum") && !strings.Contains(s, "desktop") {
		t.Fatalf("expected max-Desktop capacity warning, got %q", resp.Stderr)
	}
	if strings.Contains(s, "wrk: window:") && resp.ExitCode != 0 {
		t.Fatalf("expected soft-fail, got hard window error: %q", resp.Stderr)
	}
	script := assertItermInvokedAtPath(t, req, wt)
	assertItermModeForceNew(t, script)
	assertAgentRunNotInvoked(t, req)
}
```
