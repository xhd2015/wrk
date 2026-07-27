# Scenario

**Feature**: flag order free — full combo with modifiers before `--merge-back` still runs ordered pipeline; wt kept

```
# flag argv order free; execution order remains sync → tag-next → push
myrepo (origin, v0.0.1) + wtA + wtB
  -> wrk --push --tag-next --sync --merge-back -y
  -> same stdout/side effects as --merge-back -y --sync --tag-next --push
  -> source wt remains
```

## Steps

1. Same fixture as `sync-tag-next-push`.
2. Run with flags reordered: push/tag-next/sync before merge-back.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupMergeBackPipelineSyncWithOrigin(t, req)
	req.Args = []string{"--push", "--tag-next", "--sync", "--merge-back", "-y"}
	return nil
}
```
