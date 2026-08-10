## Expected Output

```
==== unwind graph (module) ====
  …
  edges (consumer → deps):
  root:
    → dot-pkgs   require v0.0.0  replaced
                 (latest v0.0.1)
…
```

Drift marker `(latest` is locked; exact spacing implementer-owned.

## Expected

- Exit code 0.
- Human banners present.
- Contains `(latest` drift annotation (and preferably latest tag `v0.0.1`).
- Local replace still surfaces as `replaced` (not `replace =>`).
- Collapsed edges with `→`.
- Zero mutations.

## Side Effects

- None.

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
	assertShowGraphHumanBanners(t, resp.Stdout)
	assertRequireDriftHuman(t, resp.Stdout)
	latest := req.ExpectedPinVersion
	if latest == "" {
		latest = unwindApplyOldTag
	}
	if !strings.Contains(resp.Stdout, latest) {
		t.Fatalf("drift annotation should include dep latest %q; stdout:\n%s", latest, resp.Stdout)
	}
	assertCollapsedEdgesHuman(t, resp.Stdout)
	assertReplacedHuman(t, resp.Stdout)
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertHumanNoFlatFullPathEdges(t, resp.Stdout)
	assertShowGraphZeroMutations(t, req)
}
```
