# Scenario

**Feature**: bottom-anchored geometry — UI sits at bottom of terminal

```
# bottom-anchored paint (inline Bubble Tea typical)
origin = OriginOffset  (= height - viewLines)
UI lines at absY origin .. origin+viewLines-1

# click on stage Run
absY = origin + localY
  -> dual-origin resolve
  -> correct runStage
```

## Preconditions

- `OriginOffset > 0` so bottom candidate maps localY correctly; top candidate
  alone would miss or hit the wrong row.

## Steps

1. Set `OriginOffset` so UI is painted above the bottom of a taller terminal.
2. Leaves aim StageID Run; root Run uses absY = OriginOffset + localY.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.OriginOffset = 25
	req.ExtraBlank = 0
	req.OriginYSet = false
	return nil
}
```
