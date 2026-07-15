# Scenario

**Feature**: plain `---- external ----` appears when main has a nested independent repo

```
# main + nested tools/child (no main-owned linked)
myrepo + tools/child -> wrk --status
  -> primary main
  -> ---- external ----
  -> tools/child
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on `main`.
2. Commit `.gitignore` with `tools/` so parent porcelain stays clean.
3. Initialize nested independent `{WorkRoot}/myrepo/tools/child`.
4. Run `wrk --status` from main root.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, mainRepo, "root status repo")
	ensureToolsGitignore(t, mainRepo, "tools/")
	child := initNestedIndependentRepo(t, mainRepo, "tools/child", "child status repo")

	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.DepPath = child
	return nil
}
```
