---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Live tty-watch snapshot shows dashboard UI **before** process exit.
- After `q`, session ends (`[Terminal exited]`).
- No worktrees created under WRK_HOME; compose argv log empty.
- Linked worktree still present.

## Exit Code

- Process exits cleanly (tty-watch may not expose exit code; assert exited + no compose).

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	// Non-TTY harness Run is fine (static snapshot); drive real TTY via tty-watch.
	alive, final := runDashboardTTYWatch(t, req, "cancel", nil, "q")
	if !snapshotLooksLikeDashboard(alive) {
		t.Fatalf("expected live dashboard snapshot before cancel; got=%q", alive)
	}
	if strings.Contains(alive, "[Terminal exited]") {
		t.Fatalf("dashboard must still be alive before send q; snap=%q", alive)
	}
	if !strings.Contains(final, "[Terminal exited]") {
		t.Fatalf("expected terminal exit after q; final=%q", final)
	}
	assertNoWorktreesCreated(t, req)
	assertLinkedWorktreeStillPresent(t, req)
	raw := readComposeArgvLog(t, req)
	if strings.TrimSpace(raw) != "" {
		t.Fatalf("cancel must not write compose argv log; got=%q", raw)
	}
}
```
