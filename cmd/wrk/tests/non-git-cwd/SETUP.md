# Scenario

**Feature**: wrk rejects non-git cwd

```
# plain directory without .git
plain cwd -> wrk -> error (not a git repository)
```

## Steps

1. Create a non-git directory under `{WorkRoot}/plain`.
2. Run `wrk` with cwd set to that directory.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	plainDir := filepath.Join(req.WorkRoot, "plain")
	mkdirAll(t, plainDir)
	req.RepoDir = plainDir
	return nil
}
```