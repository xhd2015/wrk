---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit 0; worktree created.
- Stderr warns that bash integration is outdated for `--here`.
- Follow-up has `cd` only (no `agent-run` line).
- Outer agent-run invoked in-process with `--dir`.

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
	if !strings.Contains(resp.Stderr, "outdated") || !strings.Contains(resp.Stderr, "--bash-integration --install") {
		t.Fatalf("stderr should warn outdated bash integration; got %q", resp.Stderr)
	}
	assertFollowupCDUX(t, req, wt)
	got := readFollowupFileUX(t, req)
	if strings.Contains(got, "agent-run") {
		t.Fatalf("outdated wrapper path should not emit agent-run follow-up; got %q", got)
	}
	assertAgentRunInvoked(t, req, wt, req.TaskDesc)
	assertItermNotInvoked(t, req)
	assertSpaceNotInvoked(t, req)
}
```
