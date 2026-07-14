# Scenario

**Feature**: auto-record from nested subdir cwd

```
myrepo/pkg/nested (cwd) -> wrk --list -> projects.json records myrepo main path
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Create nested subdir `{WorkRoot}/myrepo/pkg/nested`.
3. Run `wrk --list` with cwd = nested subpath.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	subpath := filepath.Join(mainRepo, "pkg", "nested")
	mkdirAll(t, subpath)
	req.MainRepo = mainRepo
	req.RepoDir = subpath
	return nil
}
```