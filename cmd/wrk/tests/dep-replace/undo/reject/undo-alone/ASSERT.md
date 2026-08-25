## Expected

- Non-zero exit.
- Stderr indicates `--undo` only valid with `--dep-replace`.

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
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "undo") {
		t.Fatalf("stderr should mention --undo, got %q", resp.Stderr)
	}
	if !strings.Contains(se, "dep-replace") {
		t.Fatalf("stderr should mention --dep-replace, got %q", resp.Stderr)
	}
}
```
