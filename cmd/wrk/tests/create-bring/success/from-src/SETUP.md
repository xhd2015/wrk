# Scenario

**Feature**: `wrk src --bring …` creates a default worktree then brings into it

```
# first positional is create source; --bring applies inside the new WT
WorkRoot -> wrk src --bring d1 [d2] --no-config
  -> {WRK_HOME}/worktrees/src-main-{date}
  -> external/ under that WT; src/external does not exist
```

## Steps

- `req.TargetDir = src`; `req.RepoDir = WorkRoot`.
- Leaves append `--no-config --bring …`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RepoDir = req.WorkRoot
	return nil
}
```
