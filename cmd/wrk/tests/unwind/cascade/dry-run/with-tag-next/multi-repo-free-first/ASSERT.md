## Expected Output

B1 interleaved dry-run (early free peel → cascade → deferred consumer peel):

```
==== unwind (dry-run) ====
would: peel external/dot-pkgs-main-2026-06-30
would: tag-next example.com/dot-pkgs @ v0.0.2
would: pin example.com/root <- example.com/dot-pkgs @ v0.0.2
would: peel .
```

(Optional under-peel ship lines allowed. Optional root tag-next OK. Trailing newline.)

## Expected

- Exit code 0.
- Peel free-first: nested leaf external display path, then primary `.`.
- Module cascade free-first: **tag-next leaf module** before **pin root <- leaf**.
- **B1 interleave:** free peel → free tag-next/pin → deferred consumer peel
  (matches apply `splitPeelOrderB1`; not global peels-then-cascade).
- Zero mutations.

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
	for _, display := range req.PeelOrder {
		assertPeelUsesRelDisplay(t, resp.Stdout, display)
	}

	out := resp.Stdout
	tagLeaf := cascadeTagNextLine(unwindDotPkgsModule, unwindApplyNextTag)
	pinRoot := cascadePinLine(unwindRootModule, unwindDotPkgsModule, unwindApplyNextTag)
	if !strings.Contains(out, tagLeaf) && !hasCascadeTagNext(out, unwindDotPkgsModule) {
		t.Fatalf("missing tag-next for leaf module %s\nwant %q\nstdout:\n%s",
			unwindDotPkgsModule, tagLeaf, out)
	}
	if !hasCascadePin(out, unwindRootModule, unwindDotPkgsModule) {
		t.Fatalf("missing cascade pin root <- leaf\nwant %q\nstdout:\n%s", pinRoot, out)
	}
	// Free-first across stack: leaf module before consumer pin.
	assertContainsInOrder(t, out,
		"would: tag-next "+unwindDotPkgsModule+" @",
		"would: pin "+unwindRootModule+" <- "+unwindDotPkgsModule,
	)
	// B1 interleave: free peel → cascade free tag/pin → deferred consumer peel.
	if len(req.PeelOrder) >= 2 {
		assertContainsInOrder(t, out,
			peelLine(req.PeelOrder[0]),
			"would: tag-next "+unwindDotPkgsModule+" @",
			"would: pin "+unwindRootModule+" <- "+unwindDotPkgsModule,
			peelLine(req.PeelOrder[1]),
		)
	}
	assertUnwindZeroMutations(t, req)
}
```
