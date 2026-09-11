## Expected Output

B1 interleaved dry-run + reinstall tail:

```
==== unwind (dry-run) ====
would: peel external/dot-pkgs-main-2026-06-30
would: tag-next example.com/dot-pkgs @ v0.0.2
would: pin example.com/root <- example.com/dot-pkgs @ v0.0.2
would: peel .
would: reinstall local binaries
```

(Optional under-peel ship lines allowed. Reinstall is the plan **tail** after
peels and cascade. B1 may place cascade between early and deferred peels.)

## Expected

- Exit code 0.
- Free-first peels present.
- Cascade tag-next for leaf module + pin root <- leaf present.
- Stdout contains exact tail line `would: reinstall local binaries`.
- Reinstall appears **after** cascade tag-next (when cascade present).
- Not mutually exclusive with `--unwind`.
- Zero mutations.

## Side Effects

- None. Dry-run does not reinstall or tag.

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

	out := resp.Stdout
	if !hasCascadeTagNext(out, unwindDotPkgsModule) {
		t.Fatalf("cascade tag-next for leaf required with --tag-next\nstdout:\n%s", out)
	}
	if !hasCascadePin(out, unwindRootModule, unwindDotPkgsModule) {
		t.Fatalf("cascade pin root <- leaf required\nstdout:\n%s", out)
	}
	// Phase format: would: reinstall-local; legacy: would: reinstall local binaries.
	reinstallOK := strings.Contains(out, "would: reinstall-local") ||
		strings.Contains(out, "would: reinstall local binaries")
	if !reinstallOK {
		t.Fatalf("missing reinstall tail\nstdout:\n%s", out)
	}
	// Tail after cascade tag when both present.
	reinstallNeedle := "would: reinstall-local"
	if !strings.Contains(out, reinstallNeedle) {
		reinstallNeedle = "would: reinstall local binaries"
	}
	assertContainsInOrder(t, out,
		"would: tag-next "+unwindDotPkgsModule+" @",
		reinstallNeedle,
	)
	combined := strings.ToLower(resp.Stdout + "\n" + resp.Stderr)
	if strings.Contains(combined, "mutually exclusive") {
		t.Fatalf("unwind+reinstall-local must be accepted; stderr=%q stdout=%q",
			resp.Stderr, resp.Stdout)
	}
	// B1 interleave OK: free peel → cascade → deferred consumer peel → reinstall.
	// Do not require global peels-then-cascade (assertCascadeAfterPeels).
	assertUnwindZeroMutations(t, req)
}
```
