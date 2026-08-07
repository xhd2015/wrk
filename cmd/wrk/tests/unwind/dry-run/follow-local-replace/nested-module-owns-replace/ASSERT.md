## Expected Output

```
==== unwind (dry-run) ====
would: peel ../external/dot-pkgs-main-2026-06-30
would: peel .
```

## Expected

- Exit code 0.
- Dep still in plan even though **only nested module** owns the replace.
- Peel order: dep display then `.`.
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
	assertUnwindZeroMutations(t, req)
}
```
