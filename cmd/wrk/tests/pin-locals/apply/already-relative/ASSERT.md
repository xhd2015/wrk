## Expected

- Exit 0.
- Already-up-to-date style message.
- No apply pin-local work lines.
- go.mod unchanged from baseline.
- If summary present, applied 0.

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
	assertAlreadyUpToDate(t, resp)
	assertGoModUnchanged(t, req)
	if strings.Contains(resp.Stdout, "pin-locals:") {
		assertSummaryApplied(t, resp.Stdout, 0, 0, 0)
	}
}
```
