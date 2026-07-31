# Scenario

**Feature**: free-first peel order for three dirty stack repos (leaf → mid → root)

```
# chain: root depends on agent-pro depends on dot-pkgs; all dirty
# pin+land flags present so validation passes
dot-pkgs ← agent-pro ← root (all dirty)
  -> wrk --unwind --dry-run --tag-next --push --done
  -> would: peel dot-pkgs
  -> would: peel agent-pro
  -> would: peel root
  -> exit 0; zero mutations
```

## Steps

1. Build three-repo chain under consumer linked wt + nested externals.
2. Dirtify all three stack checkouts.
3. Snapshot HEADs; run unwind dry-run with pin + land flags.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyAllThree(t, req)
	req.PeelOrder = []string{labelDotPkgs, labelAgentPro, labelRoot}
	// Flag order free; --done supplies land for linked nodes; pin via tag-next+push.
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```
