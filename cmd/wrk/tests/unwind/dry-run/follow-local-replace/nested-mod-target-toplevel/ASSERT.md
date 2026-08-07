## Expected Output

```
==== unwind (dry-run) ====
would: peel ../external/dep-main-2026-06-30
would: peel .
```

## Expected

- Exit code 0.
- Peel display for the dep is the **git toplevel** Path (`../external/dep-main-…`),
  not the nested module subdir alone (`…/nested`).
- Free-first: dep toplevel then primary `.`.
- Zero mutations.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)
	for _, display := range req.PeelOrder {
		assertPeelUsesRelDisplay(t, resp.Stdout, display)
	}
	// Must not peel the nested module subdir as if it were the stack Path.
	// Use whole-line peel match: toplevel display is a prefix of …/nested display.
	nestedCheckout := filepath.Join(req.DepsLinkedWtDir, "nested")
	nestedDisp := peelDisplay(t, req, nestedCheckout)
	toplevelDisp := peelDisplay(t, req, req.DepsLinkedWtDir)
	if nestedDisp != toplevelDisp && hasPeelLine(resp.Stdout, nestedDisp) {
		t.Fatalf("must not peel nested module subdir alone %q\nstdout:\n%s", peelLine(nestedDisp), resp.Stdout)
	}
	if !hasPeelLine(resp.Stdout, toplevelDisp) {
		t.Fatalf("want dep git toplevel peel %q\nstdout:\n%s", peelLine(toplevelDisp), resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
