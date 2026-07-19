# Scenario

**Feature**: known origin maps tag-next Run

```
# known originY
absY = OriginY + tag-next localY
  -> runStage == "tag-next"
```

## Steps

1. Aim Run for `tag-next`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StageID = "tag-next"
	req.Target = "run"
	return nil
}
```
