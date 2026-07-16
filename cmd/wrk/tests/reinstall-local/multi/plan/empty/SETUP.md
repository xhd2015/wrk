# Scenario

**Feature**: empty moduleRoots list yields empty multi plan

```
# moduleRoots=[] → no modules, no items, err=nil
[] + binDir
  -> PlanLocalReinstallsMulti
  -> MultiLocalReinstallPlan{Modules:[]}
```

## Preconditions

- Leaves pass an empty ModuleRoots slice (not nil necessarily; Run accepts either).

## Steps

1. Leave ModuleRoots empty (root Setup defaults to empty slice).
2. Expect empty Modules and nil error.

## Context

- Empty input is valid: no discovery work, no collision check.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: empty moduleRoots branch. Leaves lock WantModules=[].
	if req.ModuleRoots == nil {
		req.ModuleRoots = []string{}
	}
	return nil
}
```
