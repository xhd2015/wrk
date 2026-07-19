# Scenario

**Feature**: Run click while loading is ignored (miss)

```
# loading gate
Loading=true
click gen-commit-msg Run (would hit if idle)
  -> miss
```

## Steps

1. Aim gen-commit-msg Run under loading=true from parent.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.StageID = "gen-commit-msg"
	req.Target = "run"
	req.Loading = true
	return nil
}
```
