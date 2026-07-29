## Expected

- Non-zero exit code.
- Stderr mentions dependency (same class as `dep/not-a-dependency`).
- Preferred external path does **not** exist (no worktree created).

## Exit Code

- Non-zero

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertContains(t, resp.Stderr, "dependency")

	// Strict analyse-first: no external worktree under consumer.
	wantPath := externalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	if _, err := os.Stat(wantPath); err == nil {
		t.Fatalf("external worktree must not exist after strict not-a-dependency: %s", wantPath)
	}
	// Prefer no external/ directory at all, but tolerate empty dir only if path missing.
	extRoot := filepath.Join(req.ConsumerTop, "external")
	if entries, err := os.ReadDir(extRoot); err == nil && len(entries) > 0 {
		t.Fatalf("consumer external/ should have no worktrees after error, got %v", entries)
	}
}
```
