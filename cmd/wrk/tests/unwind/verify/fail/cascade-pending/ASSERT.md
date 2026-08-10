## Expected

- Exit code non-zero (prefer 1).
- Human banners present.
- `cascade-pending` shows `FAIL`.
- `result: fail`.
- No `Error:` for logical FAIL.
- Zero mutations (no tags/pins applied).

## Side Effects

- None.

## Exit Code

- 1

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertVerifyHumanBanners(t, resp.Stdout)
	assertVerifyCheckStatus(t, resp.Stdout, "cascade-pending", "FAIL")
	assertVerifyResult(t, resp.Stdout, "fail")
	assertVerifyNoLogicalErrorPrefix(t, resp)
	assertVerifyZeroMutations(t, req)
}
```
