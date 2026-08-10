# Scenario

**Feature**: cross-repo require edges appear in both repo and module graphs

```
# root requires agent-pro; agent-pro requires dot-pkgs (all dirty)
wrk --unwind --show-graph
  -> repo edges: root → agent-pro → dot-pkgs (From depends on To)
  -> module edges [require …] among stack modules
  -> exit 0; no pin flags required
```

## Steps

1. Build three-repo dirty chain (require edges on go.mod).
2. Run show-graph only.

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
