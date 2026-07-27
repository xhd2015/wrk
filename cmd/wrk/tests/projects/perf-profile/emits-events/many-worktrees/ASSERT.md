---
label: slow
---
## Expected

- Exit code 0; stdout unchanged (one project block, `Worktrees:    12 total, 0 dirty`).
- Perf log exists and contains lifecycle + latency events for all 12 worktree checks.

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
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if !strings.Contains(resp.Stdout, "Worktrees:    12 total, 0 dirty") {
		t.Fatalf("stdout missing worktree summary:\n%s", resp.Stdout)
	}

	events := readProjectsPerfLog(t, req.ProjectsPerfLog)
	if len(events) == 0 {
		t.Fatal("perf log is empty")
	}
	if events[0].Event != "run_start" {
		t.Fatalf("first event: want run_start, got %q", events[0].Event)
	}
	if events[len(events)-1].Event != "run_end" {
		t.Fatalf("last event: want run_end, got %q", events[len(events)-1].Event)
	}
	if perfRunEndMS(events) <= 0 {
		t.Fatal("run_end missing total_ms")
	}
	if perfWorktreeStatusCount(events) != 12 {
		t.Fatalf("worktree_status events: want 12, got %d", perfWorktreeStatusCount(events))
	}
	wtTotalMS, wtCount := perfPhaseTotalMS(events, "worktree_status_all")
	if wtCount != 12 {
		t.Fatalf("worktree_status_all count: want 12, got %d", wtCount)
	}
	if wtTotalMS <= 0 {
		t.Fatal("worktree_status_all missing duration_ms")
	}
	t.Logf("perf: run_end=%dms worktree_status_all=%dms", perfRunEndMS(events), wtTotalMS)
}
```