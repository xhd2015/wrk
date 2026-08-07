## Expected Output

```
==== unwind (dry-run) ====
would: peel ../external/dot-pkgs-main-2026-06-30
would: peel .
```

(Exact `../external/…` display is locked via `peelDisplay` / `req.PeelOrder`.)

## Expected

- Exit code 0.
- Free-first peel: **sibling dep display path** then primary `.`.
- Zero mutations (HEADs + dirty markers unchanged).

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
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
	// Sibling must not be mis-rendered as nested external/ under primary alone.
	depDisplay := peelDisplay(t, req, req.DepsLinkedWtDir)
	if depDisplay == "." {
		t.Fatalf("sibling dep display must not collapse to primary cwd")
	}
	if !hasPeelLine(resp.Stdout, depDisplay) {
		t.Fatalf("missing whole-line peel for sibling dep %q\nstdout:\n%s", peelLine(depDisplay), resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
