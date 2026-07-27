# Scenario

**Feature**: successful plans from a parseable go.mod module root

```
# go.mod present → Plan with ModuleName + Items (possibly empty / skips)
moduleRoot (go.mod) + binDir
  -> PlanLocalReinstalls
  -> LocalReinstallPlan ok
```

## Preconditions

- Every leaf under this branch writes a valid `go.mod` (unless a deeper leaf
  documents otherwise).
- `WantError` stays false.

## Steps

1. Leaves write go.mod and discovery fixtures.
2. Assert plan structure (ModuleName, BinDir, ordered Items).

## Context

- Group default: success path (no error).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.WantError = false
	return nil
}
```
