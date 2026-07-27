## Expected

- Flag layer accepts primary + reinstall-local + `--exec` together.
- Stderr must **not** contain `mutually exclusive`.
- Stderr must **not** contain `--exec is not valid with --reinstall-local`.
- Later-stage errors on main-repo cwd are acceptable for P1.

## Side Effects

- None required (flag-layer only; no real exec success required).

## Exit Code

- Any, as long as failure is not flag rejection of this combo.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer rejects --done --reinstall-local --exec; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
	if strings.Contains(se, "--exec is not valid with --reinstall-local") {
		t.Fatalf("--exec still forbidden with reinstall under primary; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
