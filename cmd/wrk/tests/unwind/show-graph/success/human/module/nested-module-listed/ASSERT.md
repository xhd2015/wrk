## Expected Output

```
==== unwind graph (module) ====
  dir
  .
  pkgs/shared
  …
  edges (consumer → deps):
  .:
    → pkgs/shared   require v0.0.0  replaced
…
==== unwind graph (repo) ====
…
peel order (dirty, free-first):
  .
```

Single-repo: module identity = **dir** (`.`, `pkgs/shared`). Collapsed edges with `→`.
Replace annotation is **`replaced`** only (no `replace =>`).

## Expected

- Exit code 0.
- Human banners present.
- Module dirs `.` and `pkgs/shared` listed (not full `example.com/root…` as identity).
- Collapsed consumer edges with unicode `→`.
- Contains `replaced`; must **not** contain `replace =>`.
- No flat full-path edge lines.
- Peel order is only `.` (intra-repo shared is not a separate repo peel).
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
	assertShowGraphHumanBanners(t, resp.Stdout)
	assertModuleDirListed(t, resp.Stdout, ".")
	assertModuleDirListed(t, resp.Stdout, "pkgs/shared")
	assertCollapsedEdgesHuman(t, resp.Stdout)
	assertReplacedHuman(t, resp.Stdout)
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertHumanNoFlatFullPathEdges(t, resp.Stdout)
	assertShowGraphPeelOrderHuman(t, resp.Stdout, req.PeelOrder)
	assertShowGraphZeroMutations(t, req)
}
```
