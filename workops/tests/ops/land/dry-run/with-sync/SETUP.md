# Scenario

**Feature**: MergeBack DryRun with Sync=true does not mutate main or remove wt

```
# main + wt with commit ahead; Sync composition requested under dry-run
Caller -> MergeBack({WorktreeDir: wt, DryRun: true, Sync: true})
  -> err nil; wt still on disk; main HEAD unchanged; no sync side effects
```

## Steps

1. Seed main + linked worktree with an extra commit on the feature branch.
2. Snapshot main HEAD.
3. Checkout = worktree; DryRun true from parent; override Sync=true.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedMainWithAheadWorktree(t, req)
	req.Checkout = req.WtDir
	req.Sync = true
	req.MainHEADBefore = revParseHEAD(t, req.MainRepo)
	return nil
}
```
