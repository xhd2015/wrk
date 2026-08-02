# Scenario

**Feature**: Status on a linked worktree reports IsWorktree and MainPath

```
# main + linked wt on feature-work
Caller -> Status(wtDir)
  -> IsWorktree true, MainPath=main, Branch non-empty, CheckoutPath=wt
```

## Steps

1. Seed main + linked worktree.
2. Set Checkout to worktree path.
3. Run Status.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedMainWithLinkedWorktree(t, req)
	req.Checkout = req.WtDir
	return nil
}
```
