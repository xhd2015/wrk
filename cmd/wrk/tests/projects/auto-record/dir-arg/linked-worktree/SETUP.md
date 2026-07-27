# Scenario

**Feature**: auto-record via `wrk <linkedWt>` resolves to main repo

```
WorkRoot -> wrk linked-wt --list -> projects.json records myrepo main (not worktree path)
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Add linked worktree at `{WorkRoot}/linked-wt`.
3. Run `wrk <linkedWt> --list` from `{WorkRoot}`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	linkedWT := setupLinkedWorktree(t, mainRepo, "linked-wt", "linked-side")
	req.MainRepo = mainRepo
	req.TargetDir = linkedWT
	return nil
}
```