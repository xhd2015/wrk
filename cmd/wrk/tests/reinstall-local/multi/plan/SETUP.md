# Scenario

**Feature**: multi-module plans succeed (no install×install collision)

```
# valid moduleRoots + binDir → MultiLocalReinstallPlan ok
moduleRoots (each with go.mod) + binDir
  -> PlanLocalReinstallsMulti
  -> MultiLocalReinstallPlan{Modules sorted, Items per module}
```

## Preconditions

- Leaves under this branch do not set install×install collisions for the same
  BinName across modules.
- `WantError` stays false.

## Steps

1. Leaves write one or more module fixtures and set WantModules.
2. Assert multi plan structure (BinDir, ordered Modules, Items).

## Context

- Group default: success path (no error).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.WantError = false
	req.WantErrSubstrs = []string{}
	return nil
}
```
