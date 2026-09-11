## Expected Output

```
==== unwind (dry-run) ====
would: peel .
```

## Expected

- Exit code 0.
- **Only** `would: peel .` — intra-repo `./pkgs/shared` is not a separate peel.
- Stdout has no peel line for the shared subpath display (if distinct from `.`).
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
	out := resp.Stdout
	assertPeelOrder(t, out, req.PeelOrder)
	assertPeelUsesRelDisplay(t, out, ".")
	// Shared path must not appear as its own peel/lane target.
	if req.DepsLinkedWtDir != "" {
		sharedDisp := peelDisplay(t, req, req.DepsLinkedWtDir)
		if sharedDisp != "." && hasPeelLine(out, sharedDisp) {
			t.Fatalf("intra-repo shared must not peel as %q\nstdout:\n%s", sharedDisp, out)
		}
	}
	// Phase format: at most one primary lane "."; legacy: exactly one would: peel.
	if n := strings.Count(out, "would: peel "); n > 1 {
		t.Fatalf("intra-repo: want at most 1 legacy peel line, got %d:\n%s", n, out)
	}
	assertUnwindZeroMutations(t, req)
}
```
