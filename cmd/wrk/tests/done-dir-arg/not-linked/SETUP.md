# Scenario

**Feature**: wrk --done <mainRepo> should fail — not a linked worktree

```
# wrk --done on the main repo checkout (not a linked worktree)
wrk --done <mainRepo> -> non-zero, "not a linked worktree"
```

## Steps

1. Create main repo with go.mod (to pass go.mod check).
2. Run `wrk --done <mainRepo>` from WorkRoot.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo, _, _ := setupWrkWorktreeFromMain(t, req)
	// Undo the worktree: we only want the main repo for this test.
	// But setupWrkWorktreeFromMain already created one — that's fine,
	// we pass the main repo dir, not the worktree dir.
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--done", mainRepo}
	return nil
}
```