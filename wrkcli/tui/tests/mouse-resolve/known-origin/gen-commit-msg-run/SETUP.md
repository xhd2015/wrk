# Scenario

**Feature**: known origin maps gen-commit-msg Run

```
# known originY
absY = OriginY + gen-commit-msg localY
  -> runStage == "gen-commit-msg"
```

## Steps

1. Aim Run for `gen-commit-msg`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StageID = "gen-commit-msg"
	req.Target = "run"
	return nil
}
```
