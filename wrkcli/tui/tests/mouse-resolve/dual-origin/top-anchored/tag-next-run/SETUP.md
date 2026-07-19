# Scenario

**Feature**: top-anchored tag-next Run resolves to tag-next

```
# top-anchored dual-origin
click tag-next Run (absY = localY, height >> viewLines)
  -> runStage == "tag-next"
```

## Steps

1. Aim Run chip for stage `tag-next`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StageID = "tag-next"
	req.Target = "run"
	return nil
}
```
