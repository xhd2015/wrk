## Expected

- Non-zero exit (push cannot resolve a remote).
- Stderr mentions remote resolution failure (e.g. `remote`, `origin`, or `upstream`).
- Primary merge-back is allowed to have already succeeded (no undo of done): worktree may be gone and main may have `feature-work`.
- Stdout must **not** contain the success push line `pushed main → origin/main`.

## Errors

- Clear non-zero error when no remote can be resolved for the main branch push.

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
		t.Fatalf("expected non-zero exit when --push has no remote; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "remote") && !strings.Contains(se, "origin") && !strings.Contains(se, "upstream") {
		t.Fatalf("stderr should explain missing remote/origin/upstream, got %q", resp.Stderr)
	}
	if strings.Contains(resp.Stdout, "pushed main → origin/main") {
		t.Fatalf("stdout must not claim successful push without a remote; got %q", resp.Stdout)
	}
}
```
