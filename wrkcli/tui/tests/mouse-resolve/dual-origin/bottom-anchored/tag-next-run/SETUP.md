# Scenario

**Feature**: bottom-anchored tag-next Run resolves correctly

```
# bottom-anchored dual-origin
absY = origin + tag-next localY
  -> runStage == "tag-next"
```

## Steps

1. Aim Run chip for `tag-next`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StageID = "tag-next"
	req.Target = "run"
	return nil
}
```
