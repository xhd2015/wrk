# Scenario

**Feature**: `--done --dry-run --reinstall-local` plans primary then reinstall dry plan; zero mutations

```
# linked wt ahead; main has ./cmd/present + GOBIN/present stub
myrepo (v0.0.1) + wt (feature-work) + gobin/present
  -> wrk --done --dry-run --reinstall-local
  -> primary MergeBack DryRun plan
  -> blank → reinstall would: go install ./cmd/present + summary
  -> exit 0; wt still linked; present stub unchanged
  -> no mutual exclusion
```

## Steps

1. Root-bump + wrk-managed linked worktree ahead.
2. Seed `cmd/present` on main + GOBIN stub (`seedDonePipelineReinstallPresent`).
3. Snapshot baseline SHAs.
4. Run `wrk --done --dry-run --reinstall-local` (no `-y`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineLocal(t, req)
	_ = seedDonePipelineReinstallPresent(t, req)
	recordComposeDryRunBaseline(t, req)
	req.Args = []string{"--done", "--dry-run", "--reinstall-local"}
	return nil
}
```
