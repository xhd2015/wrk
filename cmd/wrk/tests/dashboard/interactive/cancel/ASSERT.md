---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit **0**.
- Linked worktree still on disk (no `--done` remove).
- Last event `command: "dashboard"`, `exit_code: 0`.
- `WRK_DASHBOARD_COMPOSE_ARGV_LOG` remains empty (cancel must not build/run compose).
- No new paths under `{WRK_HOME}/worktrees/` from create.

## Side Effects

- None (no compose apply, no create).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("cancel exit: want 0, got %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertLinkedWorktreeStillPresent(t, req)
	assertNoWorktreesCreated(t, req)

	if raw := strings.TrimSpace(readComposeArgvLog(t, req)); raw != "" {
		t.Fatalf("cancel must not write compose argv log; got %q", raw)
	}

	events := readDashboardEvents(t, req.WrkHome)
	if len(events) == 0 {
		t.Fatal("expected events.jsonl for cancel dashboard")
	}
	ev := events[len(events)-1]
	if ev.Command != "dashboard" {
		t.Fatalf("event command: want %q, got %q", "dashboard", ev.Command)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
}
```
