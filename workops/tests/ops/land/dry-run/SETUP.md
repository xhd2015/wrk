# Scenario

**Feature**: MergeBack DryRun plans land without mutations (Sync default off)

```
# DryRun=true; Sync default false (with-sync leaf overrides Sync=true)
Caller -> MergeBack(..., DryRun) -> err nil; no git mutations
```

## Steps

1. Force DryRun true for this subtree; Sync defaults false.
2. Leaves provide worktree fixtures, optional Sync=true override, snapshot main HEAD.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.DryRun = true
	req.Sync = false
	return nil
}
```
