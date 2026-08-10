## Expected

- Exit 0.
- pin-local line for consumer <- dep with relative path.
- go.mod no longer uses absolute DepModDir as replace NewPath.
- Relative replace present.

## Side Effects

- Absolute replace rewritten to relative.

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
	assertPinLocalLine(t, resp.Stdout, modConsumer, modDep)
	assertRelativeReplace(t, req.ConsumerGoMod, modDep)
	body := readFile(t, req.ConsumerGoMod)
	// Absolute path from baseline must be gone as replace target.
	if strings.Contains(body, req.DepModDir) {
		t.Fatalf("absolute replace target should be rewritten; go.mod:\n%s", body)
	}
	assertSummaryApplied(t, resp.Stdout, 1, 1, 0)
}
```
