## Expected

- Flag layer accepts `--gen-commit-msg --commit --model=m --done` together.
- Stderr must **not** contain `mutually exclusive`.
- Later-stage errors on main-repo cwd (e.g. `not a linked worktree`) are acceptable for P2 flag-layer.

## Side Effects

- None required (flag-layer only; no successful merge-back).

## Exit Code

- Any, as long as failure is not flag mutual exclusion of this combo.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer still rejects --gen-commit-msg --commit --done as mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
