# Scenario

**Feature**: omit `---- external ----` when only main-owned linked worktrees exist

```
# main + WRK out-of-tree linked (primary membership via ListLinked)
myrepo + wrk external wt -> wrk --status
  -> primary main + wt
  -> no "---- external ----" substring
```

## Steps

1. Create main repo with `go.mod` and one `wrk` external worktree.
2. Run `wrk --status` from main root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, _, _ := setupWrkWorktreeFromMain(t, req)
	req.RepoDir = mainRepo
	return nil
}
```
