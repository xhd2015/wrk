
## Expected

- Exit code 0.
- Worktree still exists.
- Follow-up file empty (merge-back is out of scope for auto-cd).

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
	assertFileExists(t, req.WtDir)
	assertFollowupEmpty(t, resp)
}
```
