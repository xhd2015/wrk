## Expected Output

```
==== unwind graph (repo) ====
  path  label  kind  dirt  …
  .     root   main  clean  …

  edges: (none)
  peel order (dirty, free-first): (none)

==== unwind graph (module) ====
  dir  latest  status
  .    …

==== status summary ====
  …
```

Human identity = **dir** (`.`), not full module path. No flat full-path edges.

## Expected

- Exit code 0.
- Human graph banners present.
- Peel order empty / `(none)`.
- Repo node for primary (label `root` and/or display `.`).
- Module section lists dir `.` (single-repo identity).
- Does **not** use flat full-path module edges (`example.com/… -> …`).
- Summary present.
- Zero mutations (HEAD unchanged).

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
	assertShowGraphPeelOrderHuman(t, resp.Stdout, nil)
	assertRepoNodeListed(t, resp.Stdout, labelRoot)
	assertModuleDirListed(t, resp.Stdout, ".")
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertHumanNoFlatFullPathEdges(t, resp.Stdout)
	assertNotContains(t, resp.Stdout, "would: peel ")
	assertShowGraphZeroMutations(t, req)
}
```
