# Scenario

**Feature**: dashboard layout builds a hitmap with stage Run and focus regions

```
# hitmap build
BuildDashboardHitmap(width, addDisabled)
  -> Hit[] + viewLines
  # gen-commit-msg / tag-next Run: distinct y0, runStage ids
  # stage row: left focus | right runStage
  # add-changes disabled: no runStage region
```

## Preconditions

- Op is `hitmap` — root Run returns hits without calling resolve.

## Steps

1. Set `req.Op = "hitmap"`.
2. Leaves set `AddDisabled` or only inspect default layout.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "hitmap"
	req.StageID = ""
	req.Target = ""
	return nil
}
```
