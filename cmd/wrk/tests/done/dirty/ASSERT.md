## Expected

- Non-zero exit code.
- Stderr mentions clean / uncommitted changes.
- Worktree directory still exists.

## Exit Code

- Non-zero

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	combined := strings.ToLower(resp.Stderr + resp.Stdout)
	if !strings.Contains(combined, "uncommitted") && !strings.Contains(combined, "clean") {
		t.Fatalf("expected dirty worktree error, got stderr=%q stdout=%q", resp.Stderr, resp.Stdout)
	}

	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}
```