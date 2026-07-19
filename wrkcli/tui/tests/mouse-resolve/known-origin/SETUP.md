# Scenario

**Feature**: known originY — single mapping `localY = absY - originY`

```
# known origin resolve
OriginYSet=true, OriginY = O
absY = O + localY
  -> ResolveMouseHit uses only localY = absY - O
  -> Hit | miss
```

## Preconditions

- Leaves set `OriginYSet` and a non-zero `OriginY` (blank rows above UI).

## Steps

1. Enable known-origin mode with a fixed origin offset.
2. Leaves aim stage Run chips; root Run places absY at origin + localY.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "resolve"
	req.OriginYSet = true
	req.OriginY = 12
	// Keep Height consistent with known origin + viewLines.
	req.OriginOffset = 12
	req.ExtraBlank = 0
	return nil
}
```
