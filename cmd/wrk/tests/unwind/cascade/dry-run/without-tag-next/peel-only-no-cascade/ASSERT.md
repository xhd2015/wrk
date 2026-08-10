## Expected Output

```
==== unwind (dry-run) ====
would: peel .
```

(No cascade `would: tag-next` / `would: pin … <- …` lines.)

## Expected

- Exit code 0.
- Peel `.` present.
- **No** cascade module lines (`would: tag-next` or cascade pin with ` <- `).
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
	assertPeelUsesRelDisplay(t, resp.Stdout, ".")
	assertNoCascadeModuleLines(t, resp.Stdout)
	assertUnwindZeroMutations(t, req)
}
```
