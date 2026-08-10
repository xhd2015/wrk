## Expected

- Exit code 0.
- Human banners present.
- All six checks `pass` (require matches latest; no droppable external replace).
- `result: pass`.
- Stdout trailing `\n`.
- Zero mutations (root + nested external HEADs).

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
	assertVerifyAllChecksPass(t, resp.Stdout)
	assertVerifyResult(t, resp.Stdout, "pass")
	assertVerifyStdoutTrailingNL(t, resp.Stdout)
	assertVerifyZeroMutations(t, req)
}
```
