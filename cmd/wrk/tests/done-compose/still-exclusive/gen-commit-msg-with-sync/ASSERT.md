## Expected

- Flag layer accepts `--gen-commit-msg` + `--sync` without a done/merge-back stage.
- Stderr must **not** contain `mutually exclusive`.
- Later-stage errors (e.g. missing `--commit`, nothing to commit) are acceptable.

## Side Effects

- None required (flag-layer unlock).

## Exit Code

- Any, as long as failure is not multi-stage mutual exclusion of these two flags.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("bare --gen-commit-msg --sync still mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
