## Expected

- Non-zero exit.
- Stderr is an Error/`wrk:` style message.
- Mentions linked worktree requirement.
- Mentions `--done` (or `done`) so the gated flag is named.

## Errors

- `--done` is illegal on main checkout.

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
		t.Fatalf("expected non-zero for --done on main; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "linked worktree") && !strings.Contains(se, "not a linked worktree") {
		t.Fatalf("stderr should require linked worktree, got %q", se)
	}
	if !strings.Contains(se, "--done") && !strings.Contains(se, "done") {
		t.Fatalf("stderr should name --done, got %q", se)
	}
	if !strings.Contains(se, "Error") && !strings.Contains(se, "wrk:") && !strings.Contains(se, "wrk ") {
		// soft: many paths already use "wrk:" / "not a linked worktree"
		if !strings.Contains(se, "not a linked worktree") {
			t.Fatalf("stderr should look like a wrk error, got %q", se)
		}
	}
}
```
