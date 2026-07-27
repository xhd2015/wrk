# Scenario

**Feature**: click in blank area below top-anchored UI is a miss

```
# top-anchored: blank under UI
absY = viewLines  (first row below last UI line)
height = viewLines + ExtraBlank
  -> ResolveMouseHit
  -> miss (OK == false)
```

## Steps

1. Set Target `blank-below` so Run aims absY at/under the bottom of the UI.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.StageID = ""
	req.Target = "blank-below"
	return nil
}
```
