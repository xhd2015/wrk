# Scenario

**Feature**: `--done -y --sync --tag-next` runs sync then local tag-next (no push)

```
# wtA ahead; wtB behind; after done: sync pass2 then local v0.0.2
myrepo (v0.0.1) + wtA + wtB (feature-stays)
  -> wrk --done -y --sync --tag-next
  -> merged branch <wtA> into main
  -> <blank>
  -> feature-stays ← main (+1); synced summary
  -> <blank>
  -> tag-next apply v0.0.2 local
  -> no push line; wtA gone; wtB HEAD == main
```

## Steps

1. Root-bump seed; wrk wtA + manual wtB `feature-stays`; commit ahead on wtA.
2. Run `wrk --done -y --sync --tag-next` from wtA.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupDonePipelineSync(t, req)
	req.Args = []string{"--done", "-y", "--sync", "--tag-next"}
	return nil
}
```
