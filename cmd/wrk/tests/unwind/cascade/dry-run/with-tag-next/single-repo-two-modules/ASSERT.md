## Expected Output

Peel plan then free-first module cascade (after peels):

```
==== unwind (dry-run) ====
would: peel .
would: tag-next example.com/root/shared @ pkgs/shared/v0.0.2
would: pin example.com/root <- example.com/root/shared @ v0.0.2
```

(Optional under-peel ship lines and optional root `would: tag-next example.com/root @ …`
after the pin are allowed. Trailing newline required.)

## Expected

- Exit code 0.
- Peel display `.` present.
- Cascade free-first: **tag-next shared** before **pin root <- shared**.
- Tag line uses module path `example.com/root/shared` and next tag containing
  `v0.0.2` (full path-scope tag `pkgs/shared/v0.0.2` preferred).
- Pin line uses module paths (not bare MainRepo basename alone as the only form).
- Cascade lines appear **after** peel lines.
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
	// Free-first: tag shared leaf before pin consumer root.
	tagShared := cascadeTagNextLine(cascadeSharedModule, cascadeSharedNextTag)
	pinRoot := cascadePinLine(unwindRootModule, cascadeSharedModule, unwindApplyNextTag)
	// Allow slightly different next-tag spelling if implementer uses version-only
	// on tag-next: still require module path + v0.0.2 and ordered pin.
	if !strings.Contains(out, tagShared) {
		// Fallback: module path with any next tag that mentions v0.0.2.
		if !hasCascadeTagNext(out, cascadeSharedModule) || !strings.Contains(out, "v0.0.2") {
			t.Fatalf("missing free-first tag-next for shared leaf\nwant %q (or tag-next %s @ …v0.0.2)\nstdout:\n%s",
				tagShared, cascadeSharedModule, out)
		}
	}
	if !hasCascadePin(out, unwindRootModule, cascadeSharedModule) {
		t.Fatalf("missing cascade pin root <- shared\nwant substring pin %s <- %s\nstdout:\n%s",
			unwindRootModule, cascadeSharedModule, out)
	}
	// Order: tag shared before pin root (free-first).
	assertContainsInOrder(t, out,
		"would: tag-next "+cascadeSharedModule+" @",
		"would: pin "+unwindRootModule+" <- "+cascadeSharedModule,
	)
	// Prefer exact pin version vocabulary when present.
	if !strings.Contains(out, pinRoot) && !strings.Contains(out, " @ "+unwindApplyNextTag) {
		t.Fatalf("pin line should include version %s\nstdout:\n%s", unwindApplyNextTag, out)
	}
	assertCascadeAfterPeels(t, out, req.PeelOrder)
	assertUnwindZeroMutations(t, req)
}
```
