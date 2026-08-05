
## Expected

- Exit code 0.
- Stdout contains `worktree removed:`.
- Worktree directory gone.
- Follow-up file is empty (foreign-repo ancestor gate: parent is consumer main ≠ dest).

## Exit Code

- 0

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertContains(t, resp.Stdout, "worktree removed:")
	assertFileNotExists(t, req.WtDir)
	assertFollowupEmpty(t, resp)
}
```
