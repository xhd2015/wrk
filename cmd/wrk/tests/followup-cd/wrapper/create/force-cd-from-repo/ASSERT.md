## Expected

- Exit code 0.
- Create succeeds (stdout worktree path).
- Stderr contains a `cd <worktree>` follow-up line (wrapper prints then executes; no `wrk:` prefix on that line).
- FinalPWD is the new worktree (not main-repo start dir).

## Exit Code

- 0

```go
import (
	"regexp"
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
	assert.Output(t, resp.Stdout, "---\nversion: 3\n---\n"+regexp.QuoteMeta(wantPath)+"\n")
	assertFileExists(t, wantPath)
	wantCD := "cd " + wantPath
	assertContains(t, resp.Stderr, wantCD)
	for _, line := range strings.Split(resp.Stderr, "\n") {
		if strings.Contains(line, wantCD) && strings.HasPrefix(strings.TrimSpace(line), "wrk:") {
			t.Fatalf("follow-up cd line must not use wrk: prefix: %q", line)
		}
	}
	assertPathsEqual(t, resp.FinalPWD, wantPath)
}
```
