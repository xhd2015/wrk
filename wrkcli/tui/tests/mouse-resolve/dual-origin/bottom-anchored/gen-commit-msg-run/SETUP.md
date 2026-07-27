# Scenario

**Feature**: bottom-anchored gen-commit-msg Run resolves correctly

```
# bottom-anchored dual-origin
absY = origin + gen-commit-msg localY
  -> runStage == "gen-commit-msg"
```

## Steps

1. Aim Run chip for `gen-commit-msg`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.StageID = "gen-commit-msg"
	req.Target = "run"
	return nil
}
```
