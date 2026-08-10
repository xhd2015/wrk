## Expected

- Exit 0.
- Stdout may plan pin for `example.com/dep` (already required).
- Stdout must **not** contain a would: pin-local line for `example.com/other`.
- go.mod unchanged.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	// Must not invent replace for inventory-only other.
	assertNotContains(t, resp.Stdout, "<- "+modOther)
	assertNotContains(t, resp.Stdout, modOther+" =>")
	// Dep is a real dependency — plan may include it.
	if strings.Contains(resp.Stdout, "would: pin-local") {
		assertWouldPinLocalLine(t, resp.Stdout, modConsumer, modDep)
	}
	assertGoModUnchanged(t, req)
}
```
