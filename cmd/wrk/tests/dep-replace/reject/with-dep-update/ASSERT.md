## Expected

- Non-zero exit.
- Stderr indicates mutual exclusion (or equivalent).
- Prefer naming `--dep-replace` and/or `--dep-update`.
- No successful apply lines.

## Errors

- Cannot combine `--dep-replace` and `--dep-update`.

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
	if !strings.Contains(se, "dep-replace") && !strings.Contains(se, "dep-update") {
		t.Fatalf("stderr should name dep-replace and/or dep-update, got %q", se)
	}
	assertNotContains(t, resp.Stdout, "dep-replace ")
	assertNotContains(t, resp.Stdout, "dep-update ")
}
```
