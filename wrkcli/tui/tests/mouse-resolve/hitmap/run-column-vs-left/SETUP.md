# Scenario

**Feature**: stage row splits left focus region and right Run chip

```
# gen-commit-msg row
left hit: empty runStage, focus >= 0, smaller x
right hit: runStage="gen-commit-msg", focus unused
same y0
```

## Steps

1. Build default hitmap; assert left vs Run for gen-commit-msg.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AddDisabled = false
	return nil
}
```
