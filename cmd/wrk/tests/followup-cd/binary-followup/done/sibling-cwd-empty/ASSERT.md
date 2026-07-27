
## Expected

- Exit code 0.
- Stdout contains `worktree removed:`.
- Operated worktree B is gone.
- Sibling shell cwd A still exists.
- Follow-up file is empty (cwd gate: shell cwd still present).

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
	assertFileExists(t, req.RepoDir)
	assertFollowupEmpty(t, resp)
}
```
