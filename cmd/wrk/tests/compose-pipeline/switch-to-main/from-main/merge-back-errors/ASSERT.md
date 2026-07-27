## Expected

- Non-zero exit.
- Stderr mentions linked worktree requirement.
- Stderr names `--merge-back` (or `merge-back`).

## Errors

- `--merge-back` is illegal on main checkout.

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
		t.Fatalf("expected non-zero for --merge-back on main; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "linked worktree") && !strings.Contains(se, "not a linked worktree") {
		t.Fatalf("stderr should require linked worktree, got %q", se)
	}
	if !strings.Contains(se, "--merge-back") && !strings.Contains(se, "merge-back") {
		t.Fatalf("stderr should name --merge-back, got %q", se)
	}
}
```
