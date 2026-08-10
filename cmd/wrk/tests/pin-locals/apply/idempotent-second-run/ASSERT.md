## Expected

- Exit 0 on second run (once product exists; RED until then may be non-zero unknown flag).
- When product is present: already-up-to-date style; go.mod unchanged vs post-first baseline.
- No additional apply pin-local work lines on second run.

## Exit Code

- 0 (product contract; Classic TDD RED allowed before implement)

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	// Before product: non-zero unknown flag is the RED signal.
	// After product: exit 0 + already message + go.mod stable.
	if resp.ExitCode != 0 {
		// RED path: surface clearly
		t.Fatalf("second pin-locals expected exit 0 once implemented; got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	assertAlreadyUpToDate(t, resp)
	assertGoModUnchanged(t, req)
	if strings.Contains(resp.Stdout, "pin-locals:") {
		assertSummaryApplied(t, resp.Stdout, 0, 0, 0)
	}
}
```
