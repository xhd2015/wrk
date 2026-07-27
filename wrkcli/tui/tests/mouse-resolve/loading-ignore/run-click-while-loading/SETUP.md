# Scenario

**Feature**: Run click while loading is ignored (miss)

```
# loading gate
Loading=true
click gen-commit-msg Run (would hit if idle)
  -> miss
```

## Steps

1. Aim gen-commit-msg Run under loading=true from parent.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.StageID = "gen-commit-msg"
	req.Target = "run"
	req.Loading = true
	return nil
}
```
