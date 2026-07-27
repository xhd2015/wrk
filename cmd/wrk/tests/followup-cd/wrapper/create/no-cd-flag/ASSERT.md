---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit code 0.
- Create succeeds.
- No stderr `cd ` follow-up line.
- FinalPWD remains FakeHome (start dir / user home).

## Exit Code

- 0

```go
import (
	"regexp"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assert.Output(t, resp.Stdout, "---\nversion: 3\n---\n"+regexp.QuoteMeta(wantPath)+"\n")
	for _, line := range strings.Split(resp.Stderr, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "cd ") {
			t.Fatalf("--no-cd must not print follow-up cd; got stderr line %q", line)
		}
	}
	assertPathsEqual(t, resp.FinalPWD, req.StartDir)
	assertPathsEqual(t, resp.FinalPWD, req.FakeHome)
}
```
