# Scenario

**Feature**: top-anchored geometry — UI near terminal top, blank below

```
# top-anchored paint (screenshot layout)
UI lines at absY 0..viewLines-1
blank rows at absY viewLines..height-1
height = viewLines + ExtraBlank  (ExtraBlank large)

# click on a stage Run chip
absY = localY of that chip  (origin 0)
  -> dual-origin resolve
  -> correct runStage
```

## Preconditions

- `height >> viewLines` so a naive bottom-only map (`localY = absY - (height-viewLines)`)
  produces a **wrong** local Y for top-of-screen clicks (the reported bug class).

## Steps

1. Set large `ExtraBlank` so height greatly exceeds viewLines.
2. Leaves aim StageID Run with top-anchored absY (= localY).

## Context

- This branch seals the gen-commit-msg vs tag-next regression first.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Plenty of blank below the UI so bottom-only mapping is wrong.
	req.ExtraBlank = 40
	req.OriginOffset = 0
	req.OriginYSet = false
	return nil
}
```
