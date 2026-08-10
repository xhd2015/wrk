## Expected

- Non-zero exit.
- Stderr indicates not a go module.
- Consumer go.mod unchanged.

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
	if !strings.Contains(se, "not a go module") &&
		!strings.Contains(se, "go.mod") &&
		!strings.Contains(se, "module") {
		t.Fatalf("stderr should indicate not a go module, got %q", resp.Stderr)
	}
}
```
