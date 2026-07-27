# Scenario

**Feature**: wrk --rm from linked worktree resolves to main repo

```
wrk --add myrepo -> wrk --rm linked-wt -> stdout main path (not worktree path); entry gone
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Add linked worktree at `{WorkRoot}/linked-wt`.
3. Run `wrk --add <mainRepo>` to record.
4. Run `wrk --rm <linkedWt>`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	linkedWT := setupLinkedWorktree(t, mainRepo, "linked-wt", "linked-side")
	recordProjectViaAdd(t, req, mainRepo)
	req.MainRepo = mainRepo
	req.WtDir = linkedWT
	req.Args = []string{"--rm", linkedWT}
	return nil
}
```
