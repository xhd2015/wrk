---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Live dashboard before keys complete.
- Select DONE (from default MERGE BACK focus), RUN ALL + Enter runs DONE dry-run with `--add-all` (argv log).
- TUI re-opens after RUN (stay-in-TUI); helper then sends `q` so the process exits.
- Worktree still present (dry-run).

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	logPath := dashComposeArgvLogPath(req)
	writeFile(t, logPath, "")
	extra := []string{
		envDashboardDryRun + "=1",
		envDashboardComposeArgvLog + "=" + logPath,
	}
	// Default cursor = MERGE BACK (index 3).
	// j → DONE; space selects DONE; j×5 → RUN ALL; \r runs batch (TUI re-opens).
	keys := []string{"j", " ", "j", "j", "j", "j", "j", "\r"}
	alive, final := runDashboardTTYWatchRunThenQuit(t, req, "run-done", extra, keys...)
	if !snapshotLooksLikeDashboard(alive) {
		t.Fatalf("expected live dashboard; got=%q", alive)
	}
	if !strings.Contains(final, "[Terminal exited]") {
		t.Fatalf("expected exit after RUN then q; final=%q", final)
	}
	assertComposeArgvRecipeDone(t, req, true /* dry-run */, true /* add-all */)
	assertLinkedWorktreeStillPresent(t, req)
}
```
