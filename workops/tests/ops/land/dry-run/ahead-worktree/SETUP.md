# Scenario

**Feature**: MergeBack DryRun on an ahead worktree does not change main or remove wt

```
# main + wt with commit ahead
Caller -> MergeBack({WorktreeDir: wt, DryRun: true})
  -> err nil; wt still on disk; main HEAD unchanged
```

## Steps

1. Seed main + linked worktree with an extra commit on the feature branch.
2. Snapshot main HEAD.
3. Checkout = worktree; DryRun already true from parent.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedMainWithAheadWorktree(t, req)
	req.Checkout = req.WtDir
	req.MainHEADBefore = revParseHEAD(t, req.MainRepo)
	return nil
}
```
