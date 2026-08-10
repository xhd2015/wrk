## Expected

- Non-zero exit.
- Stderr indicates missing directory argument.
- No `dep-update ` success lines.

## Errors

- Empty paths → Error requires directory.

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
	if !strings.Contains(se, "director") &&
		!strings.Contains(se, "path") &&
		!strings.Contains(se, "requires") &&
		!strings.Contains(se, "argument") {
		t.Fatalf("stderr should indicate missing directory arg, got %q", resp.Stderr)
	}
	assertNotContains(t, resp.Stdout, "dep-update ")
}
```
