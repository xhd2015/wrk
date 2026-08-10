## Expected Output

Replace-only cascade pin for droppable external replace (no leaf tag-next).
Pure pin-consumer peel is B1-deferred after cascade (matches apply):

```
==== unwind (dry-run) ====
would: pin example.com/root <- example.com/dot-pkgs @ v0.0.1
would: peel .
```

(Optional under-peel ship lines from `--tag-next` allowed. Optional root
`would: tag-next example.com/root @ …` is allowed when root itself is
owned-changed past its baseline tag — **must not** replace the pin contract.
Trailing newline. **No** `would: tag-next` for the clean leaf; **no** peel for
`external/…`.)

## Expected

- Exit code 0.
- Peel display `.` only (clean free dep not peeled).
- Cascade pin: `would: pin example.com/root <- example.com/dot-pkgs @ v0.0.1`
  (version = current require / latest tag — D3 keep-current).
- **No** cascade `would: tag-next` for `example.com/dot-pkgs` (not owned-changed).
- B1: pure pin-consumer peel may follow cascade pin (not peels-then-cascade).
- Zero mutations (HEAD + DIRTY preserved).

## Side Effects

- None (plan only).

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertPeelOrder(t, resp.Stdout, req.PeelOrder)
	assertPeelUsesRelDisplay(t, resp.Stdout, ".")

	out := resp.Stdout

	// Clean free dep must not peel (whole-line display match).
	if req.DepsLinkedWtDir != "" {
		skipped := peelDisplay(t, req, req.DepsLinkedWtDir)
		if hasPeelLine(out, skipped) {
			t.Fatalf("clean external leaf must not peel %q\nstdout:\n%s", peelLine(skipped), out)
		}
	}

	// No tag-next on clean leaf (no owned-changed).
	if hasCascadeTagNext(out, unwindDotPkgsModule) {
		t.Fatalf("clean free dep must not get cascade tag-next\nstdout:\n%s", out)
	}

	// Core: external replace alone ⇒ needs-pin at current require version.
	pinRoot := cascadePinLine(unwindRootModule, unwindDotPkgsModule, unwindApplyOldTag)
	if !strings.Contains(out, pinRoot) {
		// Allow hasCascadePin + version fragment if spacing differs slightly.
		if !hasCascadePin(out, unwindRootModule, unwindDotPkgsModule) ||
			!strings.Contains(out, " @ "+unwindApplyOldTag) {
			t.Fatalf("missing replace-only cascade pin at current require\nwant %q\nstdout:\n%s",
				pinRoot, out)
		}
	}

	// B1: consumer peel is deferred pure pin-consumer — pin may precede peel.
	// Content contract above is the lock; order matches apply splitPeelOrderB1.
	assertUnwindZeroMutations(t, req)
}
```
