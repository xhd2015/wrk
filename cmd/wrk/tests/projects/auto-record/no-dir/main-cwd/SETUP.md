# Scenario

**Feature**: auto-record from main repo cwd

```
myrepo (main cwd) -> wrk --list -> projects.json records myrepo main path
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Run `wrk --list` with cwd = main repo.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	return nil
}
```