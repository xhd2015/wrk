# Scenario

**Feature**: wrk --list -v from main repo produces empty stderr

```
main repo cwd -> wrk --list -v -> stdout = git worktree list; stderr empty
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "list-v-main")
	initFetchVerboseRepo(t, repo, "list v main")
	req.RepoDir = repo
	req.Args = []string{"--list", "-v"}
	return nil
}
```