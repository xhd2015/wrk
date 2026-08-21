## Expected

- Non-zero exit.
- Stderr indicates missing second path (`wrk:` + no such dir or equivalent).
- No banner / no success summary (validate every dir before apply).
- Consumer go.mod unchanged (no partial first replace).

## Side Effects

- None: bad second arg aborts before any replace/tidy.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertNoBanner(t, resp.Stdout)
	assertGoModUnchanged(t, req)
	assertNotContains(t, resp.Stdout, "dep-replace: replaced")
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "no such") &&
		!strings.Contains(se, "not found") &&
		!strings.Contains(se, "missing") &&
		!strings.Contains(se, "exist") &&
		!strings.Contains(se, "stat") {
		t.Fatalf("stderr should indicate second path missing, got %q", resp.Stderr)
	}
}
```
