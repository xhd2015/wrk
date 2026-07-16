# Scenario

**Feature**: PlanLocalReinstalls errors when moduleRoot is not a Go module

```
# missing or unparseable go.mod → error, no plan items to assert
moduleRoot (no go.mod)
  -> PlanLocalReinstalls
  -> error
```

## Preconditions

- Leaves under this branch set `WantError = true`.
- ModuleRoot may exist as a directory but must not contain a valid `go.mod`.

## Steps

1. Leaves arrange an invalid/missing go.mod situation.
2. Assert Run returns a non-nil error.

## Context

- Product must not invent a module path when go.mod is absent.

```go
func Setup(t *testing.T, req *Request) error {
	req.WantError = true
	req.WantModuleName = ""
	req.WantItems = []WantPlanItem{}
	return nil
}
```
