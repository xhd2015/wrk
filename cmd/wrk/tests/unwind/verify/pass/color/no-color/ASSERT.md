## Expected

- Exit code 0.
- Human banners present.
- Stdout has no ANSI escapes.
- All checks pass; `result: pass`.
- Zero mutations.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertVerifyHumanBanners(t, resp.Stdout)
	assertVerifyNoANSI(t, resp.Stdout)
	assertVerifyAllChecksPass(t, resp.Stdout)
	assertVerifyResult(t, resp.Stdout, "pass")
	assertVerifyZeroMutations(t, req)
}
```
