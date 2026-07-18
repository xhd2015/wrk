# Scenario

**Feature**: dry-run with **zero** cascade targets omits both phase headers; plan body still printed

```
# linked wt only (no nested linked WTs under top)
myrepo (v0.0.1) + wt (feature-work ahead)
  -> wrk --done --dry-run
  -> (no ==> cascade, no ==> own)
  -> primary MergeBack dry-run plan (merge --ff-only / worktree remove / branch -D)
  -> zero mutations
```

## Steps

1. Root-bump seed + wrk-managed linked worktree ahead (`feature-work`).
2. Snapshot baseline SHAs.
3. Run `wrk --done --dry-run` from the worktree (no `-y`).

```go
func Setup(t *testing.T, req *Request) error {
	setupDoneOutputLocalAhead(t, req)
	req.Args = []string{"--done", "--dry-run"}
	return nil
}
```
