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
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)
	assertPeelUsesRelDisplay(t, resp.Stdout, ".")
	// Shared path must not appear as its own peel target (whole-line match).
	if req.DepsLinkedWtDir != "" {
		sharedDisp := peelDisplay(t, req, req.DepsLinkedWtDir)
		if sharedDisp != "." && hasPeelLine(resp.Stdout, sharedDisp) {
			t.Fatalf("intra-repo shared must not peel as %q\nstdout:\n%s", peelLine(sharedDisp), resp.Stdout)
		}
	}
	// Exactly one peel line (primary only).
	if strings.Count(resp.Stdout, "would: peel ") != 1 {
		t.Fatalf("intra-repo: want exactly 1 peel line, got:\n%s", resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
