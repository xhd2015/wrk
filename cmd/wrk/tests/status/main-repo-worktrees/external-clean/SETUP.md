# Scenario

**Feature**: main repo appends full block for one clean external wrk worktree

```
# wrk creates external linked wt under WRK_HOME
myrepo -> wrk -> external wt

# --status from main: scan `.` + appended full block (Dir via statusDirLine, Master)
wrk --status from main cwd -> scan + append external
```

## Steps

1. Create `{WorkRoot}/myrepo` with `go.mod` and run `wrk` once.
2. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, _, _ := createExternalWrkWorktree(t, req)
	req.RepoDir = mainRepo
	return nil
}
```
