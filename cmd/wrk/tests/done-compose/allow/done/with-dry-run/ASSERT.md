## Expected

- Flag layer accepts `--done` + `--dry-run`.
- Stderr must **not** contain `--dry-run is only valid with`.
- Stderr must **not** contain `mutually exclusive`.
- Full dry-run plan body is **not** asserted here (see `done-pipeline/dry-run/`); flag-layer acceptance only.

## Side Effects

- None required for flag-matrix leaf.

## Exit Code

- Any, as long as failure is not dry-run host-mode rejection.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "--dry-run is only valid with") {
		t.Fatalf("--done --dry-run still rejected as invalid dry-run host; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("--done --dry-run rejected as mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
