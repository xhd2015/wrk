
## Expected

- Exit code 0.
- Create succeeds (stdout path).
- Follow-up file remains empty (env was not set; binary must not write to a guessed path).

## Exit Code

- 0

```go
import (
	"regexp"
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
	assertFollowupEmpty(t, resp)
}
```
