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
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StageID = ""
	req.Target = "blank-below"
	return nil
}
```
