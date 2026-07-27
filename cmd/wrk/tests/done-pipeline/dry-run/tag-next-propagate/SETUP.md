# Scenario

**Feature**: dry-run plans primary + tag-next + would-propagate (planned next tag); zero mutations

```
# lib wt ahead; app requires v0.0.1; no -y
linked wt + registered app
  -> wrk --done --tag-next --propagate-tags --dry-run
  -> primary MergeBack DryRun plan
  -> blank → tag-next plan v0.0.2 (would-be main tip)
  -> blank → would: update app to planned v0.0.2
  -> exit 0; wt still linked; no v0.0.2 tag; app go.mod/HEAD unchanged
```

## Steps

1. Multi-project tag-next→propagate fixture.
2. Snapshot app baseline (fixture helper).
3. Run with `--dry-run` and **without** `-y`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelinePropagateTagNext(t, req)
	req.Args = []string{"--done", "--tag-next", "--propagate-tags", "--dry-run"}
	return nil
}
```
