# Scenario

**Feature**: WhereMain on a linked worktree returns the main abs path

```
# main + linked wt
Caller -> WhereMain(wtDir) -> mainAbs ≠ wtDir; mainAbs == MainRepo
```

## Steps

1. Seed main repo + linked worktree on `feature-work`.
2. Set Checkout to the worktree path.
3. Run WhereMain.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedMainWithLinkedWorktree(t, req)
	req.Checkout = req.WtDir
	return nil
}
```
