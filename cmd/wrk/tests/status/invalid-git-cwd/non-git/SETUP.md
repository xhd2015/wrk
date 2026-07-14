# Scenario

**Feature**: wrk --status reports a clear error for non-git cwd

```
# no .git ancestor exists
plain cwd -> wrk --status -> non-zero stderr
```

## Steps

1. Create `{WorkRoot}/plain` without a `.git` directory.
2. Run `wrk --status` from that directory.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	plain := filepath.Join(req.WorkRoot, "plain")
	mkdirAll(t, plain)

	req.RepoDir = plain
	return nil
}
```
