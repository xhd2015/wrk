# Scenario

**Feature**: `--done -y --tag-next --propagate-tags` creates next source tag then bumps consumer

```
# lib wt ahead (owned change); app requires example.com/lib@v0.0.1
linked wt + registered app
  -> wrk --done -y --tag-next --propagate-tags
  -> primary merge+remove
  -> blank → tag-next: create v0.0.2 at main HEAD
  -> blank → propagate: app v0.0.1 -> v0.0.2; build+commit chore(deps)
  -> event command "done"
```

## Steps

1. Multi-project fixture (`setupDonePipelinePropagateTagNext`).
2. Run `wrk --done -y --tag-next --propagate-tags` from linked worktree.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelinePropagateTagNext(t, req)
	req.Args = []string{"--done", "-y", "--tag-next", "--propagate-tags"}
	return nil
}
```
