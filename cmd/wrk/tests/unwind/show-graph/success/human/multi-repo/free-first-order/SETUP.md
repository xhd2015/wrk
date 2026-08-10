# Scenario

**Feature**: free-first peel order with relative display paths on show-graph

```
# chain: root depends on agent-pro depends on dot-pkgs; all dirty
# no pin/land flags required for show-graph
dot-pkgs ← agent-pro ← root (all dirty)
  -> wrk --unwind --show-graph
  -> peel: external/dot-pkgs-main-… then external/agent-pro-main-… then .
  -> exit 0; zero mutations
```

## Steps

1. Build three-repo chain under consumer linked wt + nested externals.
2. Dirtify all three stack checkouts.
3. Snapshot HEADs; run show-graph only.
4. PeelOrder = statusDirLine-style display paths.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyAllThree(t, req)
	setPeelOrderDisplays(t, req, req.DepsLinkedWtDir, req.ExternalWtDir, req.WtDir)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
