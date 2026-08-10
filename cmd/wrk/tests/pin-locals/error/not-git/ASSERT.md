## Expected

- Non-zero exit.
- Stderr mentions `not a git repository`.
- No pin-local success lines; no go.mod writes (none present).

## Errors

- Hard preflight: cwd is not a git repository.

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
	assertContains(t, resp.Stderr, "not a git repository")
	assertNotContains(t, resp.Stdout, "pin-local ")
}
```
