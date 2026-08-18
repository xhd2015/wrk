## Expected

- Non-zero exit.
- Stderr indicates missing consumer go.mod / cannot find module (walk-up failed).
- No success `dep-replace ` line.

## Errors

- No nearest consumer go.mod from workDir (D6).

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertNotContains(t, resp.Stdout, "dep-replace ")
	assertNoBanner(t, resp.Stdout)
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "go.mod") &&
		!strings.Contains(se, "module") &&
		!strings.Contains(se, "consumer") &&
		!strings.Contains(se, "find") {
		t.Fatalf("stderr should indicate no consumer go.mod, got %q", resp.Stderr)
	}
}
```
