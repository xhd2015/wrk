## Expected

- Flag layer accepts `--done` + `--reinstall-local` (no mutual-exclusion error).
- Stderr must **not** contain `mutually exclusive`.
- Later-stage errors on main-repo cwd (e.g. `not a linked worktree`) are acceptable for P1.

## Side Effects

- None required (flag-layer only).

## Exit Code

- Any, as long as failure is not flag mutual exclusion of this combo.

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer still rejects --done --reinstall-local as mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
