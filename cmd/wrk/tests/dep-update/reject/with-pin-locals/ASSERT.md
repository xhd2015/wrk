## Expected

- Non-zero exit.
- Stderr indicates mutual exclusion.
- Prefer naming dep-update and/or pin-locals.

## Errors

- `--dep-update` cannot be combined with `--pin-locals`.

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
	assertMutualExclusion(t, resp)
	se := resp.Stderr
	if !strings.Contains(se, "dep-update") &&
		!strings.Contains(se, "pin-locals") &&
		!strings.Contains(se, "--pin-locals") {
		t.Fatalf("stderr should name dep-update and/or pin-locals, got %q", se)
	}
}
```
