# Scenario

**Feature**: main repo without upstream shows Remote: (no upstream)

```
plain main repo (no tracking remote) -> wrk --status -> root Remote: (no upstream)
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "no-upstream")
	initFetchVerboseRepo(t, repo, "no upstream base")
	req.MainRepo = repo
	req.RepoDir = repo
	req.Args = []string{"--status"}
	return nil
}
```