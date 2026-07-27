# Scenario

**Feature**: partition places non-primary scan paths into ordered external

```
# nested/dep scan hits not in main/ListLinked → External (path-sorted)
mainRoot + scan(main, nested…) + linkedOrdered
  -> PartitionStatusPaths
  -> Primary; External lexicographic by normalized path
```

## Preconditions

- Every leaf under this branch has at least one external path.
- External order is **not** scan input order; it is path-sorted.

## Steps

1. Default `MainRoot` to the synthetic main fixture path (leaves may refine).
2. Leaves include one or more nested/dep paths only in `scanPaths`.
3. Assert those paths appear only in `External`, sorted.

## Context

- Later phases print `---- external ----` when this list is non-empty (P2+).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Shared baseline for has-external leaves: main fixture root is known.
	req.MainRoot = pathMain()
	return nil
}
```
