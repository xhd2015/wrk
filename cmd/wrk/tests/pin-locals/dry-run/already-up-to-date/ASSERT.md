## Expected

- Exit 0.
- Already-up-to-date style message (mentions already and/or up to date).
- No `would: pin-local` work lines.
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
	assertAlreadyUpToDate(t, resp)
	assertNotContains(t, resp.Stdout, "would: pin-local ")
	assertGoModUnchanged(t, req)
}
```
