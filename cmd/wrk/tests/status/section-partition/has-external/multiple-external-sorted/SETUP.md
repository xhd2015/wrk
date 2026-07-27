# Scenario

**Feature**: multiple external paths are sorted by normalized path, not scan order

```
# scan lists tools/child before external/child before task-hub (anti-sorted)
# external must be lexicographic: external/child, task-hub, tools/child
PartitionStatusPaths(main, [main, tools, external, taskHub], [])
  -> Primary=[main]
  -> External=sort([external, taskHub, tools])
```

## Steps

1. Put three nested/dep paths in `ScanPaths` in **non-sorted** order.
2. Linked empty.
3. Expect external sorted by normalized absolute path.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	main := pathMain()
	// Path sort order among these three:
	//   external/child  <  task-hub  <  tools/child
	// (under the same main prefix).
	ext := pathNestedExternal() // .../external/child
	hub := pathNestedTaskHub()  // .../task-hub
	tools := pathNestedTools()  // .../tools/child
	req.MainRoot = main
	// Deliberately anti-sorted scan input order.
	req.ScanPaths = []string{main, tools, hub, ext}
	req.LinkedOrdered = []string{}
	req.WantPrimary = []string{main}
	req.WantExternal = []string{ext, hub, tools}
	return nil
}
```
