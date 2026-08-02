# Scenario

**Feature**: MergeBack lands a linked worktree into main without removing it

```
# linked wt ahead of main
Caller -> workops.MergeBack(ctx, {WorktreeDir, DryRun, Sync})
  -> plan or land; worktree kept (Remove=false)
```

## Steps

1. Grouping only: set Op to merge-back.
2. Dry-run leaves set DryRun and seed ahead worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Op = OpMergeBack
	return nil
}
```
