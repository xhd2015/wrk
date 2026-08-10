## Expected

- Exit 0.
- pin-local line for root <- tools.
- go.mod replace NewPath is `./tools` (or equivalent relative).
- Summary applied >= 1, tidy failed 0.

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
	assertPinLocalLine(t, resp.Stdout, modRoot, modTools)
	assertRelativeReplace(t, req.ConsumerGoMod, modTools)
	body := readFile(t, req.ConsumerGoMod)
	if !strings.Contains(body, "./tools") && !strings.Contains(body, "tools") {
		t.Fatalf("expected relative tools replace, go.mod:\n%s", body)
	}
	// Prefer exact ./tools
	if !strings.Contains(body, "=> ./tools") && !strings.Contains(body, "=>./tools") {
		// accept if pin-local line locked the path
		assertContains(t, resp.Stdout, "=> ./tools")
	}
	assertSummaryApplied(t, resp.Stdout, 1, 1, 0)
}
```
