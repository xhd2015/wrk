# Scenario

**Feature**: two wrk --task calls with same task produce incremental -N suffix

```
wrk --task "same task" (1st) -> {basename}-{token}-{date}-same-task
wrk --task "same task" (2nd) -> {basename}-{token}-{date}-same-task-1
```

## Steps

1. Create repo.
2. Setup creates the first worktree via runWrkWithArgs.
3. Run creates the second worktree (gets suffix -1).
4. Verify first has no suffix; second has suffix 1.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	taskDesc := "same task"
	req.TaskDesc = taskDesc
	req.RepoDir = mainRepo
	slug := slugify(taskDesc)

	// First worktree (Setup runs before Run)
	req.WtDir = runWrkWithArgs(t, req, mainRepo, "--task", taskDesc)
	req.WtBranch = branchNameWithTask("main", wrkDate, slug, 0)
	return nil
}
```