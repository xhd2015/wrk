## Expected Output

One absolute path for the linked worktree; closed state must not block lookup.

## Expected

- Exit code 0.
- Stdout is the linked worktree path plus trailing `\n`.
- Stderr is empty (no “closed” / “not open” refuse).

## Side Effects

- Read-only location lookup.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("closed PR should still resolve; exit %d stderr=%q stdout=%q",
			resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutExactPath(t, resp.Stdout, resolvePath(t, req.WtDir))
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if strings.Contains(se, "closed") || strings.Contains(se, "not open") ||
		strings.Contains(se, "merged only") {
		t.Fatalf("must not refuse closed PR for location lookup; stderr=%q", resp.Stderr)
	}
}
```
