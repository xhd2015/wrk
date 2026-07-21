# Scenario

**Feature**: wrk -t "fix login bug" produces a worktree with slug in dir and branch names

```
wrk -> [-t "fix login bug"] -> git worktree add -> path on stdout
```

## Preconditions

- A fresh git repo on `master` branch exists.

## Steps

1. Create repo on `master`.
2. Run `wrk -t "fix login bug"` (handled by Run via req.TaskDesc + req.TaskFlag).
3. Verify stdout is the worktree path.
4. Verify dir and branch names include `-fix-login-bug` after the date.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	runGitIsolated(t, mainRepo, "branch", "-m", "master")
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")

	req.TaskDesc = "fix login bug"
	req.TaskFlag = "-t"
	req.RepoDir = mainRepo

	slug := slugify("fix login bug")
	req.WtBranch = branchNameWithTask("master", wrkDate, slug, 0)
	return nil
}
```