## Expected

- Flag layer accepts `--reinstall-local` + `--sync` without done/merge-back.
- Stderr must **not** contain `mutually exclusive`.

## Side Effects

- None required (flag-layer unlock).

## Exit Code

- Any, as long as failure is not multi-stage mutual exclusion of these two flags.

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("bare --reinstall-local --sync still mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
