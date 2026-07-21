# Scenario

**Feature**: pre-existing branch with task-slug name causes -N suffix increment

```
# pre-create branch main-{date}-fix-login
# then wrk --task "fix login" -> dir and branch get -1 suffix
```

## Steps

1. Create repo.
2. Pre-create the branch.
3. Verify suffix 1.

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

	taskDesc := "fix login"
	slug := slugify(taskDesc)
	req.TaskDesc = taskDesc
	req.RepoDir = mainRepo

	// Pre-create the exact branch
	collisionBranch := branchNameWithTask("main", wrkDate, slug, 0)
	runGitIsolated(t, mainRepo, "branch", collisionBranch)
	req.WtBranch = branchNameWithTask("main", wrkDate, slug, 1)
	return nil
}
```