## Expected

- Non-zero exit code.
- Stderr mentions mutual exclusivity / mode conflict (`mutually exclusive`).
- No external worktree created under consumer.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
	"os"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "mutually exclusive")

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	if _, err := os.Stat(wantPath); err == nil {
		t.Fatalf("external worktree should not exist after mutual-exclusion error: %s", wantPath)
	}
}
```
