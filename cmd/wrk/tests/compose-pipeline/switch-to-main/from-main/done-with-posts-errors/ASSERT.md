## Expected

- Non-zero exit.
- Stderr requires linked worktree and names `--done`.
- Must not look like a successful multi-stage on main (no apply success-only path that skips the gate).

## Errors

- Compose with `--done` does not bypass the linked-worktree gate when cwd is main.

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
		t.Fatalf("expected non-zero for --done+posts on main; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "linked worktree") && !strings.Contains(se, "not a linked worktree") {
		t.Fatalf("stderr should require linked worktree, got %q", se)
	}
	if !strings.Contains(se, "--done") && !strings.Contains(se, "done") {
		t.Fatalf("stderr should name --done, got %q", se)
	}
}
```
