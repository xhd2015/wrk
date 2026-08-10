## Expected Output

Intended B1 interleaved dry-run (early free → cascade → deferred consumer):

```
==== unwind (dry-run) ====
would: peel external/dot-pkgs-main-2026-06-30
  would: merge-back linked worktree into main
  would: create release tag
  would: push branch and created tag
  would: pin stack consumers
would: tag-next example.com/dot-pkgs @ v0.0.2
would: pin example.com/root <- example.com/dot-pkgs @ v0.0.2
would: peel .
  would: merge-back linked worktree into main
  would: create release tag
  would: push branch and created tag
  would: pin stack consumers
```

(Under-peel ship lines optional/variable. Noise pin `root <- shared @ …` may
appear in the cascade block. Free peel display is `external/…` relative path.
Trailing newline. Exact under-peel set is soft; **order** of free peel → free
tag-next → consumer free pin → consumer peel is locked.)

## Expected

- Exit code 0.
- Free-first peel identity: free external display before consumer primary
  (PeelOrder free-first among peels).
- **B1 interleave (intended):**
  1. early `would: peel <free external>`
  2. cascade: `would: tag-next example.com/dot-pkgs @ …v0.0.2` then
     `would: pin example.com/root <- example.com/dot-pkgs` (noise shared pin OK)
  3. deferred `would: peel <consumer>` **after** free cascade pin of free
- Consumer peel must **not** appear before free `would: tag-next` (false freeHost
  early peel of monorepo is wrong on dry-run just as on apply).
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
	out := resp.Stdout

	if len(req.PeelOrder) < 2 {
		t.Fatal("G1 PeelOrder needs free then consumer displays")
	}
	freeDisplay := req.PeelOrder[0]
	consDisplay := req.PeelOrder[1]

	// Free-first among peel lines (even when cascade interleaves).
	assertPeelOrder(t, out, req.PeelOrder)
	assertPeelUsesRelDisplay(t, out, freeDisplay)

	// Cascade free tag-next + consumer pin of free present.
	if !hasCascadeTagNext(out, unwindDotPkgsModule) || !strings.Contains(out, "v0.0.2") {
		t.Fatalf("G1: missing would: tag-next for free %s @ …v0.0.2\nstdout:\n%s",
			unwindDotPkgsModule, out)
	}
	if !hasCascadePin(out, unwindRootModule, unwindDotPkgsModule) {
		t.Fatalf("G1: missing would: pin %s <- %s\nstdout:\n%s",
			unwindRootModule, unwindDotPkgsModule, out)
	}
	assertContainsInOrder(t, out,
		"would: tag-next "+unwindDotPkgsModule+" @",
		"would: pin "+unwindRootModule+" <- "+unwindDotPkgsModule,
	)

	// B1 interleave: free peel → free tag-next → free pin → deferred consumer peel.
	// This is the intended user-facing dry-run order (may RED while FormatUnwindDryRun
	// still emits peels-then-cascade globally).
	assertContainsInOrder(t, out,
		peelLine(freeDisplay),
		"would: tag-next "+unwindDotPkgsModule+" @",
		"would: pin "+unwindRootModule+" <- "+unwindDotPkgsModule,
		peelLine(consDisplay),
	)

	// Explicit false freeHost guard: consumer peel must not precede free tag-next.
	consIdx := indexPeelLine(out, consDisplay)
	tagIdx := strings.Index(out, "would: tag-next "+unwindDotPkgsModule+" @")
	if consIdx < 0 || tagIdx < 0 {
		t.Fatalf("G1: missing consumer peel or free tag-next\nstdout:\n%s", out)
	}
	if consIdx < tagIdx {
		t.Fatalf("G1 B1 order: consumer peel must be deferred after free tag-next cascade\nconsumer peel@%d free tag-next@%d\nstdout:\n%s",
			consIdx, tagIdx, out)
	}

	assertUnwindZeroMutations(t, req)
}
```
