## Expected Output

```
==== unwind (dry-run) ====
would: peel external/agent-pro-main-2026-06-30
would: peel .
```

(No peel line for clean leaf external/dot-pkgs-….)

## Expected

- Exit code 0.
- Peel order: mid external display path then primary `.`.
- Stdout does **not** contain a peel line for the clean leaf checkout display path.
- Zero mutations.

## Side Effects

- None.

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
	// Clean leaf must not appear as a peel (use full display path of leaf external).
	skipped := peelDisplay(t, req, req.DepsLinkedWtDir)
	assertNotContains(t, resp.Stdout, peelLine(skipped))
	assertUnwindZeroMutations(t, req)
}
```
