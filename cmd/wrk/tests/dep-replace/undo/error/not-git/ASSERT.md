## Expected

- Non-zero exit.
- Stderr mentions undo and git/HEAD.
- No undo banner.

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
		t.Fatalf("stderr should mention undo, got %q", resp.Stderr)
	}
	if !strings.Contains(se, "git") && !strings.Contains(se, "head") {
		t.Fatalf("stderr should mention git HEAD, got %q", resp.Stderr)
	}
	assertNotContains(t, resp.Stdout, "==== dep-replace --undo")
}
```
