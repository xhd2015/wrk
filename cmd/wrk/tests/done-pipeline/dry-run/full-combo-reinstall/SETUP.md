# Scenario

**Feature**: full combo dry-run with reinstall tail plans every stage in order; zero mutations

```
# wtA ahead; wtB behind would-be main; origin; root-bump; GOBIN present stub
myrepo (origin, v0.0.1) + wtA + wtB + gobin/present
  -> wrk --done --sync --tag-next --push --reinstall-local --dry-run
  -> primary MergeBack DryRun plan
  -> blank → would: sync …
  -> blank → tag-next plan v0.0.2
  -> blank → would: git push origin main / v0.0.2
  -> blank → would: go install ./cmd/present + reinstall summary
  -> exit 0; wt still linked; no new tags; origin unchanged; stub unchanged
```

## Steps

1. Root-bump + bare origin + two worktrees; commit ahead on wtA.
2. Seed reinstall present on main + GOBIN.
3. Snapshot baseline SHAs.
4. Run full combo with `--reinstall-local` and `--dry-run` (no `-y`).

```go
func Setup(t *testing.T, req *Request) error {
	setupDonePipelineSyncWithOrigin(t, req)
	_ = seedDonePipelineReinstallPresent(t, req)
	recordComposeDryRunBaseline(t, req)
	req.Args = []string{"--done", "--sync", "--tag-next", "--push", "--reinstall-local", "--dry-run"}
	return nil
}
```
