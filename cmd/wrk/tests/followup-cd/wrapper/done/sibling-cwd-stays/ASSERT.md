## Expected

- Exit code 0.
- Stdout contains `worktree removed:`.
- Operated worktree B is gone; sibling A still exists.
- Stderr has no `cd …` follow-up line.
- FinalPWD remains the sibling shell cwd (A).

## Exit Code

- 0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertContains(t, resp.Stdout, "worktree removed:")
	assertFileNotExists(t, req.WtDir)
	assertFileExists(t, req.StartDir)
	if strings.Contains(resp.Stderr, "cd ") {
		t.Fatalf("expected no follow-up cd on stderr when sibling cwd survives; stderr=%q", resp.Stderr)
	}
	assertPathsEqual(t, resp.FinalPWD, req.StartDir)
}
```
