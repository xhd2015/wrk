# Scenario

**Feature**: auto-record happens even when command fails later

```
myrepo + linked wt (dirty) -> wrk --done -> error but project already recorded
```

## Steps

1. Create main repo and linked worktree via `wrk`.
2. Write an uncommitted file in the worktree.
3. Run `wrk --done` from the dirty worktree.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	writeFile(t, filepath.Join(wtDir, "dirty-file"), "uncommitted")
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```