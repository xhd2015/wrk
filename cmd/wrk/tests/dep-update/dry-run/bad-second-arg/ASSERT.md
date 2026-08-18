## Expected

- Non-zero exit.
- Stderr is a `wrk:` error indicating missing dir (`no such dir` or equivalent).
- No banner; no checkout/module/pin tree; first dep is not described as a half-plan.
- Consumer go.mod unchanged.

## Errors

- Validate every dir arg first; any bad arg aborts before the plan.

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
	assertGoModUnchanged(t, req)
	assertNoBanner(t, resp.Stdout)
	assertNotContains(t, resp.Stdout, "checkout")
	assertNotContains(t, resp.Stdout, "would: pin")
	assertNotContains(t, resp.Stdout, "pin  ")
	// First dep must not be described as a half-plan.
	assertNotContains(t, resp.Stdout, "dep  "+modDep)
	assertNotContains(t, resp.Stderr, "dep  "+modDep)
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "wrk:") {
		t.Fatalf("stderr should be wrk: error, got %q", resp.Stderr)
	}
	if !strings.Contains(se, "no such") &&
		!strings.Contains(se, "not found") &&
		!strings.Contains(se, "missing") &&
		!strings.Contains(se, "exist") &&
		!strings.Contains(se, "stat") {
		t.Fatalf("stderr should indicate missing dir, got %q", resp.Stderr)
	}
}
```
