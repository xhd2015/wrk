# Scenario

**Feature**: clean leaf still listed in repo nodes; not in peel order

```
# same 3-repo chain; only agent-pro + root dirty; dot-pkgs clean
dot-pkgs (clean) · agent-pro (dirty) · root (dirty)
  -> wrk --unwind --show-graph
  -> repo nodes include dot-pkgs (clean)
  -> peel: external/agent-pro-… then . only
  -> no peel line for clean leaf
```

## Steps

1. Build three-repo chain.
2. Dirtify mid + root only; leave leaf clean.
3. Run show-graph; PeelOrder = mid external + primary `.`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupThreeRepoChain(t, req)
	dirtyMidAndRoot(t, req)
	setPeelOrderDisplays(t, req, req.ExternalWtDir, req.WtDir)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
