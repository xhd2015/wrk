# Scenario

**Feature**: `wrk -y` create mode matches bare `wrk`

```
main repo -> wrk -y -> worktree path on stdout; same side effects as wrk
```

## Steps

1. Create main repo.
2. Run `wrk -y` from main repo.

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
	req.RepoDir = mainRepo
	req.Args = []string{"-y"}
	return nil
}
```
