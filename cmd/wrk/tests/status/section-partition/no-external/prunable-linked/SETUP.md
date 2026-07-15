# Scenario

**Feature**: prunable/dead path only in ListLinked remains primary in porcelain order

```
# dead path not in scanPaths; still primary via linked membership
PartitionStatusPaths(main, [main], [dead, wtEarly])
  -> Primary=[main, dead, wtEarly], External=[]
```

## Steps

1. Put a dead/prunable checkout path only in `LinkedOrdered` (not in scan).
2. Follow with a live out-of-tree linked path.
3. Expect both in primary after main, preserving linked order.

```go
func Setup(t *testing.T, req *Request) error {
	main := pathMain()
	dead := pathDeadLinked()
	early := pathLinkedEarly()
	req.MainRoot = main
	req.ScanPaths = []string{main}
	// Dead entry appears only in ListLinked porcelain list.
	req.LinkedOrdered = []string{dead, early}
	req.WantPrimary = []string{main, dead, early}
	req.WantExternal = []string{}
	return nil
}
```
