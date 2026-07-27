# Scenario

**Feature**: flag order free — full combo with modifiers before `--done` still runs ordered pipeline

```
# flag argv order free; execution order remains sync → tag-next → push
myrepo (origin, v0.0.1) + wtA + wtB
  -> wrk --push --tag-next --sync --done -y
  -> same stdout/side effects as --done -y --sync --tag-next --push
```

## Steps

1. Same fixture as `sync-tag-next-push`.
2. Run with flags reordered: push/tag-next/sync before done.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineSyncWithOrigin(t, req)
	req.Args = []string{"--push", "--tag-next", "--sync", "--done", "-y"}
	return nil
}
```
