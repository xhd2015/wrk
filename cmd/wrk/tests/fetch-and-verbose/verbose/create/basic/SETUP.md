# Scenario

**Feature**: wrk -v create logs worktree add

```
main repo -> wrk -v -> stderr contains worktree add
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "create-v-main")
	initFetchVerboseRepo(t, repo, "create v main")
	req.RepoDir = repo
	req.Args = []string{"-v"}
	return nil
}
```