## Expected

- Exit code non-zero.
- Combined output mentions `verify` and `merge-back`.
- No successful verify body.
- Zero mutations.

## Side Effects

- None.

## Exit Code

- non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertVerifyReject(t, resp, "merge-back")
	assertVerifyZeroMutations(t, req)
}
```
