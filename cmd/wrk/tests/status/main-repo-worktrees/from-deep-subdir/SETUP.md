# Scenario

**Feature**: --status from deep main subdirectory (>2 ups) falls back to absolute Dir

```
# nested cwd four levels under main
cwd = main/a/b/c/d
wrk --status -> main Dir: absolute (Rel would be ../../../..); Remote still present
```

## Steps

1. Create main + one external wrk worktree.
2. Create nested dirs `a/b/c/d` under main.
3. Run `wrk --status` with cwd = `main/a/b/c/d`.

## Context

- Leading `..` count after Clean is 4 (>2) → absolute `statusNormalizePath(main)`.
- External also absolute under the same rule.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	mainRepo, _, _ := createExternalWrkWorktree(t, req)
	subdir := filepath.Join(mainRepo, "a", "b", "c", "d")
	mkdirAll(t, subdir)
	req.RepoDir = subdir
	return nil
}
```
