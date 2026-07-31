# Scenario

**Feature**: clean leaf is omitted from pending; peel starts at dirty mid then root

```
# same 3-repo chain; only agent-pro + root dirty; dot-pkgs clean
dot-pkgs (clean) · agent-pro (dirty) · root (dirty)
  -> wrk --unwind --dry-run --tag-next --push --done
  -> would: peel agent-pro
  -> would: peel root
  -> no would: peel dot-pkgs
```

## Steps

1. Build three-repo chain.
2. Dirtify mid + root only; leave leaf clean.
3. Run unwind dry-run with pin + land flags.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyMidAndRoot(t, req)
	req.PeelOrder = []string{labelAgentPro, labelRoot}
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push", "--done"}
	recordUnwindBaseline(t, req)
	return nil
}
```
