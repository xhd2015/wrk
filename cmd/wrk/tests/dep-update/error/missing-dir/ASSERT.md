## Expected

- Non-zero exit.
- Stderr indicates missing dir.
- Consumer go.mod unchanged.
- No `dep-update ` success line.

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
	assertGoModUnchanged(t, req)
	assertNotContains(t, resp.Stdout, "dep-update ")
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "no such") &&
		!strings.Contains(se, "not found") &&
		!strings.Contains(se, "missing") &&
		!strings.Contains(se, "exist") &&
		!strings.Contains(se, "stat") {
		t.Fatalf("stderr should indicate missing dir, got %q", resp.Stderr)
	}
}
```
