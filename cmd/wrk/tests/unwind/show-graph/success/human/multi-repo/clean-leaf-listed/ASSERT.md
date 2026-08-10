## Expected Output

```
==== unwind graph (repo) ====
  nodes/table:
    … dot-pkgs … clean …
    … agent-pro … dirty …
    … root … dirty …
  peel order (dirty, free-first):
    external/agent-pro-main-2026-06-30
    .
```

## Expected

- Exit code 0.
- Human banners present.
- Repo nodes list all three labels including clean **dot-pkgs**.
- Peel order is mid external then `.` only — clean leaf display must not appear
  in the peel-order section as a peel step.
- Multi-repo human module grouping (`modules @`) present.
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
	// Clean leaf still listed as a node.
	assertRepoNodeListed(t, resp.Stdout, labelDotPkgs)
	assertRepoNodeListed(t, resp.Stdout, labelAgentPro)
	assertRepoNodeListed(t, resp.Stdout, labelRoot)
	assertModulesGroupedByRepo(t, resp.Stdout)
	assertHumanNoFullModulePaths(t, resp.Stdout)

	// Clean leaf display path must not appear inside peel-order section.
	leafDisp := peelDisplay(t, req, req.DepsLinkedWtDir)
	lower := strings.ToLower(resp.Stdout)
	peelIdx := strings.Index(lower, graphPeelSection)
	if peelIdx < 0 {
		t.Fatalf("missing peel order section\nstdout:\n%s", resp.Stdout)
	}
	section := resp.Stdout[peelIdx:]
	for _, stop := range []string{graphModuleBanner, graphSummaryBanner} {
		if i := strings.Index(section, stop); i > 0 {
			section = section[:i]
			break
		}
	}
	if strings.Contains(section, leafDisp) {
		t.Fatalf("clean leaf display %q must not appear in peel order section:\n%s", leafDisp, section)
	}
	assertShowGraphZeroMutations(t, req)
}
```
