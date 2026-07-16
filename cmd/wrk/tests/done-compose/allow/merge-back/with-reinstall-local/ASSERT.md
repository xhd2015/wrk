## Expected

- Flag layer accepts `--merge-back` + `--reinstall-local` (no mutual-exclusion error).
- Stderr must **not** contain `mutually exclusive`.
- Later-stage errors on main-repo cwd are acceptable for P1.

## Side Effects

- None required for P1.

## Exit Code

- Any, as long as failure is not flag mutual exclusion.

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer still rejects --merge-back --reinstall-local as mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
