
## Expected

- Exit code 0.
- Stdout contains `task unchanged`.
- Worktree path unchanged.
- Follow-up file empty.

## Exit Code

- 0

```go
import (
	"strings"
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
	if !strings.Contains(resp.Stdout, "task unchanged") {
		t.Fatalf("expected 'task unchanged' in stdout, got %q", resp.Stdout)
	}
	assertFileExists(t, req.WtDir)
	assertFollowupEmpty(t, resp)
}
```
