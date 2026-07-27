# Scenario

**Feature**: main repo prints full primary block for one clean out-of-tree wrk worktree

```
# wrk creates out-of-tree linked wt under WRK_HOME (primary membership)
myrepo -> wrk -> out-of-tree wt

# --status from main: primary main + wt (Dir via statusDirLine, Master); no section header
wrk --status from main cwd -> primary only
```

## Steps

1. Create `{WorkRoot}/myrepo` with `go.mod` and run `wrk` once.
2. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, _, _ := createExternalWrkWorktree(t, req)
	req.RepoDir = mainRepo
	return nil
}
```
