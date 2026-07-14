# Scenario

**Feature**: wrk --add from linked worktree resolves to main repo

```
wrk --add linked-wt -> stdout myrepo main path (not worktree path)
```

## Steps

1. Initialize git repo at `{WorkRoot}/myrepo`.
2. Add linked worktree at `{WorkRoot}/linked-wt`.
3. Run `wrk --add <linkedWt>`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := initProjectsRepo(t, req.WorkRoot, "myrepo")
	linkedWT := setupLinkedWorktree(t, mainRepo, "linked-wt", "linked-side")
	req.MainRepo = mainRepo
	req.WtDir = linkedWT
	req.Args = []string{"--add", linkedWT}
	return nil
}
```