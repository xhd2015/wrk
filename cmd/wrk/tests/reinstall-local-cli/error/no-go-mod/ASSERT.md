## Expected

- Non-zero exit code.
- Stderr is non-empty and clearly reports missing module / go.mod
  (substring `go.mod` required).
- Stdout empty (or no successful plan summary).

## Errors

- Cannot plan reinstalls without a parseable Go module root.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertEmptyStdout(t, resp.Stdout)
	assertContains(t, strings.ToLower(resp.Stderr), "go.mod")
}
```
