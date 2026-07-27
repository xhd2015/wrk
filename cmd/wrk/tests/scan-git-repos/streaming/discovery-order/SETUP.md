# Scenario

**Feature**: multi-root scan always prints valid mains in CLI root order (discovery), not sorted paths

```
root-b/main-b + root-a/main-a
  -> wrk --scan-git-repos root-b root-a
  -> discovery order: main-b then main-a (always-print as found)
  -> NOT lexicographic: main-a then main-b
  -> projects.json empty; both paths on stdout
```

## Steps

1. Create two sibling scan roots under `{WorkRoot}`: `root-b` and `root-a`.
2. Init one main repo under each: `root-b/main-b`, `root-a/main-a`.
3. Pass roots on the CLI as `root-b` then `root-a` so discovery order is main-b before main-a.
4. Choose names so absolute path of main-a sorts before main-b (lex), ensuring batch sort would invert discovery order.
5. Run `wrk --scan-git-repos <root-b> <root-a>` from non-git WorkRoot.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	// root-b before root-a on CLI → discovery order main-b then main-a.
	// Lex sort of abs paths is main-a then main-b (root-a < root-b).
	rootB := filepath.Join(req.WorkRoot, "root-b")
	rootA := filepath.Join(req.WorkRoot, "root-a")
	mkdirAll(t, rootB)
	mkdirAll(t, rootA)

	mainB := initScanMainRepo(t, rootB, "main-b")
	mainA := initScanMainRepo(t, rootA, "main-a")

	pathB := resolveScanPath(t, mainB)
	pathA := resolveScanPath(t, mainA)
	if !(pathA < pathB) {
		t.Fatalf("fixture invariant: pathA must sort before pathB so discovery≠lex; pathA=%q pathB=%q", pathA, pathB)
	}

	// MainRepo = first discovered; SecondRepo = second discovered.
	req.MainRepo = mainB
	req.SecondRepo = mainA
	req.Args = []string{"--scan-git-repos", rootB, rootA}
	return nil
}
```
