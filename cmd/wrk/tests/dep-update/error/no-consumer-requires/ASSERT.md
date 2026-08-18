## Expected

- Non-zero exit.
- Stderr is a `wrk:` error containing `requires` (e.g. no module under root requires path).
- No banner. Consumer go.mod unchanged; no successful pin line.

## Errors

- Zero matching requirers under the consumer root.

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
	assertNoBanner(t, resp.Stdout)
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "requires") {
		t.Fatalf("stderr should contain requires, got %q", resp.Stderr)
	}
}
```
