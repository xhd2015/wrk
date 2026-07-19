# Scenario

**Feature**: disabled add-changes Run does not register a runStage hit

```
# add-changes with AddDisabled=true
BuildDashboardHitmap(AddDisabled=true)
  -> no Hit with runStage == "add-changes"
  # other stages still have Run hits
```

## Steps

1. Set `AddDisabled` true before hitmap build.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.AddDisabled = true
	return nil
}
```
