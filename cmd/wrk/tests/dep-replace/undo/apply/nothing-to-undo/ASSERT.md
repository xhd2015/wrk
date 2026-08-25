## Expected

- Exit 0.
- Stdout contains `dep-replace: nothing to undo`.
- No undo banner / drop lines.
- go.mod unchanged.

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
	assertContains(t, resp.Stdout, "dep-replace: nothing to undo")
	assertNotContains(t, resp.Stdout, "==== dep-replace --undo")
	assertGoModUnchanged(t, req)
}
```
