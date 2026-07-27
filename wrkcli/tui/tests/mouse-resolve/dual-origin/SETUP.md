# Scenario

**Feature**: origin unknown — dual-origin top then bottom fallback

```
# dual-origin resolve (OriginY unset)
absY, height, viewLines, hitmap
  -> try top: localY = absY
  -> try bottom: localY = absY - max(0, height-viewLines)
  -> prefer runStage hit when candidates disagree
  -> Hit | miss
```

## Preconditions

- Leaves leave `OriginYSet` false so `ResolveMouseHit` uses dual-origin.

## Steps

1. Mark resolve as dual-origin (no known originY).
2. Child nodes set top-anchored vs bottom-anchored absolute geometry.

## Context

- Top-anchored: paint origin 0; `ExtraBlank` makes `height >> viewLines`.
- Bottom-anchored: paint origin `OriginOffset`; absY = origin + localY.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = "resolve"
	req.OriginYSet = false
	return nil
}
```
