# Scenario

**Feature**: path present in both scan and ListLinked appears once in primary only

```
# in-tree linked is discovered by scan AND listed by ListLinked
PartitionStatusPaths(main, [main, inTree], [inTree])
  -> Primary=[main, inTree], External=[]
```

## Steps

1. Put an in-tree linked path in both `ScanPaths` and `LinkedOrdered`.
2. Expect a single primary occurrence after main; external empty.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	main := pathMain()
	inTree := pathLinkedInTree()
	req.MainRoot = main
	// Scan discovers main + in-tree linked (typical scan_repo result).
	req.ScanPaths = []string{main, inTree}
	req.LinkedOrdered = []string{inTree}
	req.WantPrimary = []string{main, inTree}
	req.WantExternal = []string{}
	return nil
}
```
