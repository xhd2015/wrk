## Expected

- Exit 0.
- Drops only `example.com/dep`.
- `replace example.com/kool => ./external/kool` still present.
- Summary undid 1.

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
	assertDropLine(t, resp.Stdout, modDep, false)
	assertNotContains(t, resp.Stdout, "drop  "+modKool)
	assertUndoSummary(t, resp.Stdout, 1, 1, 1, false)
	assertNoReplaceFor(t, req.ConsumerGoMod, modDep)
	assertHasReplaceFor(t, req.ConsumerGoMod, modKool)
}
```
