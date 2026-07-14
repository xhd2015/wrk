# Scenario

**Feature**: wrk --done rejects main repo checkout (not a linked worktree)

```
# cwd is main repo checkout, not a linked worktree
myrepo (main) -> wrk --done -> not a linked worktree
```

## Steps

1. Create main repo (no linked worktree as cwd).
2. Run `wrk --done` from main repo root.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	// wrk --done reaches the linked-worktree check only after the go.mod check.
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\n\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	req.RepoDir = mainRepo
	req.Args = []string{"--done"}
	return nil
}
```