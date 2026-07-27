# Scenario

**Feature**: same BinName across modules is ok when not install×install

```
# skip-only duplicates (same BinName, Action=skip on both) → no hard error
moduleA Action=skip + moduleB Action=skip (same BinName, bin absent)
  -> PlanLocalReinstallsMulti
  -> multi plan ok (both modules listed)
```

## Preconditions

- Leaves arrange two modules that discover the same BinName with Action=skip
  (shared binDir entry absent).

## Steps

1. Leaves write two modules that share a candidate bin name as skip.
2. Assert nil error and both modules present.

## Context

- Only install×install is a hard cross-module collision.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping: same-bin mixed-action success (no collision).
	if req.ModuleRoots == nil {
		req.ModuleRoots = []string{}
	}
	return nil
}
```
