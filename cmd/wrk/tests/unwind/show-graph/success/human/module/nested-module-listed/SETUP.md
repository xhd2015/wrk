# Scenario

**Feature**: nested go.mod under a stack checkout appears in module nodes

```
# root + intra-repo pkgs/shared module; primary dirty
wrk --unwind --show-graph
  -> modules list example.com/root and example.com/root/shared
  -> not a separate repo peel for shared
```

## Steps

1. Seed intra-repo local replace fixture (root + nested shared module).
2. Run show-graph.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowIntraRepoOnly(t, req)
	req.Args = showGraphArgs()
	recordUnwindBaseline(t, req)
	return nil
}
```
