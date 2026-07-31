## Expected Output

Dry-run peel plan (labels = main-repo basenames), free-first:

```
==== unwind (dry-run) ====
would: peel dot-pkgs
would: peel agent-pro
would: peel root
```

(Banner punctuation implementer-owned; asserts lock `would: peel <label>` order
and unwind/dry-run intent. Optional per-flag ship `would:` lines allowed.)

## Expected

- Exit code 0.
- Stdout contains peel lines in order: `dot-pkgs` → `agent-pro` → `root`.
- Stdout mentions unwind (banner or plan).
- Zero mutations: worktrees still linked; HEADs match baseline; dirty files remain.

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
	assertUnwindZeroMutations(t, req)
}
```
