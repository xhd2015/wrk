## Expected

- Exit 0.
- Stdout contains `pin-local example.com/consumer <- example.com/dep => ` with relative path.
- go.mod has relative replace for `example.com/dep` (not absolute).
- Summary includes `pin-locals: applied N, tidy ok M, tidy failed F` with N>=1, F=0.
- No `would:` vocabulary.

## Side Effects

- Consumer go.mod gains/normalizes relative replace to external/dep.

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
	assertNotContains(t, resp.Stdout, "would:")
	assertPinLocalLine(t, resp.Stdout, modConsumer, modDep)
	assertRelativeReplace(t, req.ConsumerGoMod, modDep)
	// Absolute path must not remain as NewPath.
	body := readFile(t, req.ConsumerGoMod)
	if strings.Contains(body, req.DepModDir) {
		t.Fatalf("go.mod must not use absolute dep path as replace target:\n%s", body)
	}
	assertSummaryApplied(t, resp.Stdout, 1, 1, 0)
}
```
