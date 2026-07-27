# Scenario

**Feature**: mixed main + in-tree linked + out-of-tree linked + nested scan paths

```
# primary = [main] + ListLinked porcelain (in-tree then out-of-tree)
# external = nested-only scan hits, path-sorted; no primary dups
PartitionStatusPaths(
  main,
  scan=[taskHub, main, inTree, tools, external],  # scrambled
  linked=[inTree, late, early],
)
  -> Primary=[main, inTree, late, early]
  -> External=sort([external, taskHub, tools])
```

## Steps

1. Build full mixed inputs: main, in-tree linked (also in scan), two out-of-tree
   linked (ListLinked order anti path-sort), three nested scan-only paths.
2. Scramble scan order.
3. Expect primary ListLinked order after main; external nesteds only, sorted;
   in-tree linked not in external; no duplicates.

```go
import (
	"github.com/xhd2015/doctest/session"
	"sort"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	main := pathMain()
	inTree := pathLinkedInTree()
	late := pathLinkedLate()
	early := pathLinkedEarly()
	ext := pathNestedExternal()
	hub := pathNestedTaskHub()
	tools := pathNestedTools()

	req.MainRoot = main
	// Scrambled scan order: nesteds + main + in-tree; out-of-tree linked absent.
	req.ScanPaths = []string{hub, main, inTree, tools, ext}
	// ListLinked porcelain: in-tree first, then out-of-tree anti path-sort.
	req.LinkedOrdered = []string{inTree, late, early}

	req.WantPrimary = []string{main, inTree, late, early}
	// external/child < task-hub < tools/child
	req.WantExternal = []string{ext, hub, tools}
	return nil
}
```
