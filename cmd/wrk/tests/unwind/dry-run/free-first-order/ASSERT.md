## Expected Output

Dry-run peel plan (display = checkout relpath vs cwd), free-first:

```
==== unwind (dry-run) ====
would: peel external/dot-pkgs-main-2026-06-30
would: peel external/agent-pro-main-2026-06-30
would: peel .
```

(Banner punctuation implementer-owned; asserts lock `would: peel <display-path>`
order and external/ / `.` vocabulary. Optional per-flag ship `would:` lines allowed.)

## Expected

- Exit code 0.
- Stdout contains peel lines in free-first order with **relative display paths**
  (nested `external/…` then primary `.`).
- Stdout does **not** use bare MainRepo basename alone as the full peel path
  for nested members (must include `external/`).
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
	for _, display := range req.PeelOrder {
		assertPeelUsesRelDisplay(t, resp.Stdout, display)
	}
	assertUnwindZeroMutations(t, req)
}
```
