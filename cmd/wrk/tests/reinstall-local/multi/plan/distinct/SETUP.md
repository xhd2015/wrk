# Scenario

**Feature**: two modules contribute distinct bins to one multi plan

```
# root module + nested tools module; different BinNames; both install
[root, tools] + shared binDir
  -> PlanLocalReinstallsMulti
  -> two ModuleReinstallPlan blocks, sorted by ModuleRoot
```

## Preconditions

- Leaves use two go.mod roots with non-overlapping installable bin names.

## Steps

1. Leaves create root + nested (or sibling) modules and shared GOBIN stubs.
2. Assert both modules appear with their install items.

## Context

- Modules sorted lex by absolute ModuleRoot path regardless of ModuleRoots order.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping: multi-module distinct-bins success branch.
	if req.ModuleRoots == nil {
		req.ModuleRoots = []string{}
	}
	return nil
}
```
