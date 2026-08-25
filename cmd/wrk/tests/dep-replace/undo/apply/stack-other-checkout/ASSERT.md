## Expected

- Exit 0.
- Undoes on primary and `external/kool`.
- Neither go.mod has dep replace afterward.
- HEAD kool replace on primary (if present) stays; only introduced dep drops.

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
	assertUndoBanner(t, resp.Stdout, false)
	assertContains(t, resp.Stdout, "checkout  .")
	assertContains(t, resp.Stdout, "checkout  "+checkoutKool)
	assertDropLine(t, resp.Stdout, modDep, false)
	assertUndoSummary(t, resp.Stdout, 2, 2, 2, false)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertNoReplaceFor(t, req.Consumer2GoMod, modDep)
	assertHasReplaceFor(t, req.ConsumerGoMod, modKool)
}
```
