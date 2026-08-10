## Expected

- Exit code 0.
- Human banners present (plain or ANSI-wrapped).
- Stdout contains at least one ANSI CSI escape (`\x1b[`).
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
	assertShowGraphHumanBannersMaybeColored(t, resp.Stdout)
	assertShowGraphHasANSI(t, resp.Stdout)
	assertShowGraphZeroMutations(t, req)
}
```
