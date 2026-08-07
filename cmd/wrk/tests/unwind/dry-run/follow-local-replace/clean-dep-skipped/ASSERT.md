## Expected Output

```
==== unwind (dry-run) ====
would: peel .
```

## Expected

- Exit code 0.
- Only primary `.` peels; clean sibling dep display is absent from peel lines.
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
	skipped := peelDisplay(t, req, req.DepsLinkedWtDir)
	if hasPeelLine(resp.Stdout, skipped) {
		t.Fatalf("clean dep must not appear as peel line %q\nstdout:\n%s", peelLine(skipped), resp.Stdout)
	}
	assertUnwindZeroMutations(t, req)
}
```
