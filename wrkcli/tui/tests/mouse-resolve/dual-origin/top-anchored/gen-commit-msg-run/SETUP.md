# Scenario

**Bug**: top-anchored click on gen-commit-msg Run must not fire tag-next

```
# bug regression (top-anchored, dual-origin)
BuildDashboardHitmap -> hitmap with gen-commit-msg Run localY
absY = that localY  (UI origin 0)
height = viewLines + 40
  -> ResolveMouseHit (origin unknown)
  -> runStage == "gen-commit-msg"
  # NOT "tag-next"
```

## Steps

1. Aim Run chip for stage `gen-commit-msg`.
2. Use dual-origin top-anchored geometry from ancestors.

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
