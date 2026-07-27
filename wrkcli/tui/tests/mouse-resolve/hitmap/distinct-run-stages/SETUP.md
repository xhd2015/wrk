# Scenario

**Feature**: gen-commit-msg and tag-next Run hits are distinct

```
# hitmap layout seal
BuildDashboardHitmap
  -> hit runStage="gen-commit-msg" at y_g
  -> hit runStage="tag-next" at y_t
  # y_g != y_t
```

## Steps

1. Build default dashboard hitmap (enabled Runs).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.AddDisabled = false
	return nil
}
```
