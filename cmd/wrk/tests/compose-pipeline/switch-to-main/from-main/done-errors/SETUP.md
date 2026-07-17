# Scenario

**Feature**: bare `--done` on main hard-errors (linked worktree required)

```
myrepo (main, go.mod) -> wrk --done
  -> non-zero
  -> stderr names --done and linked worktree requirement
```

## Steps

1. Main repo with go.mod (past early checks).
2. Run `wrk --done` from main.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\n\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")
	req.RepoDir = mainRepo
	req.Args = []string{"--done"}
	return nil
}
```
