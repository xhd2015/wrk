## Expected Output

```
==== unwind graph (module) ====
  modules @ dot-pkgs (…):
    dir  …
    .
  modules @ root (.):
    dir  …
    .

  edges (consumer → deps):
  root:
    → dot-pkgs   require v0.0.0  replaced
…
peel order (dirty, free-first):
  ../external/dot-pkgs-main-…
  .
```

Multi-repo keys: `label` / `label/dir`. Replace → word **`replaced`** only.

## Expected

- Exit code 0.
- Human banners present.
- Multi-repo module grouping (`modules @`) **or** multi-repo keys using labels.
- Edge uses `replaced` (require + replace may share one line).
- Must **not** contain `replace =>`.
- Collapsed edges with `→`.
- No flat full-path module edges.
- Peel order free-first: dep display then `.`.
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
	// Multi-repo: group header and/or label keys (not bare full module paths).
	sec := moduleSection(resp.Stdout)
	if !strings.Contains(sec, "modules @") &&
		!strings.Contains(sec, labelRoot) &&
		!strings.Contains(sec, labelDotPkgs) {
		t.Fatalf("multi-repo module section must use modules @ or label keys; section:\n%s", sec)
	}
	// Prefer grouping when ≥2 repos.
	assertModulesGroupedByRepo(t, resp.Stdout)
	assertCollapsedEdgesHuman(t, resp.Stdout)
	assertReplacedHuman(t, resp.Stdout)
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertHumanNoFlatFullPathEdges(t, resp.Stdout)
	assertShowGraphPeelOrderHuman(t, resp.Stdout, req.PeelOrder)
	assertShowGraphZeroMutations(t, req)
}
```
