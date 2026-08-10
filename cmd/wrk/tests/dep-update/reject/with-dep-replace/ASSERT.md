## Expected

- Non-zero exit.
- Stderr indicates mutual exclusion.
- Prefer naming dep-update and/or dep-replace.

## Errors

- Cannot combine `--dep-update` and `--dep-replace`.

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
	if !strings.Contains(se, "dep-update") && !strings.Contains(se, "dep-replace") {
		t.Fatalf("stderr should name dep-update and/or dep-replace, got %q", se)
	}
}
```
