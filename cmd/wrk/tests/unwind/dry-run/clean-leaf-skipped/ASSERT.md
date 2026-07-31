## Expected Output

```
==== unwind (dry-run) ====
would: peel agent-pro
would: peel root
```

(No `would: peel dot-pkgs` — clean leaf skipped from pending.)

## Expected

- Exit code 0.
- Peel order: `agent-pro` then `root`.
- Stdout does **not** contain `would: peel dot-pkgs`.
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
	assertNotContains(t, resp.Stdout, peelLine(labelDotPkgs))
	assertUnwindZeroMutations(t, req)
}
```
