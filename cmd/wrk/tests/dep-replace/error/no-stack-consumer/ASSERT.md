## Expected

- Non-zero exit.
- Stderr is a `wrk:` error containing `replace` or `consumer` (or equivalent).
- No banner. Both go.mods unchanged. No successful replace line.

## Errors

- Zero gated consumers on the whole stack (do **not** fall back to “always write nearest”).

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	assertNoBanner(t, resp.Stdout)
	assertNotContains(t, resp.Stdout, "dep-replace ")
	assertGoModUnchanged(t, req)
	assertGoModUnchangedAt(t, req.Consumer2GoMod, req.Baseline2GoMod)
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "wrk:") {
		t.Fatalf("stderr should be wrk: error, got %q", resp.Stderr)
	}
	if !strings.Contains(se, "replace") && !strings.Contains(se, "consumer") {
		t.Fatalf("stderr should mention replace or consumer, got %q", resp.Stderr)
	}
}
```
