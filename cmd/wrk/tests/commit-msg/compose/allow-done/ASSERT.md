## Expected

- Flag layer accepts `--commit -m … --done` together (flags are recognized).
- Stderr must **not** contain `mutually exclusive` for this combo.
- Stderr must **not** be `unrecognized flag` for `--commit`, `-m`, or `--message`.
- Later-stage errors on main-repo cwd (e.g. `not a linked worktree`) are acceptable.

## Side Effects

- None required (flag-layer only; no successful merge-back).

## Exit Code

- Any, as long as failure is not flag mutual exclusion or unrecognized message/commit flags.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "unrecognized flag") {
		t.Fatalf("flag layer does not recognize --commit/-m yet (unrecognized flag); stderr=%q exit=%d",
			se, resp.ExitCode)
	}
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer still rejects --commit -m --done as mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
