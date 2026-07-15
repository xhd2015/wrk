# Scenario

**Feature**: main-owned linked prints before header even when nested path sorts earlier under scan

```
# nested aaa/ would appear before out-of-tree linked under pure scan-then-append
# (scan finds aaa early; append puts WRK wt last). Primary/external flips that:
myrepo + wrk external + myrepo/aaa (nested main)
  -> primary: main, wrk-wt
  -> ---- external ----
  -> aaa
```

## Steps

1. Create main + one WRK external worktree.
2. Commit `.gitignore` with `aaa/` (parent stays clean).
3. Initialize nested independent `myrepo/aaa` (lexically early path).
4. Run `wrk --status` from main root.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo, _, _ := setupWrkWorktreeFromMain(t, req)
	// Nested independent repo under a path that sorts before typical scan relatives.
	ensureToolsGitignore(t, mainRepo, "aaa/")
	nested := initNestedIndependentRepo(t, mainRepo, "aaa", "nested early path")

	req.RepoDir = mainRepo
	req.DepPath = nested
	return nil
}
```
