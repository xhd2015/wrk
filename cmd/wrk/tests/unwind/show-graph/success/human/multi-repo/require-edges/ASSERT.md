## Expected Output

```
==== unwind graph (repo) ====
  …
  edges (From depends on To):
  root:
    → agent-pro
  agent-pro:
    → dot-pkgs

==== unwind graph (module) ====
  modules @ agent-pro (…):
    dir  .
  modules @ dot-pkgs (…):
    dir  .
  modules @ root (.):
    dir  .

  edges (consumer → deps):
  root:
    → agent-pro   require …
  agent-pro:
    → dot-pkgs    require …
```

(Collapsed `→` form locked; full module paths forbidden in human edges.)

## Expected

- Exit code 0.
- Human banners present.
- Repo edges connect labels root → agent-pro → dot-pkgs (depends-on; collapsed).
- Module multi-repo grouping `modules @` present.
- Module edges mention `require` and use short keys / `→` (not flat full paths).
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
	out := resp.Stdout
	// Repo short names present.
	assertRepoNodeListed(t, out, labelRoot)
	assertRepoNodeListed(t, out, labelAgentPro)
	assertRepoNodeListed(t, out, labelDotPkgs)
	// Collapsed human edges (repo and/or module).
	assertCollapsedEdgesHuman(t, out)
	assertModulesGroupedByRepo(t, out)
	assertHumanNoFlatFullPathEdges(t, out)
	// Require kind somewhere in module section.
	if !strings.Contains(strings.ToLower(moduleSection(out)), "require") {
		t.Fatalf("module graph must show require edges; stdout:\n%s", out)
	}
	// Cross-repo dependency direction via labels (order of appearance soft).
	assertContainsInOrder(t, out, labelRoot, labelAgentPro)
	assertContainsInOrder(t, out, labelAgentPro, labelDotPkgs)
	assertHumanNoFullModulePaths(t, resp.Stdout)
	assertShowGraphZeroMutations(t, req)
}
```
