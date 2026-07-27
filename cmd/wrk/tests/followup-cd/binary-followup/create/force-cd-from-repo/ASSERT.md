
## Expected

- Exit code 0.
- Stdout is the new worktree absolute path (trailing `\n`).
- Worktree directory exists.
- Follow-up file is exactly `cd <abs-worktree>\n` (home gate bypassed by `--force-cd`).
- Stderr empty (follow-ups are file-only when channel open; no install-hint shell path).

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
	assertFollowupCD(t, resp, wantPath)
	if resp.Stderr != "" {
		t.Fatalf("binary stderr should be empty on Branch A force-cd (file-only), got %q", resp.Stderr)
	}
	assertFileExists(t, wantPath)
}
```
