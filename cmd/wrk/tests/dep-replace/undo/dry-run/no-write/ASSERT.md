## Expected

- Exit 0.
- Dry-run undo banner; `would: drop`; `would: skip tidy  (vendor/)`.
- go.mod unchanged vs baseline (still has introduced replace).

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
	assertUndoBanner(t, resp.Stdout, true)
	assertDropLine(t, resp.Stdout, modDep, true)
	assertContains(t, resp.Stdout, "would: skip tidy  (vendor/)")
	assertUndoSummary(t, resp.Stdout, 1, 1, 1, true)
	assertGoModUnchanged(t, req)
	assertHasReplaceFor(t, req.ConsumerGoMod, modDep)
}
```
