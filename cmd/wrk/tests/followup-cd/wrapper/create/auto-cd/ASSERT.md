## Expected

- Exit code 0.
- Stdout ends with worktree path line (user-facing create output).
- Stderr contains exactly a `cd <worktree>` line (no `wrk:` prefix required on that line).
- Final shell `pwd` equals the new worktree path.

## Exit Code

- 0

```go
import (
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	// stdout from wrk binary should still print path
	assert.Output(t, resp.Stdout, "---\nversion: 2\n---\n"+wantPath+"\n")
	wantCD := "cd " + wantPath
	assertContains(t, resp.Stderr, wantCD)
	// Must not prefix follow-up with wrk:
	for _, line := range strings.Split(resp.Stderr, "\n") {
		if strings.Contains(line, wantCD) && strings.HasPrefix(strings.TrimSpace(line), "wrk:") {
			t.Fatalf("follow-up cd line must not use wrk: prefix: %q", line)
		}
	}
	assertPathsEqual(t, resp.FinalPWD, wantPath)
}
```
