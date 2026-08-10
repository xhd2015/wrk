## Expected

- Exit code non-zero (prefer 1).
- Human banners present.
- `needs-land` shows `FAIL`.
- `result: fail`.
- No `Error:` for logical FAIL.
- Zero mutations (linked wt + externals remain).

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
	assertVerifyCheckStatus(t, resp.Stdout, "needs-land", "FAIL")
	assertVerifyResult(t, resp.Stdout, "fail")
	assertVerifyNoLogicalErrorPrefix(t, resp)
	assertVerifyZeroMutations(t, req)
}
```
