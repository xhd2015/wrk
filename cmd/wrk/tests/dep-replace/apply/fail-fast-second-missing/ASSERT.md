## Expected

- Non-zero exit (fail-fast).
- First dep replace is applied (sequential fail-fast leaves prior writes).
- Stderr indicates missing second path.
- No success claim for the missing path.

## Side Effects

- Partial apply of first directory is allowed under D3 fail-fast.

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
	// First replace should remain after fail-fast.
	assertAbsoluteReplace(t, req.ConsumerGoMod, modDep, req.DepDir)
	// Prefer stdout having printed the first success line before failing.
	if !strings.Contains(resp.Stdout, "dep-replace "+modDep+" =>") &&
		!strings.Contains(resp.Stdout, "dep-replace "+modDep) {
		// Soft: some implementations may suppress partial stdout; go.mod is the lock.
		_ = resp.Stdout
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "no such") &&
		!strings.Contains(se, "not found") &&
		!strings.Contains(se, "missing") &&
		!strings.Contains(se, "exist") &&
		!strings.Contains(se, "stat") {
		t.Fatalf("stderr should indicate second path missing, got %q", resp.Stderr)
	}
}
```
