# Scenario

**Feature**: `wrk src dst --bring …` creates at the spawn path then brings into it

```
# relocated from target-dir/with-other-mode/with-bring (no longer a reject)
src + dest missing, parent exists
  -> wrk src dst --no-config --bring d1
  -> worktree exactly at dst; external/ under dst; src untouched
```

## Steps

- `req.TargetDir = src`; `req.SpawnDir = dst`; `req.RepoDir = WorkRoot`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.RepoDir = req.WorkRoot
	return nil
}
```
