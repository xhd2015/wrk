## Expected Output

```
==== unwind graph (repo) ====
  …
  peel order (dirty, free-first):
    external/dot-pkgs-main-2026-06-30
    external/agent-pro-main-2026-06-30
    .

==== unwind graph (module) ====
  modules @ …
…
==== status summary ====
…
```

(Node line format implementer-owned; peel free-first display paths locked.)

## Expected

- Exit code 0.
- Human banners present.
- Peel order free-first: leaf external → mid external → primary `.`.
- Nested peel displays include `external/` (not bare MainRepo basename alone).
- Repo labels listed: root, agent-pro, dot-pkgs.
- Multi-repo module grouping when ≥2 repos (`modules @`).
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
	assertShowGraphPeelOrderHuman(t, resp.Stdout, req.PeelOrder)
	for _, label := range []string{labelRoot, labelAgentPro, labelDotPkgs} {
		assertRepoNodeListed(t, resp.Stdout, label)
	}
	if !strings.Contains(resp.Stdout, "external/") {
		t.Fatalf("nested peel/display must use external/ relative paths; stdout:\n%s", resp.Stdout)
	}
	assertModulesGroupedByRepo(t, resp.Stdout)
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertHumanNoFlatFullPathEdges(t, resp.Stdout)
	assertShowGraphZeroMutations(t, req)
}
```
