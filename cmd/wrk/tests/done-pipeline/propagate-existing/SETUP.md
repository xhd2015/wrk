# Scenario

**Feature**: primary without `--tag-next` still runs propagate on **existing** source tags

```
# lib already tagged v0.0.2; app still requires v0.0.1; wt ahead (feature only)
linked wt + registered app
  -> wrk --done -y --propagate-tags
  -> primary merge+remove
  -> blank → propagate using existing v0.0.2 (no tag-next stage)
  -> app require v0.0.1 -> v0.0.2; build+commit
  -> event command "done"
```

## Steps

1. Fixture with pre-existing `v0.0.2` on source and outdated consumer (`setupDonePipelinePropagateExisting`).
2. Run `wrk --done -y --propagate-tags` (no `--tag-next`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelinePropagateExisting(t, req)
	req.Args = []string{"--done", "-y", "--propagate-tags"}
	return nil
}
```
