# Scenario

**Feature**: multi-repo nested-external stack for human show-graph

```
# 3-repo chain under consumer linked wt
root → agent-pro → dot-pkgs
  -> free-first peel; clean nodes listed; require edges; apply hint
```

## Steps

1. Grouping scopes multi-repo human leaves (reuses setupThreeRepoChain).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	return nil
}
```
