## Expected Output

Stdout:

```
==== unwind (dry-run) ====
would: peel .
```

Stderr (text, ANSI ignored): contains `warning:` (missing / non-git replace target).

## Expected

- Exit code 0 (do not fail unwind solely for missing replace target).
- Stderr contains `warning:`.
- Plan continues: `would: peel .` for dirty primary.
- Zero mutations.

## Side Effects

- None.

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
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)
	assertPeelUsesRelDisplay(t, resp.Stdout, ".")
	// Product: warning: prefix on stderr (yellow in TTY; harness is plain text).
	if !strings.Contains(resp.Stderr, "warning:") && !strings.Contains(resp.Stderr, "Warning:") {
		t.Fatalf("missing/non-git replace target must emit warning: on stderr; stderr=%q stdout=%q",
			resp.Stderr, resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
