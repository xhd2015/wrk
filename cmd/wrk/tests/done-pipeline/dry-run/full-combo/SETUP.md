# Scenario

**Feature**: full combo dry-run `--done --sync --tag-next --push --dry-run` plans every stage; zero mutations

```
# wtA ahead; wtB behind would-be main; origin present; root-bump seed
myrepo (origin, v0.0.1) + wtA + wtB
  -> wrk --done --sync --tag-next --push --dry-run
  -> primary MergeBack DryRun plan
  -> blank → would: sync (pass-2 distribute to feature-stays using would-be main tip)
  -> blank → tag-next plan v0.0.2 (tip = would-be main = wt HEAD for ahead/FF)
  -> blank → would: git push origin main / v0.0.2
  -> exit 0; wt still linked; no new tags; origin unchanged
```

## Steps

1. Root-bump + bare origin + two worktrees; commit ahead on wtA.
2. Snapshot baseline SHAs (main, wt, wt2, origin).
3. Run full combo with `--dry-run` and **without** `-y`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineSyncWithOrigin(t, req)
	recordComposeDryRunBaseline(t, req)
	// Flag order free; no -y — dry-run must not prompt.
	req.Args = []string{"--done", "--sync", "--tag-next", "--push", "--dry-run"}
	return nil
}
```
