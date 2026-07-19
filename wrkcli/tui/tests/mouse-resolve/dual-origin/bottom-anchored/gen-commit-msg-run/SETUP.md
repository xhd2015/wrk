# Scenario

**Feature**: bottom-anchored gen-commit-msg Run resolves correctly

```
# bottom-anchored dual-origin
absY = origin + gen-commit-msg localY
  -> runStage == "gen-commit-msg"
```

## Steps

1. Aim Run chip for `gen-commit-msg`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StageID = "gen-commit-msg"
	req.Target = "run"
	return nil
}
```
