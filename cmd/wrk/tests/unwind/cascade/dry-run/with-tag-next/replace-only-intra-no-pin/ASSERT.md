## Expected Output

```
==== unwind (dry-run) ====
would: peel .
```

(No cascade `would: tag-next` / `would: pin … <- …` solely because of intra-repo replace.)

## Expected

- Exit code 0.
- Peel `.` present.
- **No** cascade module lines (`would: tag-next` or cascade pin with ` <- `).
- Zero mutations.

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
	assertPeelUsesRelDisplay(t, resp.Stdout, ".")
	// D4: intra keep-local replace alone must not invent cascade pin/tag lines.
	assertNoCascadeModuleLines(t, resp.Stdout)
	assertUnwindZeroMutations(t, req)
}
```
