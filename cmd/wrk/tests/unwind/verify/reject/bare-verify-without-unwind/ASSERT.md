## Expected Output

```
Error: wrk: --verify requires --unwind
```

(Exact wording implementer-owned; must mention `verify` and `unwind`.)

## Expected

- Exit code non-zero.
- Combined stdout/stderr mentions `verify` and `unwind`.
- No successful verify banners / `result:`.
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
	assertVerifyReject(t, resp, "unwind")
	assertVerifyZeroMutations(t, req)
}
```
