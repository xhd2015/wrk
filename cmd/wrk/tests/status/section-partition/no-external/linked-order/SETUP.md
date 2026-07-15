# Scenario

**Feature**: primary linked order follows ListLinked porcelain, not path sort

```
# linkedOrdered=[wtLate, wtEarly] where path(wtEarly) < path(wtLate)
# primary must be [main, wtLate, wtEarly] — preserve porcelain order
PartitionStatusPaths(main, [main], [wtLate, wtEarly])
  -> Primary=[main, wtLate, wtEarly], External=[]
```

## Steps

1. Use two out-of-tree linked paths whose **path sort** is Early then Late.
2. Pass them in ListLinked order **Late then Early**.
3. Expect primary to preserve Late-before-Early (not lexicographic).

```go
func Setup(t *testing.T, req *Request) error {
	main := pathMain()
	late := pathLinkedLate()   // .../zzz-late  (sorts after)
	early := pathLinkedEarly() // .../aaa-early (sorts before)
	req.MainRoot = main
	// Scan sees only main (out-of-tree linked not under scan root).
	req.ScanPaths = []string{main}
	// Porcelain / ListLinked order intentionally anti-sorted by path.
	req.LinkedOrdered = []string{late, early}
	req.WantPrimary = []string{main, late, early}
	req.WantExternal = []string{}
	return nil
}
```
