---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit code 0.
- Stdout contains `worktree removed:`.
- Stderr contains `cd <main-repo-abs>`.
- FinalPWD equals main repo.
- Worktree directory gone.

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
	mainAbs := resolvePath(t, req.MainRepo)
	assertContains(t, resp.Stderr, "cd "+mainAbs)
	assertPathsEqual(t, resp.FinalPWD, req.MainRepo)
	assertFileNotExists(t, req.WtDir)
}
```
