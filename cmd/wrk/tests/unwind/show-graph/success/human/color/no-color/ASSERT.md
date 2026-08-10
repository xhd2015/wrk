## Expected

- Exit code 0.
- Human banners present.
- Stdout has **no** ANSI CSI escapes.
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
	assertShowGraphHumanBanners(t, resp.Stdout)
	assertShowGraphNoANSI(t, resp.Stdout)
	assertShowGraphZeroMutations(t, req)
}
```
