# Scenario

**Feature**: single-element moduleRoots list

```
# one module root → multi plan with one ModuleReinstallPlan
[moduleRoot] + binDir
  -> PlanLocalReinstallsMulti
  -> Modules length 1; items ≡ single-module planner
```

## Preconditions

- Leaves create exactly one go.mod module under WorkRoot.

## Steps

1. Leaves write one module + bins.
2. Assert one module block and item equivalence where specified.

## Context

- Caller order is irrelevant with one root; ModuleRoot must be absolute.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: single-module multi API branch.
	if req.ModuleRoots == nil {
		req.ModuleRoots = []string{}
	}
	return nil
}
```
