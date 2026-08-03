# Scenario

**Feature**: free-first peel order with relative display paths (leaf → mid → primary)

```
# chain: root depends on agent-pro depends on dot-pkgs; all dirty
# pin+land flags present so validation passes
# display = checkout relpath vs cwd (external/… then .)
dot-pkgs ← agent-pro ← root (all dirty)
  -> wrk --unwind --dry-run --tag-next --push --done
  -> would: peel external/dot-pkgs-main-2026-06-30
  -> would: peel external/agent-pro-main-2026-06-30
  -> would: peel .
  -> exit 0; zero mutations
```

## Steps

1. Build three-repo chain under consumer linked wt + nested externals.
2. Dirtify all three stack checkouts.
3. Snapshot HEADs; run unwind dry-run with pin + land flags.
4. PeelOrder = statusDirLine-style display paths (not bare MainRepo basenames).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyAllThree(t, req)
	// Free-first among dirty: leaf external, mid external, primary wt at cwd.
	setPeelOrderDisplays(t, req, req.DepsLinkedWtDir, req.ExternalWtDir, req.WtDir)
	// Flag order free; --done supplies land for linked nodes; pin via tag-next+push.
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```
