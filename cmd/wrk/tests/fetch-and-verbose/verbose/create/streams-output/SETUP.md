# Scenario

**Feature**: wrk -v create streams git worktree add subprocess output to stderr

```
main repo -> wrk -v -> stderr has timestamp worktree add log + git's own lines
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "create-v-stream")
	initFetchVerboseRepo(t, repo, "create v stream")
	req.RepoDir = repo
	req.Args = []string{"-v"}
	return nil
}
```