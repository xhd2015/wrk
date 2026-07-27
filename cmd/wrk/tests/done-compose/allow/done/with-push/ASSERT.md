## Expected

- Flag layer accepts `--done` + `--push` without requiring `--tag-next`.
- Stderr must **not** contain `--push is only valid with --tag-next`.
- Stderr must **not** contain `mutually exclusive` for this combo.
- Later-stage errors on main (e.g. `not a linked worktree`) are acceptable for P1.

## Side Effects

- None required for P1.

## Exit Code

- Any, as long as failure is not the push-only-with-tag-next rule.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "--push is only valid with --tag-next") {
		t.Fatalf("--done --push still rejected as tag-next-only; stderr=%q stdout=%q exit=%d",
			se, resp.Stdout, resp.ExitCode)
	}
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer rejects --done --push as mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
