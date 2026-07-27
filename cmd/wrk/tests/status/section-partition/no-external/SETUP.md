# Scenario

**Feature**: partition yields empty external when only main ± linked-to-main paths

```
# no nested/dep scan hits → External == []
mainRoot + scan(main, linked…) + linkedOrdered
  -> PartitionStatusPaths
  -> Primary ordered; External empty
```

## Preconditions

- Every leaf under this branch expects `WantExternal` to be empty (or omit
  external-only paths from scan).
- Primary still carries main and any ListLinked members (including dead).

## Steps

1. Leaves compose scan/linked so every scan path is primary membership.
2. Assert external is empty and primary order matches product rules.

## Context

- Empty external is the P2 signal to omit `---- external ----` (not asserted here).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Grouping default: no external paths expected unless a leaf overrides.
	req.WantExternal = []string{}
	return nil
}
```
