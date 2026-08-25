---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; worktree path on stdout.
- Stderr contains bash-integration install warning.
- Fake bash launched with cwd = worktree; args include `--rcfile`.
- No space/iterm; outer agent-run not invoked (fake bash exits without sourcing rc).

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
	wt := wantCreateUXWorktreeWithTask(req, req.TaskDesc)
	assertNativeCreateOK(t, req, resp, err, wt)
	assertInstallHintUX(t, resp.Stderr)
	assertSpaceNotInvoked(t, req)
	assertItermNotInvoked(t, req)
	assertAgentRunNotInvoked(t, req)
	assertFakeShellLaunchedUX(t, req)
	assertFakeShellCwdUX(t, req, wt)
	log := readFileEmptyOK(t, req.FakeShellLog)
	if !strings.Contains(log, "--rcfile") {
		t.Fatalf("fake bash args should include --rcfile for --here bash startup; log=%q", log)
	}
}
```
