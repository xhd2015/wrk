## Expected

- Flag layer accepts `--done --sync --reinstall-local` together.
- Stderr must **not** contain `mutually exclusive`.
- Must not reject with bare-sync / bare-reinstall exclusivity wording that ignores primary.

## Side Effects

- None required for P1 (flag-layer only).

## Exit Code

- Any, as long as failure is not flag-matrix rejection of this combo.

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("--done --sync --reinstall-local still mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
}
```
