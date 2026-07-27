---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit **0**.
- Last `events.jsonl` event has `command: "dashboard"`, `exit_code: 0`.
- Event `work_dir` / `main_repo` resolve to the main checkout.
- No create worktree under `{WRK_HOME}/worktrees/`.
- Stdout is a real dashboard snapshot (identity **dashboard**; not a create stub).

## Side Effects

- One events.jsonl line for this invocation; no create.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
	"time"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNoWorktreesCreated(t, req)

	events := readDashboardEvents(t, req.WrkHome)
	if len(events) == 0 {
		t.Fatal("expected events.jsonl entry for bare wrk dashboard")
	}
	ev := events[len(events)-1]
	if ev.Command != "dashboard" {
		t.Fatalf("event command: want %q, got %q (all=%v)", "dashboard", ev.Command, events)
	}
	if ev.ExitCode != 0 {
		t.Fatalf("event exit_code: want 0, got %d", ev.ExitCode)
	}
	if resolvePath(t, ev.MainRepo) != resolvePath(t, req.MainRepo) {
		t.Fatalf("event main_repo: want %q, got %q", req.MainRepo, ev.MainRepo)
	}
	if resolvePath(t, ev.WorkDir) != resolvePath(t, req.MainRepo) {
		t.Fatalf("event work_dir: want %q, got %q", req.MainRepo, ev.WorkDir)
	}
	if ev.TS == "" {
		t.Fatal("event missing ts")
	}
	if _, err := time.Parse(time.RFC3339, ev.TS); err != nil {
		t.Fatalf("event ts not RFC3339: %q", ev.TS)
	}
	if !strings.Contains(strings.ToLower(resp.Stdout), "dashboard") {
		t.Fatalf("stdout should mention dashboard, got %q", resp.Stdout)
	}
	// Soft identity of MVP snapshot (not create path)
	if strings.Contains(strings.ToLower(resp.Stdout), "create a worktree") {
		t.Fatalf("dashboard event leaf must not be create-hint stub; stdout=%q", resp.Stdout)
	}
}
```
