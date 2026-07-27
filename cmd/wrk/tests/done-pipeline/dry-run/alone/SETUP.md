# Scenario

**Feature**: `--done --dry-run` alone plans merge/remove only; zero mutations; no post stages

```
# primary dry-run only (no --sync/--tag-next/--push)
myrepo (v0.0.1) + wt (feature-work ahead)
  -> wrk --done --dry-run
  -> MergeBack DryRun planned commands (ff-merge + remove + branch -D)
  -> no would: sync / tag planned / would: git push
  -> wt still linked; main tip unchanged; feature-work only on wt
```

## Steps

1. Root-bump seed + wrk-managed linked worktree ahead (`feature-work`).
2. Snapshot baseline SHAs.
3. Run `wrk --done --dry-run` from the worktree (no `-y`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineLocal(t, req)
	recordComposeDryRunBaseline(t, req)
	req.Args = []string{"--done", "--dry-run"}
	return nil
}
```
